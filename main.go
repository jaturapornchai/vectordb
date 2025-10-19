package main

import (
	"log"
	"net/http"
)

var cfg *Config

func main() {
	// โหลด config
	cfg = loadConfig()

	log.Println("🚀 เริ่มต้น Simple Text Search API Server")
	log.Println("📁 ค้นหาในโฟลเดอร์: ./doc")

	// Routes
	http.HandleFunc("/health", healthHandlerSimple)
	http.HandleFunc("/search", searchHandlerSimple)

	log.Println("✅ เปิดใช้งาน HTTP server ที่พอร์ต 8080")
	log.Println("  POST http://localhost:8080/search")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
