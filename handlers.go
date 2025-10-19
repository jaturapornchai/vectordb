package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM a").Scan(&totalCount)
	if err != nil {
		log.Printf("ข้อผิดพลาดฐานข้อมูล: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := StatusResponse{
		Status:       "healthy",
		TotalRecords: totalCount,
		Message:      "Vector database กำลังทำงาน",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Total records
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM a").Scan(&totalCount)
	if err != nil {
		log.Printf("ข้อผิดพลาด: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// By Shop
	rowsShop, err := db.Query("SELECT shopid, COUNT(*) FROM a GROUP BY shopid ORDER BY shopid")
	if err != nil {
		log.Printf("ข้อผิดพลาด: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rowsShop.Close()

	var byShop []ShopStats
	for rowsShop.Next() {
		var stat ShopStats
		rowsShop.Scan(&stat.ShopID, &stat.Count)
		byShop = append(byShop, stat)
	}

	// By File
	rowsFile, err := db.Query("SELECT filename, COUNT(*) FROM a GROUP BY filename ORDER BY filename")
	if err != nil {
		log.Printf("ข้อผิดพลาด: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rowsFile.Close()

	var byFile []FileStats
	for rowsFile.Next() {
		var stat FileStats
		rowsFile.Scan(&stat.Filename, &stat.Count)
		byFile = append(byFile, stat)
	}

	// By Shop + File
	rowsShopFile, err := db.Query("SELECT shopid, filename, COUNT(*) FROM a GROUP BY shopid, filename ORDER BY shopid, filename")
	if err != nil {
		log.Printf("ข้อผิดพลาด: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rowsShopFile.Close()

	var byShopFile []ShopFileStats
	for rowsShopFile.Next() {
		var stat ShopFileStats
		rowsShopFile.Scan(&stat.ShopID, &stat.Filename, &stat.Count)
		byShopFile = append(byShopFile, stat)
	}

	response := StatsResponse{
		TotalRecords: totalCount,
		ByShop:       byShop,
		ByFile:       byFile,
		ByShopFile:   byShopFile,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func cleanShopHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ShopID string `json:"shopid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]string{"error": "รูปแบบ JSON ไม่ถูกต้อง"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.ShopID == "" {
		response := map[string]string{"error": "ต้องระบุ ShopID"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("คำขอลบข้อมูล shopid: %s", req.ShopID)

	result, err := db.Exec("DELETE FROM a WHERE shopid = $1", req.ShopID)
	if err != nil {
		log.Printf("ไม่สามารถลบข้อมูล: %v", err)
		response := map[string]string{"error": fmt.Sprintf("ไม่สามารถลบข้อมูล: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	deletedRows, _ := result.RowsAffected()
	log.Printf("ลบข้อมูล shopid=%s สำเร็จ: %d รายการ", req.ShopID, deletedRows)

	response := map[string]interface{}{
		"shopid":  req.ShopID,
		"deleted": deletedRows,
		"message": fmt.Sprintf("ลบข้อมูล %d รายการสำเร็จ", deletedRows),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := SearchResponse{Error: "รูปแบบ JSON ไม่ถูกต้อง"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Query == "" {
		response := SearchResponse{Error: "ต้องระบุคำค้นหา"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.ShopID == "" {
		response := SearchResponse{Error: "ต้องระบุ ShopID"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Limit <= 0 {
		req.Limit = 5
	}

	log.Printf("คำขอค้นหา: %s (shopid: %s, limit: %d)", req.Query, req.ShopID, req.Limit)

	// ใช้ bge-m3 โดยตรง (ยกเลิก HyDE และ Query Rewriting)
	wordCount := len(strings.Fields(req.Query))
	searchText := req.Query
	technique := "bge-m3 (ตรง)"

	log.Printf("� ใช้ bge-m3 โดยตรง - คำถาม %d คำ: '%s'", wordCount, req.Query)

	// สร้าง embedding
	log.Printf("🔍 กำลังสร้าง embedding...")

	embedding, err := getEmbedding(cfg, searchText)
	if err != nil {
		log.Printf("❌ ไม่สามารถสร้าง embedding: %v", err)
		response := SearchResponse{
			Query:  req.Query,
			ShopID: req.ShopID,
			Error:  fmt.Sprintf("ไม่สามารถสร้าง embedding: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("✅ สร้าง embedding สำเร็จ (dimension: %d, technique: %s)", len(embedding), technique)

	embeddingStr := "["
	for i, val := range embedding {
		if i > 0 {
			embeddingStr += ","
		}
		embeddingStr += fmt.Sprintf("%f", val)
	}
	embeddingStr += "]"

	// ดึงผลลัพธ์ตาม limit ที่ร้องขอ
	query := `
		SELECT content, source_file, filename, shopid, chunk_index, 1 - (embedding <=> $1::vector) as similarity
		FROM a 
		WHERE shopid = $2
		ORDER BY embedding <=> $1::vector
		LIMIT $3`

	log.Printf("🔎 ค้นหาในฐานข้อมูล (shopid: %s, limit: %d)", req.ShopID, req.Limit)

	rows, err := db.Query(query, embeddingStr, req.ShopID, req.Limit)
	if err != nil {
		log.Printf("❌ ข้อผิดพลาดในการค้นหา: %v", err)
		response := SearchResponse{
			Query:  req.Query,
			ShopID: req.ShopID,
			Error:  fmt.Sprintf("การค้นหาล้มเหลว: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	var results []SearchResult
	var filteredCount int
	var allResults []SearchResult // เก็บทั้งหมดเพื่อ debug

	for rows.Next() {
		var result SearchResult
		err := rows.Scan(&result.Content, &result.File, &result.Filename, &result.ShopID, &result.Chunk, &result.Similarity)
		if err != nil {
			log.Printf("❌ ข้อผิดพลาดในการอ่านข้อมูล: %v", err)
			continue
		}

		allResults = append(allResults, result)

		// กรอง similarity > 0.15
		if result.Similarity > 0.15 {
			results = append(results, result)
			log.Printf("   📍 chunk #%d, similarity: %.4f, file: %s", result.Chunk, result.Similarity, result.Filename)
		} else {
			filteredCount++
		}
	}

	// แสดง Top 3 ที่ถูกกรองทิ้ง (เพื่อ debug)
	if filteredCount > 0 {
		log.Printf("🔍 กรองทิ้ง %d รายการ (similarity ≤ 0.15)", filteredCount)
		log.Printf("   Top 3 ที่ถูกกรอง:")
		for i := 0; i < 3 && i < len(allResults); i++ {
			if allResults[i].Similarity <= 0.15 {
				log.Printf("      - chunk #%d: %.4f (%s)", allResults[i].Chunk, allResults[i].Similarity, allResults[i].Filename)
			}
		}
	}
	log.Printf("📊 พบผลลัพธ์รวม %d รายการสำหรับ shopid: %s", len(results), req.ShopID)

	// สรุปผลลัพธ์ด้วย AI (Gemini -> DeepSeek)
	var summary string
	if len(results) > 0 {
		log.Printf("🤖 กำลังสรุปผลลัพธ์ด้วย AI...")
		summary = summarizeResults(cfg, req.Query, results)
		if summary != "" {
			log.Printf("✅ สรุปผลสำเร็จ")
		} else {
			log.Printf("⚠️  ไม่สามารถสรุปผลได้")
		}
	}

	response := SearchResponse{
		Query:   req.Query,
		ShopID:  req.ShopID,
		Results: results,
		Total:   len(results),
		Summary: summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func buildDocHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BuildDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := BuildDocResponse{Error: "รูปแบบ JSON ไม่ถูกต้อง"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.ShopID == "" || req.Filename == "" {
		response := BuildDocResponse{Error: "ต้องระบุ ShopID และ Filename"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("คำขอสร้าง doc: shopid=%s, filename=%s", req.ShopID, req.Filename)

	filePath := fmt.Sprintf("doc/%s", req.Filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("ไม่สามารถอ่านไฟล์ %s: %v", filePath, err)
		response := BuildDocResponse{
			ShopID:   req.ShopID,
			Filename: req.Filename,
			Error:    fmt.Sprintf("ไม่สามารถอ่านไฟล์: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	contentStr := string(content)
	log.Printf("อ่านไฟล์ %s สำเร็จ, ขนาด: %d ตัวอักษร", filePath, len(contentStr))

	deleteQuery := "DELETE FROM a WHERE shopid = $1 AND filename = $2"
	result, err := db.Exec(deleteQuery, req.ShopID, req.Filename)
	if err != nil {
		log.Printf("ไม่สามารถลบข้อมูลเก่า: %v", err)
		response := BuildDocResponse{
			ShopID:   req.ShopID,
			Filename: req.Filename,
			Error:    fmt.Sprintf("ไม่สามารถลบข้อมูลเก่า: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	deletedRows, _ := result.RowsAffected()
	log.Printf("ลบข้อมูลเก่า %d รายการ", deletedRows)

	chunks := splitIntoChunks(contentStr, 400)
	log.Printf("แบ่งเนื้อหาเป็น %d chunks (chunk size: 400 chars)", len(chunks))

	// ประมวลผลพร้อมกัน 100 threads
	const maxWorkers = 100
	var wg sync.WaitGroup
	chunkChan := make(chan int, len(chunks))
	successChan := make(chan bool, len(chunks))

	// Worker pool
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range chunkChan {
				chunk := chunks[i]
				embedding, err := getEmbedding(cfg, chunk)
				if err != nil {
					log.Printf("ไม่สามารถสร้าง embedding สำหรับ chunk %d: %v", i+1, err)
					successChan <- false
					continue
				}

				embeddingStr := "["
				for j, val := range embedding {
					if j > 0 {
						embeddingStr += ","
					}
					embeddingStr += fmt.Sprintf("%f", val)
				}
				embeddingStr += "]"

				insertQuery := `
					INSERT INTO a (content, source_file, filename, shopid, chunk_index, embedding) 
					VALUES ($1, $2, $3, $4, $5, $6::vector)`

				_, err = db.Exec(insertQuery, chunk, req.Filename, req.Filename, req.ShopID, i+1, embeddingStr)
				if err != nil {
					log.Printf("ไม่สามารถบันทึก chunk %d: %v", i+1, err)
					successChan <- false
					continue
				}

				successChan <- true
				if (i+1)%10 == 0 {
					log.Printf("ประมวลผลแล้ว %d/%d chunks", i+1, len(chunks))
				}
			}
		}()
	}

	// ส่ง jobs ไปยัง workers
	for i := range chunks {
		chunkChan <- i
	}
	close(chunkChan)

	// รอ workers ทั้งหมดเสร็จ
	wg.Wait()
	close(successChan)

	// นับจำนวนที่สำเร็จ
	successCount := 0
	for success := range successChan {
		if success {
			successCount++
		}
	}

	log.Printf("สร้าง embeddings สำเร็จ %d/%d chunks", successCount, len(chunks))

	// บันทึกประวัติการสร้างลง table shopidfilename
	upsertHistoryQuery := `
		INSERT INTO shopidfilename (shopid, filename, emailusercreate, createdate, updatedate)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (shopid, filename) 
		DO UPDATE SET updatedate = CURRENT_TIMESTAMP, emailusercreate = $3`

	emailUser := "system"
	_, err = db.Exec(upsertHistoryQuery, req.ShopID, req.Filename, emailUser)
	if err != nil {
		log.Printf("คำเตือน: ไม่สามารถบันทึกประวัติ: %v", err)
	} else {
		log.Printf("บันทึกประวัติการสร้าง shopid=%s, filename=%s", req.ShopID, req.Filename)
	}

	response := BuildDocResponse{
		ShopID:     req.ShopID,
		Filename:   req.Filename,
		Chunks:     len(chunks),
		Embeddings: successCount,
		Message:    fmt.Sprintf("ประมวลผลสำเร็จ %d จาก %d chunks", successCount, len(chunks)),
	}

	if successCount < len(chunks) {
		response.Error = fmt.Sprintf("บาง chunks ประมวลผลไม่สำเร็จ (%d/%d สำเร็จ)", successCount, len(chunks))
		w.WriteHeader(http.StatusPartialContent)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
