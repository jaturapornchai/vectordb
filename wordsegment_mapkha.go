package main

import (
	"log"
	"strings"

	m "github.com/veer66/mapkha"
)

var wordcutter *m.Wordcut

// InitWordSegmentation โหลด dictionary สำหรับตัดคำไทย
func initWordSegmentation() error {
	// ลองโหลด dictionary จาก mapkha
	// ถ้าไม่ได้ก็ไม่เป็นไร - ใช้ simple cleanup แทน
	dict, err := m.LoadDefaultDict()
	if err != nil {
		log.Printf("⚠️  mapkha dictionary ไม่พบ (ok ไม่ขัดข้อง): %v", err)
		log.Printf("   → ใช้ simple cleanup แทน (ลบ special characters)")
		return nil // ไม่ crash - ยังคงทำงานต่อได้
	}
	wordcutter = m.NewWordcut(dict)
	log.Printf("✅ Word Segmentation พร้อมใช้งาน (mapkha)")
	return nil
}

// SegmentThaiText ตัดคำภาษาไทยให้แยกออก (fallback: รีเทิร์นคำเดิมถ้า wordcutter ไม่พร้อม)
func segmentThaiText(text string) []string {
	if wordcutter == nil {
		// Fallback: ถ้าไม่มี wordcutter ก็รีเทิร์นคำเดิม
		// แต่สาธารณะการลบ special characters จะทำแล้วใน cleanSpecialCharacters
		return []string{text}
	}

	// ตัดคำ
	segments := wordcutter.Segment(text)

	// ทำความสะอาด - ลบ space และคำว่าง
	var cleanedSegments []string
	for _, seg := range segments {
		cleaned := strings.TrimSpace(seg)
		// เก็บเฉพาะคำที่มีความยาว >= 2 และไม่มีช่องว่าง
		if cleaned != "" && len(cleaned) >= 2 && !strings.Contains(cleaned, " ") {
			cleanedSegments = append(cleanedSegments, cleaned)
		}
	}

	return cleanedSegments
}

// ExtractSearchKeywords ดึงคำค้นหาสำคัญจาก Ollama response
// และตัดคำทั้ง compound words ด้วย
func extractSearchKeywords(ollmamaResponse string) []string {
	// ทำความสะอาด: ลบสัญญาลักษณ์พิเศษออก
	cleaned := cleanSpecialCharacters(ollmamaResponse)

	// แยกคำค้นหา (คั่นด้วย |)
	keywords := strings.Split(cleaned, "|")

	uniqueKeywords := make(map[string]bool)
	var result []string

	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)

		// ลบ newline, dash, bullet
		kw = strings.ReplaceAll(kw, "\n", " ")
		kw = strings.ReplaceAll(kw, "\r", " ")
		kw = strings.Trim(kw, "- •\t")
		kw = strings.TrimSpace(kw)

		// ข้ามคำว่างและคำสั้น
		if kw == "" || len(kw) < 2 {
			continue
		}

		// เพิ่มคำทั้งหมด
		lower := strings.ToLower(kw)
		if !uniqueKeywords[lower] {
			uniqueKeywords[lower] = true
			result = append(result, kw)

			// ตัดคำภาษาไทยถ้ามี
			if hasThaiCharacters(kw) {
				segments := segmentThaiText(kw)
				for _, seg := range segments {
					segLower := strings.ToLower(seg)
					// เพิ่มคำที่ตัด (ถ้ายังไม่มี)
					if !uniqueKeywords[segLower] && len(seg) >= 2 {
						uniqueKeywords[segLower] = true
						result = append(result, seg)
					}
				}
			}
		}
	}

	return result
}

// cleanSpecialCharacters ลบสัญญาลักษณ์พิเศษ *, ", :, **, etc
func cleanSpecialCharacters(text string) string {
	// ลบสัญญาลักษณ์ที่ Ollama เพิ่มเข้ามา
	specialChars := []string{"*", "\"", "**", ":", "`", "- ", "• ", "【", "】", "《", "》", "「", "」"}
	result := text

	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "")
	}

	// ลบ multiple spaces
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return strings.TrimSpace(result)
}

// hasThaiCharacters ตรวจสอบว่ามีตัวอักษรไทยหรือไม่
func hasThaiCharacters(text string) bool {
	for _, r := range text {
		if r >= 0x0E00 && r <= 0x0E7F {
			return true
		}
	}
	return false
}
