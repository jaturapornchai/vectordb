package main

import (
	"strings"
	"unicode"
)

// ThaiWordSegmenter แบ่งคำภาษาไทยอย่างง่าย
func segmentThaiWords(text string) []string {
	var words []string
	var currentWord strings.Builder

	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// ถ้าเป็นภาษาไทย
		if isThaiChar(r) {
			currentWord.WriteRune(r)
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) {
			// ถ้าเป็นภาษาอังกฤษหรือตัวเลข
			currentWord.WriteRune(r)
		} else {
			// ถ้าเป็นช่องว่างหรือสัญลักษณ์
			if currentWord.Len() > 0 {
				word := currentWord.String()
				if len(word) >= 2 { // เก็บเฉพาะคำที่มีความยาว >= 2 ตัวอักษร
					words = append(words, word)
				}
				currentWord.Reset()
			}
		}
	}

	// เก็บคำสุดท้าย
	if currentWord.Len() > 0 {
		word := currentWord.String()
		if len(word) >= 2 {
			words = append(words, word)
		}
	}

	return words
}

// isThaiChar ตรวจสอบว่าเป็นอักษรไทยหรือไม่
func isThaiChar(r rune) bool {
	return (r >= 0x0E00 && r <= 0x0E7F) // Unicode range for Thai
}

// ExtractKeywords แยกคำสำคัญจากข้อความ
func extractKeywords(query string) []string {
	// แบ่งคำ
	words := segmentThaiWords(query)

	// ถ้าไม่มีคำ ให้ใช้ query เดิม
	if len(words) == 0 {
		return []string{query}
	}

	// ลบคำซ้ำ
	uniqueWords := make(map[string]bool)
	var keywords []string

	for _, word := range words {
		lower := strings.ToLower(word)
		if !uniqueWords[lower] && len([]rune(word)) >= 2 {
			uniqueWords[lower] = true
			keywords = append(keywords, word)
		}
	}

	// ถ้ามีแค่ 1 คำและคำนั้นสั้น ให้ใช้ query เดิมด้วย
	if len(keywords) == 1 && len([]rune(keywords[0])) < 4 {
		keywords = append(keywords, query)
	}

	return keywords
}

// RemoveDuplicateMatches ลบผลลัพธ์ซ้ำ
func removeDuplicateMatches(matches []Match) []Match {
	seen := make(map[string]bool)
	var unique []Match

	for _, match := range matches {
		// สร้าง key จาก filename และ line number
		key := match.Filename + ":" + string(rune(match.LineNum))
		if !seen[key] {
			seen[key] = true
			unique = append(unique, match)
		}
	}

	return unique
}
