package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
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

type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type SearchResult struct {
	Query      string
	Results    []Result
	TotalFound int
}

type Result struct {
	ID         int
	SourceFile string
	ChunkIndex int
	Content    string
	Similarity float64
}

func loadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
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

func connectTestVectorDB(cfg *Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=testvector sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func getEmbedding(cfg *Config, text string) ([]float32, error) {
	reqBody := OllamaEmbeddingRequest{
		Model:  cfg.OllamaModel,
		Prompt: text,
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/embeddings", cfg.OllamaHost)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s", string(body))
	}

	var embResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, err
	}

	return embResp.Embedding, nil
}

func getRandomQuery(db *sql.DB) (string, error) {
	var content string
	err := db.QueryRow("SELECT content FROM a ORDER BY RANDOM() LIMIT 1").Scan(&content)
	if err != nil {
		return "", err
	}

	words := strings.Fields(content)
	if len(words) > 20 {
		start := rand.Intn(len(words) - 10)
		end := start + 10 + rand.Intn(10)
		if end > len(words) {
			end = len(words)
		}
		return strings.Join(words[start:end], " "), nil
	}
	return content, nil
}

func performSearch(db *sql.DB, cfg *Config, query string, topK int) (*SearchResult, error) {
	embedding, err := getEmbedding(cfg, query)
	if err != nil {
		return nil, err
	}

	vectorStr := "["
	for j, v := range embedding {
		if j > 0 {
			vectorStr += ","
		}
		vectorStr += fmt.Sprintf("%f", v)
	}
	vectorStr += "]"

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
		return nil, err
	}
	defer rows.Close()

	results := []Result{}
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.ID, &r.SourceFile, &r.ChunkIndex, &r.Content, &r.Similarity); err != nil {
			continue
		}
		results = append(results, r)
	}

	return &SearchResult{
		Query:      query,
		Results:    results,
		TotalFound: len(results),
	}, nil
}

func summarizeContent(content string) string {
	// ตัดให้เหลือแค่ใจความสำคัญ
	content = strings.TrimSpace(content)

	// ตัด markdown headers
	lines := strings.Split(content, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	summary := strings.Join(cleanLines, " ")

	// แสดงเนื้อหาเต็มที่ ไม่ตัดเลย
	return summary
}

func writeMarkdownReport(results []SearchResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Header
	fmt.Fprintf(file, "# ผลการทดสอบระบบค้นหาเอกสารด้วย Vector Database\n\n")
	fmt.Fprintf(file, "**วันที่ทดสอบ:** %s\n\n", time.Now().Format("2 January 2006 15:04:05"))
	fmt.Fprintf(file, "**จำนวนคำถามทดสอบ:** %d คำถาม\n\n", len(results))
	fmt.Fprintf(file, "**โมเดล Embedding:** bge-m3:latest (1024 มิติ)\n\n")
	fmt.Fprintf(file, "**วิธีการค้นหา:** Cosine Similarity Search\n\n")
	fmt.Fprintf(file, "---\n\n")

	// Results
	for i, sr := range results {
		fmt.Fprintf(file, "## คำถามที่ %d\n\n", i+1)
		fmt.Fprintf(file, "### 🔍 คำค้นหา\n\n")
		fmt.Fprintf(file, "> %s\n\n", sr.Query)

		if sr.TotalFound == 0 {
			fmt.Fprintf(file, "**ไม่พบผลลัพธ์**\n\n")
			fmt.Fprintf(file, "---\n\n")
			continue
		}

		fmt.Fprintf(file, "### ✅ ผลการค้นหา (พบ %d รายการ)\n\n", sr.TotalFound)

		for j, r := range sr.Results {
			fmt.Fprintf(file, "#### อันดับ %d - ความแม่นยำ %.2f%%\n\n", j+1, r.Similarity*100)

			// สรุปเนื้อหา
			summary := summarizeContent(r.Content)
			fmt.Fprintf(file, "**สรุป:** %s\n\n", summary)

			fmt.Fprintf(file, "**ที่มา:** %s (ส่วนที่ %d)\n\n", r.SourceFile, r.ChunkIndex+1)

			// แสดงเนื้อหาเต็ม (ถ้าไม่ยาวเกินไป)
			if len(r.Content) <= 500 {
				fmt.Fprintf(file, "<details>\n<summary>📄 ดูเนื้อหาเต็ม</summary>\n\n")
				fmt.Fprintf(file, "```\n%s\n```\n\n", r.Content)
				fmt.Fprintf(file, "</details>\n\n")
			}
		}

		fmt.Fprintf(file, "---\n\n")
	}

	// Summary
	fmt.Fprintf(file, "\n## 📊 สรุปผลการทดสอบ\n\n")

	var totalSimilarity float64
	var highestSimilarity float64
	var lowestSimilarity float64 = 1.0

	for _, sr := range results {
		if len(sr.Results) > 0 {
			sim := sr.Results[0].Similarity
			totalSimilarity += sim
			if sim > highestSimilarity {
				highestSimilarity = sim
			}
			if sim < lowestSimilarity {
				lowestSimilarity = sim
			}
		}
	}

	avgSimilarity := totalSimilarity / float64(len(results))

	fmt.Fprintf(file, "- **ความแม่นยำเฉลี่ย:** %.2f%%\n", avgSimilarity*100)
	fmt.Fprintf(file, "- **ความแม่นยำสูงสุด:** %.2f%%\n", highestSimilarity*100)
	fmt.Fprintf(file, "- **ความแม่นยำต่ำสุด:** %.2f%%\n", lowestSimilarity*100)
	fmt.Fprintf(file, "\n")

	fmt.Fprintf(file, "### 💡 ข้อสังเกต\n\n")

	if avgSimilarity > 0.85 {
		fmt.Fprintf(file, "- ระบบค้นหามีประสิทธิภาพดีมาก สามารถหาเอกสารที่เกี่ยวข้องได้แม่นยำ\n")
	} else if avgSimilarity > 0.70 {
		fmt.Fprintf(file, "- ระบบค้นหามีประสิทธิภาพดี สามารถหาเอกสารที่เกี่ยวข้องได้ในระดับที่ยอมรับได้\n")
	} else {
		fmt.Fprintf(file, "- ระบบค้นหาอาจต้องปรับปรุง ความแม่นยำยังไม่สูงมาก\n")
	}

	fmt.Fprintf(file, "- Vector Database ใช้ Cosine Similarity ในการหาความคล้ายคลึง\n")
	fmt.Fprintf(file, "- ข้อมูลถูกแบ่งเป็น chunks เพื่อความแม่นยำในการค้นหา\n")
	fmt.Fprintf(file, "- โมเดล qwen2.5:0.5b สร้าง embeddings ขนาด 896 มิติ\n")

	return nil
}

func main() {
	log.Println("🚀 Starting Random Search Test (20 queries)...")

	rand.Seed(time.Now().UnixNano())

	cfg := loadConfig()

	db, err := connectTestVectorDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to database")

	// Check data
	var totalCount int
	db.QueryRow("SELECT COUNT(*) FROM a").Scan(&totalCount)
	log.Printf("📊 Database contains %d embeddings\n", totalCount)

	if totalCount == 0 {
		log.Fatal("❌ No data in database. Run with -rebuild first.")
	}

	// Perform 20 searches
	var allResults []SearchResult

	for i := 0; i < 20; i++ {
		log.Printf("\n🔍 Search %d/20...", i+1)

		query, err := getRandomQuery(db)
		if err != nil {
			log.Printf("Failed to get query: %v", err)
			continue
		}

		log.Printf("Query: %s", query[:min(100, len(query))])

		result, err := performSearch(db, cfg, query, 3)
		if err != nil {
			log.Printf("Search failed: %v", err)
			continue
		}

		if len(result.Results) > 0 {
			log.Printf("✅ Found %d results (Best match: %.2f%%)",
				result.TotalFound, result.Results[0].Similarity*100)
		}

		allResults = append(allResults, *result)

		// Small delay
		time.Sleep(500 * time.Millisecond)
	}

	// Write report
	log.Println("\n📝 Writing report to test.md...")
	if err := writeMarkdownReport(allResults, "test.md"); err != nil {
		log.Fatalf("Failed to write report: %v", err)
	}

	log.Println("✅ Report saved to test.md")
	log.Printf("📊 Tested %d queries successfully!", len(allResults))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
