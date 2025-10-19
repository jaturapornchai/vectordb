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

// ExpandQueryWithOllama ใช้ Ollama LLM ขยายคำค้นหา + แปลภาษา
func expandQueryWithOllama(cfg *Config, query string) []string {
	prompt := fmt.Sprintf(`คุณเป็นผู้เชี่ยวชาญด้านการค้นหาข้อมูลภาษาไทยและอังกฤษ

คำค้นหาของผู้ใช้: "%s"

กรุณาสร้างรายการคำค้นหาที่เกี่ยวข้อง โดย:
1. **แปลภาษา**: ถ้าเป็นไทยแปลเป็นอังกฤษ / ถ้าเป็นอังกฤษแปลเป็นไทย
2. คำพ้องเสียงภาษาไทย (เช่น กระเบื้อง กะเบื้อง)
3. แก้คำสะกดผิด (ถ้ามี)

ตอบเฉพาะคำค้นหาที่เกี่ยวข้อง แยกด้วยช่องว่างเท่านั้น ไม่ต้องอธิบาย

**สำคัญ**: คำที่เป็นคำประสม ให้แยกออกเป็นคำเดี่ยวด้วย เพื่อให้หาเจอง่าย 
ห้ามมีตัวอักษระพิเศษอื่นใด เช่น , . / \ ' " ( ) [ ] { } < > @ # $ % ^ & * - + = ~  ! ?
ไม่ต้องบอกสิ่งที่ ai คิด ต้องการผลลัพธ์อย่างเดียว
ให้มีทั้งคำติดกันและคำแยก เพื่อเพิ่มโอกาสหาเจอ ไม่ต้องมีเครื่องหมายพิเศษอื่นใด

ถ้าไม่สามารถหาคำที่เกี่ยวข้องได้ ให้ตอบคำเดียวว่า: fail`, query)

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
		return []string{"fail"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Ollama API error: %s", string(body))
		return []string{"fail"}
	}

	var ollamaResp OllamaQueryExpansionResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Printf("❌ Decode response ไม่สำเร็จ: %v", err)
		return []string{"fail"}
	}

	// ถ้า Ollama ตอบว่า "fail" → ใช้คำค้นหาเดิม
	response := strings.TrimSpace(ollamaResp.Response)
	if strings.ToLower(response) == "fail" {
		log.Printf("⚠️  Ollama ไม่สามารถขยายคำค้นหาได้ → ใช้คำเดิม")
		return []string{query}
	}

	// แยกคำค้นหาจาก Ollama + ตัดคำภาษาไทยด้วย mapkha
	result := extractSearchKeywords(response)

	// เพิ่ม query เดิมถ้ายังไม่มี
	originalLower := strings.ToLower(query)
	hasOriginal := false
	for _, r := range result {
		if strings.ToLower(r) == originalLower {
			hasOriginal = true
			break
		}
	}
	if !hasOriginal {
		result = append([]string{query}, result...)
	}

	// จำกัดไม่เกิน 15 คำค้นหา (เพิ่มจาก 10 เพราะตัดคำแล้ว)
	if len(result) > 15 {
		result = result[:15]
	}

	log.Printf("🔄 ขยายคำค้นหาได้ (พร้อม mapkha segmentation): %v", result)

	return result
}

// SmartSearchKeywords รวมระบบขยายคำค้นหาอัจฉริยะ
func smartSearchKeywords(cfg *Config, query string) []string {
	// 1. ขยายคำค้นหาด้วย Ollama (แปลภาษา + คำพ้องเสียง + คำที่เกี่ยวข้อง)
	expandedQueries := expandQueryWithOllama(cfg, query)

	// ถ้า Ollama fail → ใช้คำเดิม + แบ่งคำไทย
	if len(expandedQueries) == 1 && expandedQueries[0] == "fail" {
		log.Printf("⚠️  Ollama fail → ใช้คำค้นหาเดิม + แบ่งคำไทย")
		simpleWords := extractKeywords(query)
		return simpleWords
	}

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
