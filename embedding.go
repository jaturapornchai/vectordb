package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

func sanitizeText(text string) string {
	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, "")
	}

	var result strings.Builder
	for _, r := range text {
		if utf8.ValidRune(r) && r != 0 {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func getEmbedding(cfg *Config, text string) ([]float32, error) {
	text = sanitizeText(text)

	reqBody := OllamaEmbeddingRequest{
		Model:  cfg.OllamaModel,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(cfg.OllamaHost+"/api/embeddings", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", string(body))
	}

	var ollamaResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	return ollamaResp.Embedding, nil
}

// generateHyDE ใช้ LLM สร้างคำตอบสมมติจากคำถาม เพื่อเพิ่มความแม่นยำในการค้นหา
func generateHyDE(cfg *Config, query string) (string, error) {
	hydePrompt := fmt.Sprintf(`คุณเป็นผู้เชี่ยวชาญ

คำถาม: %s

โปรดเขียนย่อหน้าสั้นๆ (2-3 ประโยค) ที่ตอบคำถามนี้:`, query)

	reqBody := OllamaGenerateRequest{
		Model:  "llama3.2",
		Prompt: hydePrompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.5,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(cfg.OllamaHost+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama generate API error: %s", string(body))
	}

	var ollamaResp OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// rewriteQuery เขียนคำถามใหม่ให้ชัดเจนขึ้น
func rewriteQuery(cfg *Config, query string) (string, error) {
	rewritePrompt := fmt.Sprintf(`คุณเป็นผู้เชี่ยวชาญในการปรับปรุงคำถาม

คำถามเดิม: %s

โปรดเขียนคำถามใหม่ให้ชัดเจนและเฉพาะเจาะจงขึ้น เหมาะสำหรับการค้นหาข้อมูล:`, query)

	reqBody := OllamaGenerateRequest{
		Model:  "llama3.2",
		Prompt: rewritePrompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.3,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(cfg.OllamaHost+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama generate API error: %s", string(body))
	}

	var ollamaResp OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

func splitIntoChunks(text string, chunkSize int) []string {
	text = sanitizeText(text)
	words := strings.Fields(text)
	var chunks []string
	var currentChunk strings.Builder

	for _, word := range words {
		if currentChunk.Len()+len(word)+1 > chunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(word)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}
