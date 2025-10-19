package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// OllamaQueryExpansionRequest for Ollama API
type OllamaQueryExpansionRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaQueryExpansionResponse from Ollama
type OllamaQueryExpansionResponse struct {
	Response string `json:"response"`
}

// ExpandQueryWithOllama ใช้ Ollama LLM ขยายคำค้นหา
func expandQueryWithOllama(cfg *Config, query string) []string {
	prompt := fmt.Sprintf(`คุณเป็นผู้เชี่ยวชาญด้านการค้นหาข้อมูลภาษาไทย

คำค้นหาของผู้ใช้: "%s"

กรุณาสร้างรายการคำค้นหาที่เกี่ยวข้อง โดย:
1. คำพ้องเสียงภาษาไทย (เช่น กระเบื้อง → กะเบื้อง)
2. คำภาษาอังกฤษที่เกี่ยวข้อง (เช่น กระเบื้อง → tile, roof tile)
3. แก้คำสะกดผิด (ถ้ามี)
4. คำที่เกี่ยวข้อง (เช่น กระเบื้อง → กระเบื้องหลังคา, กระเบื้องปูพื้น, กระเบื้องเซรามิก)
5. คำย่อหรือชื่อทางการค้า

ตอบเฉพาะคำค้นหาที่เกี่ยวข้อง แยกด้วยเครื่องหมาย | เท่านั้น ไม่ต้องอธิบาย
ตัวอย่าง: กระเบื้อง|กะเบื้อง|tile|roof tile|กระเบื้องหลังคา|กระเบื้องปูพื้น`, query)

	reqBody := OllamaQueryExpansionRequest{
		Model:  "llama3.2", // ใช้ model เล็กๆ เพื่อความเร็ว
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("❌ สร้าง JSON ไม่สำเร็จ: %v", err)
		return []string{query}
	}

	resp, err := http.Post(cfg.OllamaHost+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ เรียก Ollama ไม่สำเร็จ: %v", err)
		return []string{query}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Ollama API error: %s", string(body))
		return []string{query}
	}

	var ollamaResp OllamaQueryExpansionResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Printf("❌ Decode response ไม่สำเร็จ: %v", err)
		return []string{query}
	}

	// แยกคำค้นหา
	response := strings.TrimSpace(ollamaResp.Response)
	keywords := strings.Split(response, "|")

	// ทำความสะอาดและเก็บเฉพาะคำที่ไม่ซ้ำ
	uniqueKeywords := make(map[string]bool)
	var result []string

	// เพิ่ม query เดิมก่อนเสมอ
	result = append(result, query)
	uniqueKeywords[strings.ToLower(query)] = true

	for _, kw := range keywords {
		cleaned := strings.TrimSpace(kw)
		lower := strings.ToLower(cleaned)

		// ตรวจสอบว่าไม่ซ้ำ และไม่ใช่คำว่าง
		if cleaned != "" && !uniqueKeywords[lower] && len(cleaned) >= 2 {
			uniqueKeywords[lower] = true
			result = append(result, cleaned)
		}
	}

	// จำกัดไม่เกิน 10 คำค้นหา
	if len(result) > 10 {
		result = result[:10]
	}

	log.Printf("🔄 ขยายคำค้นหาได้: %v", result)

	return result
}

// SmartSearchKeywords รวมระบบขยายคำค้นหาอัจฉริยะ
func smartSearchKeywords(cfg *Config, query string) []string {
	// 1. ขยายคำค้นหาด้วย Ollama
	expandedQueries := expandQueryWithOllama(cfg, query)

	// 2. เพิ่มการแบ่งคำภาษาไทยแบบง่าย
	simpleWords := extractKeywords(query)

	// รวมทุกคำค้นหา
	allKeywords := append(expandedQueries, simpleWords...)

	// ลบซ้ำ
	uniqueKeywords := make(map[string]bool)
	var final []string

	for _, kw := range allKeywords {
		lower := strings.ToLower(kw)
		if !uniqueKeywords[lower] && len(kw) >= 2 {
			uniqueKeywords[lower] = true
			final = append(final, kw)
		}
	}

	return final
}
