package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	db  *sql.DB
	cfg *Config
)

func main() {
	cfg = loadConfig()

	var err error
	db, err = connectDB(cfg)
	if err != nil {
		log.Fatalf("ไม่สามารถเชื่อมต่อฐานข้อมูล: %v", err)
	}
	defer db.Close()

	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM a").Scan(&totalCount)
	if err != nil {
		log.Fatalf("ไม่สามารถตรวจสอบข้อมูลในฐานข้อมูล: %v", err)
	}

	log.Printf("ฐานข้อมูลมี %d embeddings", totalCount)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/clean", cleanShopHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/build", buildDocHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == "OPTIONS" {
			return
		}

		response := map[string]interface{}{
			"service": "Vector Database API",
			"version": "2.0.0",
			"endpoints": map[string]string{
				"GET /health":  "ตรวจสอบสถานะระบบ",
				"GET /stats":   "สถิติข้อมูลในระบบ (แยกตาม shop, file)",
				"POST /clean":  "ลบข้อมูลทั้งหมดของ shop (JSON: {\"shopid\": \"default\"})",
				"POST /search": "ค้นหาเนื้อหา (JSON: {\"query\": \"คำค้นหา\", \"shopid\": \"shop001\", \"limit\": 5})",
				"POST /build":  "สร้าง vectors (JSON: {\"shopid\": \"shop001\", \"filename\": \"doc01.md\"})",
			},
			"total_records": totalCount,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	port := getEnv("PORT", "8080")
	log.Printf("เริ่มต้น HTTP server บนพอร์ต %s", port)
	log.Printf("  GET  http://localhost:%s/health", port)
	log.Printf("  GET  http://localhost:%s/stats", port)
	log.Printf("  POST http://localhost:%s/clean", port)
	log.Printf("  POST http://localhost:%s/search", port)
	log.Printf("  POST http://localhost:%s/build", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server เริ่มต้นไม่สำเร็จ: %v", err)
	}
}
