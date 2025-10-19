package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// SearchRequest for text search
type SearchRequestSimple struct {
	Query      string `json:"query"`
	UseSummary bool   `json:"useSummary"`
}

// SearchResponse for text search
type SearchResponseSimple struct {
	Query   string               `json:"query"`
	Results []SearchResultSimple `json:"results"`
	Total   int                  `json:"total"`
	Summary string               `json:"summary,omitempty"`
	Error   string               `json:"error,omitempty"`
}

type SearchResultSimple struct {
	Content  string `json:"content"`
	Filename string `json:"filename"`
	LineNum  int    `json:"line_number"`
}

func enableCORSSimple(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func healthHandlerSimple(w http.ResponseWriter, r *http.Request) {
	enableCORSSimple(w)
	if r.Method == "OPTIONS" {
		return
	}

	status := map[string]interface{}{
		"status":  "healthy",
		"service": "text-search-api",
		"message": "ค้นหาในไฟล์ markdown โดยตรง",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func searchHandlerSimple(w http.ResponseWriter, r *http.Request) {
	enableCORSSimple(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequestSimple
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := SearchResponseSimple{Error: "รูปแบบ JSON ไม่ถูกต้อง"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Query == "" {
		response := SearchResponseSimple{Error: "ต้องระบุคำค้นหา"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("🔍 ค้นหาคำว่า: '%s'", req.Query)

	// ใช้ Ollama ขยายคำค้นหา (แปลงภาษา, คำพ้องเสียง, แก้คำผิด, ทำนายคำ)
	keywords := smartSearchKeywords(cfg, req.Query)
	log.Printf("🧠 Ollama ขยายคำค้นหาได้ %d คำ: %v", len(keywords), keywords)

	// ค้นหาในโฟลเดอร์ doc/
	docPath := "./doc"

	// รวมผลลัพธ์จากทุกคำสำคัญ
	var allMatches []Match
	for _, keyword := range keywords {
		log.Printf("   🔎 ค้นหาคำ: '%s'", keyword)
		matches := searchInDirectory(docPath, "", keyword, 3, 3) // 3 บรรทัดก่อน-หลัง
		allMatches = append(allMatches, matches...)
		log.Printf("      พบ %d ผลลัพธ์", len(matches))
	}

	// ลบผลลัพธ์ซ้ำ
	uniqueMatches := removeDuplicateMatches(allMatches)
	log.Printf("📊 พบทั้งหมด %d ผลลัพธ์ (หลังลบซ้ำจาก %d)", len(uniqueMatches), len(allMatches))

	// แปลง matches เป็น SearchResultSimple format
	var results []SearchResultSimple
	for _, match := range uniqueMatches {
		// รวม context เป็น string เดียว
		contextText := strings.Join(match.Context, "\n")

		results = append(results, SearchResultSimple{
			Content:  contextText,
			Filename: filepath.Base(match.Filename),
			LineNum:  match.LineNum,
		})
	}

	// สร้างสรุปด้วย AI ถ้าต้องการ
	var summary string
	if req.UseSummary && len(uniqueMatches) > 0 {
		log.Printf("🤖 กำลังสรุปผลด้วย AI...")
		contextForAI := formatMatchesForAI(uniqueMatches, req.Query)
		summary = summarizeResultsSimple(contextForAI, req.Query)
		if summary != "" {
			log.Printf("✅ สรุปด้วย AI สำเร็จ")
		}
	}

	response := SearchResponseSimple{
		Query:   req.Query,
		Results: results,
		Total:   len(results),
		Summary: summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// summarizeResultsSimple calls AI to summarize search results
func summarizeResultsSimple(context, query string) string {
	// ลอง Gemini ก่อน
	summary, err := summarizeWithGeminiText(cfg.GeminiAPIKey, context, query)
	if err == nil && summary != "" {
		log.Printf("✅ ใช้ Gemini สรุปผลสำเร็จ")
		return summary
	}

	log.Printf("⚠️  Gemini ล้มเหลว, ลอง DeepSeek...")

	// ถ้า Gemini ล้มเหลว ลอง DeepSeek
	summary, err = summarizeWithDeepSeekText(cfg.DeepSeekAPIKey, context, query)
	if err == nil && summary != "" {
		log.Printf("✅ ใช้ DeepSeek สรุปผลสำเร็จ")
		return summary
	}

	log.Printf("❌ ทั้ง Gemini และ DeepSeek ล้มเหลว")
	return fmt.Sprintf("พบผลลัพธ์ที่เกี่ยวข้องกับ '%s'", query)
}
