package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GeminiRequest สำหรับ Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// DeepSeekRequest สำหรับ DeepSeek API (OpenAI-compatible)
type DeepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []DeepSeekMessage `json:"messages"`
}

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekResponse struct {
	Choices []DeepSeekChoice `json:"choices"`
	Error   *DeepSeekError   `json:"error,omitempty"`
}

type DeepSeekChoice struct {
	Message DeepSeekMessage `json:"message"`
}

type DeepSeekError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// summarizeWithGemini ใช้ Gemini API สรุปผลลัพธ์
func summarizeWithGemini(apiKey string, query string, results []SearchResult) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY ไม่ได้ตั้งค่า")
	}

	// สร้าง prompt
	prompt := fmt.Sprintf(`คำถามจากผู้ใช้: %s

ผลการค้นหาที่เกี่ยวข้อง:
`, query)

	for i, result := range results {
		prompt += fmt.Sprintf("\n%d. [Similarity: %.4f]\n%s\n", i+1, result.Similarity, result.Content)
	}

	prompt += "\nกรุณาตอบคำถามจากข้อมูลข้างต้นอย่างชัดเจน โดยใช้รูปแบบ Markdown (สามารถใช้ heading, bullet points, numbering, bold, italic ได้)"

	// สร้าง request
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// ส่ง request
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error: %s", string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", err
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("ไม่มีคำตอบจาก Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// summarizeWithDeepSeek ใช้ DeepSeek API สรุปผลลัพธ์
func summarizeWithDeepSeek(apiKey string, query string, results []SearchResult) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("DEEPSEEK_API_KEY ไม่ได้ตั้งค่า")
	}

	// สร้าง prompt
	prompt := fmt.Sprintf(`คำถามจากผู้ใช้: %s

ผลการค้นหาที่เกี่ยวข้อง:
`, query)

	for i, result := range results {
		prompt += fmt.Sprintf("\n%d. [Similarity: %.4f]\n%s\n", i+1, result.Similarity, result.Content)
	}

	prompt += "\nกรุณาตอบคำถามจากข้อมูลข้างต้นอย่างชัดเจน โดยใช้รูปแบบ Markdown (สามารถใช้ heading, bullet points, numbering, bold, italic ได้)"

	// สร้าง request
	reqBody := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// สร้าง HTTP request
	req, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// ส่ง request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API error: %s", string(body))
	}

	var deepseekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return "", err
	}

	if deepseekResp.Error != nil {
		return "", fmt.Errorf("DeepSeek error: %s", deepseekResp.Error.Message)
	}

	if len(deepseekResp.Choices) == 0 {
		return "", fmt.Errorf("ไม่มีคำตอบจาก DeepSeek")
	}

	return deepseekResp.Choices[0].Message.Content, nil
}

// summarizeResults สรุปผลลัพธ์ด้วย AI (ลอง Gemini ก่อน แล้ว fallback เป็น DeepSeek)
func summarizeResults(cfg *Config, query string, results []SearchResult) string {
	if len(results) == 0 {
		return ""
	}

	// ลอง Gemini ก่อน
	summary, err := summarizeWithGemini(cfg.GeminiAPIKey, query, results)
	if err == nil {
		return summary
	}

	// ถ้า Gemini error ให้ลอง DeepSeek
	summary, err = summarizeWithDeepSeek(cfg.DeepSeekAPIKey, query, results)
	if err == nil {
		return summary
	}

	// ถ้าทั้ง 2 ตัว error ให้ return empty
	return ""
}
