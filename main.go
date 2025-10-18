package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	OllamaHost  string
	OllamaModel string
}

// OllamaEmbeddingRequest represents the request to Ollama API
type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbeddingResponse represents the response from Ollama API
type OllamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func loadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://ollama:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "bge-m3:latest"
	}

	return &Config{
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		OllamaHost:  ollamaHost,
		OllamaModel: ollamaModel,
	}
}

func createTestVectorDB(cfg *Config) error {
	// Connect to default postgres database to create new database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'testvector')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		// Create database
		_, err = db.Exec("CREATE DATABASE testvector")
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Println("Database 'testvector' created successfully")
	} else {
		log.Println("Database 'testvector' already exists")
	}

	return nil
}

func connectTestVectorDB(cfg *Config) (*sql.DB, error) {
	// Create connection string for testvector database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=testvector sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func setupTableA(db *sql.DB) error {
	ctx := context.Background()

	// Create pgvector extension if not exists
	_, err := db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}

	log.Println("pgvector extension created successfully in testvector database")

	// Create table a with vector column (dimension depends on model)
	// bge-m3:latest produces 1024-dimensional embeddings
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS a (
		id SERIAL PRIMARY KEY,
		content TEXT NOT NULL,
		chunk_index INTEGER NOT NULL,
		source_file TEXT NOT NULL,
		embedding vector(1024),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.ExecContext(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table a: %w", err)
	}

	log.Println("Table 'a' created successfully")

	// Create index for vector similarity search
	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS a_embedding_idx 
	ON a USING ivfflat (embedding vector_cosine_ops)
	WITH (lists = 100)`

	_, err = db.ExecContext(ctx, createIndexSQL)
	if err != nil {
		// Index creation might fail if there's no data yet, which is okay
		log.Printf("Warning: failed to create index (this is normal if table is empty): %v", err)
	} else {
		log.Println("Vector index created successfully on table 'a'")
	}

	return nil
}

func getEmbedding(cfg *Config, text string) ([]float32, error) {
	reqBody := OllamaEmbeddingRequest{
		Model:  cfg.OllamaModel,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", cfg.OllamaHost)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var embResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embResp.Embedding, nil
}

func chunkText(text string, maxChunkSize int) []string {
	// Split by paragraphs first
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var currentChunk strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// If adding this paragraph would exceed max size, save current chunk
		if currentChunk.Len()+len(para) > maxChunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)

		// If current chunk is already too large, save it
		if currentChunk.Len() >= maxChunkSize {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}
	}

	// Add remaining chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

func processDocuments(db *sql.DB, cfg *Config, docDir string) error {
	// Read all markdown files in the directory
	files, err := filepath.Glob(filepath.Join(docDir, "*.md"))
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		log.Println("No markdown files found in", docDir)
		return nil
	}

	log.Printf("Found %d markdown files to process\n", len(files))

	for _, file := range files {
		log.Printf("Processing file: %s\n", file)

		// Read file content
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Failed to read file %s: %v\n", file, err)
			continue
		}

		// Chunk the text (max 1000 characters per chunk)
		chunks := chunkText(string(content), 1000)
		log.Printf("Split %s into %d chunks\n", filepath.Base(file), len(chunks))

		// Process each chunk
		for i, chunk := range chunks {
			log.Printf("Processing chunk %d/%d from %s\n", i+1, len(chunks), filepath.Base(file))

			// Get embedding from Ollama
			embedding, err := getEmbedding(cfg, chunk)
			if err != nil {
				log.Printf("Failed to get embedding for chunk %d: %v\n", i, err)
				continue
			}

			log.Printf("Received embedding with dimension: %d\n", len(embedding))

			// Convert embedding to PostgreSQL array format
			vectorStr := "["
			for j, v := range embedding {
				if j > 0 {
					vectorStr += ","
				}
				vectorStr += fmt.Sprintf("%f", v)
			}
			vectorStr += "]"

			// Insert into database
			insertSQL := `
			INSERT INTO a (content, chunk_index, source_file, embedding) 
			VALUES ($1, $2, $3, $4)`

			_, err = db.Exec(insertSQL, chunk, i, filepath.Base(file), vectorStr)
			if err != nil {
				log.Printf("Failed to insert chunk %d: %v\n", i, err)
				continue
			}

			log.Printf("Successfully saved chunk %d/%d from %s\n", i+1, len(chunks), filepath.Base(file))

			// Small delay to avoid overwhelming the API
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

func getRandomQuery(db *sql.DB) (string, error) {
	// Get a random chunk from the database
	var content string
	err := db.QueryRow("SELECT content FROM a ORDER BY RANDOM() LIMIT 1").Scan(&content)
	if err != nil {
		return "", fmt.Errorf("failed to get random query: %w", err)
	}

	// Extract a portion of the content (first 100-200 characters)
	words := strings.Fields(content)
	if len(words) > 20 {
		// Take 10-20 random words from the content
		start := rand.Intn(len(words) - 10)
		end := start + 10 + rand.Intn(10)
		if end > len(words) {
			end = len(words)
		}
		return strings.Join(words[start:end], " "), nil
	}
	return content, nil
}

func performSimilaritySearch(db *sql.DB, cfg *Config, query string, topK int) error {
	log.Printf("\n=== Similarity Search ===")
	log.Printf("Query: %s\n", query)

	// Get embedding for the query
	log.Println("Generating embedding for query...")
	embedding, err := getEmbedding(cfg, query)
	if err != nil {
		return fmt.Errorf("failed to get query embedding: %w", err)
	}

	log.Printf("Query embedding dimension: %d\n", len(embedding))

	// Convert embedding to PostgreSQL array format
	vectorStr := "["
	for j, v := range embedding {
		if j > 0 {
			vectorStr += ","
		}
		vectorStr += fmt.Sprintf("%f", v)
	}
	vectorStr += "]"

	// Perform similarity search using cosine distance
	searchSQL := fmt.Sprintf(`
		SELECT id, source_file, chunk_index, content,
		       1 - (embedding <=> '%s'::vector) as similarity
		FROM a
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> '%s'::vector
		LIMIT %d
	`, vectorStr, vectorStr, topK)

	rows, err := db.Query(searchSQL)
	if err != nil {
		return fmt.Errorf("failed to execute search: %w", err)
	}
	defer rows.Close()

	log.Println("\n=== Search Results ===")
	count := 0
	for rows.Next() {
		var id, chunkIndex int
		var sourceFile, content string
		var similarity float64
		if err := rows.Scan(&id, &sourceFile, &chunkIndex, &content, &similarity); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		count++
		preview := content
		if len(preview) > 150 {
			preview = preview[:150] + "..."
		}
		log.Printf("\n[Result %d] (Similarity: %.4f)", count, similarity)
		log.Printf("  File: %s, Chunk: %d", sourceFile, chunkIndex)
		log.Printf("  Content: %s", preview)
	}

	if count == 0 {
		log.Println("No results found.")
	}

	return nil
}

func main() {
	// Parse command-line flags
	rebuild := flag.Bool("rebuild", false, "Rebuild the vector database from documents")
	flag.Parse()

	log.Println("Starting Vector Database Application...")

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Load configuration
	cfg := loadConfig()

	// Connect to testvector database
	log.Println("\n=== Connecting to testvector database ===")
	db, err := connectTestVectorDB(cfg)
	if err != nil {
		// If database doesn't exist, create it
		log.Println("Database doesn't exist, creating...")
		if err := createTestVectorDB(cfg); err != nil {
			log.Fatalf("Failed to create testvector database: %v", err)
		}
		db, err = connectTestVectorDB(cfg)
		if err != nil {
			log.Fatalf("Failed to connect to testvector database: %v", err)
		}
	}
	defer db.Close()

	log.Printf("Successfully connected to testvector database at %s:%s", cfg.DBHost, cfg.DBPort)

	if *rebuild {
		// Rebuild mode: Process documents and create embeddings
		log.Println("\n=== REBUILD MODE ===")

		// Setup table 'a' with pgvector
		log.Println("\n=== Setting up table 'a' ===")
		if err := setupTableA(db); err != nil {
			log.Fatalf("Failed to setup table 'a': %v", err)
		}

		// Clear existing data
		log.Println("\n=== Clearing existing data ===")
		_, err = db.Exec("TRUNCATE TABLE a")
		if err != nil {
			log.Printf("Warning: failed to truncate table: %v", err)
		} else {
			log.Println("Existing data cleared")
		}

		// Process documents from doc/ directory
		log.Println("\n=== Processing documents ===")
		log.Printf("Using Ollama at %s with model %s\n", cfg.OllamaHost, cfg.OllamaModel)

		if err := processDocuments(db, cfg, "doc"); err != nil {
			log.Fatalf("Failed to process documents: %v", err)
		}

		// Display summary
		var totalCount int
		err = db.QueryRow("SELECT COUNT(*) FROM a").Scan(&totalCount)
		if err != nil {
			log.Printf("Failed to get total count: %v", err)
		} else {
			log.Printf("\n=== Build Complete ===")
			log.Printf("Total embeddings in table 'a': %d\n", totalCount)
		}

	} else {
		// Search mode: Perform random similarity search
		log.Println("\n=== SEARCH MODE ===")

		// Check if database has data
		var totalCount int
		err = db.QueryRow("SELECT COUNT(*) FROM a").Scan(&totalCount)
		if err != nil {
			log.Fatalf("Failed to query database: %v", err)
		}

		if totalCount == 0 {
			log.Println("No data found in database.")
			log.Println("Please run with -rebuild flag to build the vector database first.")
			log.Println("\nUsage: go run main.go -rebuild")
			return
		}

		log.Printf("Database contains %d embeddings\n", totalCount)

		// Perform 3 random searches
		for i := 0; i < 3; i++ {
			log.Printf("\n=== Random Search #%d ===", i+1)

			// Get random query from existing documents
			query, err := getRandomQuery(db)
			if err != nil {
				log.Printf("Failed to get random query: %v", err)
				continue
			}

			// Perform similarity search
			if err := performSimilaritySearch(db, cfg, query, 3); err != nil {
				log.Printf("Search failed: %v", err)
				continue
			}

			// Wait a bit before next search
			if i < 2 {
				time.Sleep(2 * time.Second)
			}
		}
	}

	log.Println("\n=== Application completed successfully! ===")
}
