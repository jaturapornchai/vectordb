package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Gemini API
func summarizeWithGeminiText(apiKey, context, query string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("ไม่มี Gemini API Key")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", apiKey)

	prompt := fmt.Sprintf(`กรุณาสรุปข้อมูลต่อไปนี้ที่เกี่ยวข้องกับคำถาม: "%s"

%s

กรุณาสรุปเป็นภาษาไทยอย่างกระชับและตรงประเด็น`, query, context)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error: %s", string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("ไม่มีข้อมูลจาก Gemini")
}

// DeepSeek API  
func summarizeWithDeepSeekText(apiKey, context, query string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("ไม่มี DeepSeek API Key")
	}

	url := "https://api.deepseek.com/chat/completions"

	prompt := fmt.Sprintf(`กรุณาสรุปข้อมูลต่อไปนี้ที่เกี่ยวข้องกับคำถาม: "%s"

%s

กรุณาสรุปเป็นภาษาไทยอย่างกระชับและตรงประเด็น`, query, context)

	reqBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DeepSeek API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("ไม่มีข้อมูลจาก DeepSeek")
}
