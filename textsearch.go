package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Match represents a search result with context
type Match struct {
	LineNum   int
	Context   []string
	MatchLine int
	Filename  string
}

// SearchInFile searches for a word in a file and returns matches with context
func searchInFile(filename, searchWord string, beforeLines, afterLines int) []Match {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var matches []Match
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(searchWord)) {
			start := max(0, i-beforeLines)
			end := min(len(lines)-1, i+afterLines)

			context := make([]string, 0)
			for j := start; j <= end; j++ {
				context = append(context, lines[j])
			}

			matches = append(matches, Match{
				LineNum:   i + 1,
				Context:   context,
				MatchLine: i - start,
				Filename:  filename,
			})
		}
	}

	return matches
}

// SearchInDirectory searches for a word in all markdown files in a directory
func searchInDirectory(dirPath, shopID, searchWord string, beforeLines, afterLines int) []Match {
	var allMatches []Match
	var mu sync.Mutex
	var wg sync.WaitGroup

	// ค้นหาไฟล์ markdown ทั้งหมดก่อน
	var mdFiles []string
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})

	// ค้นหาแต่ละไฟล์แบบ concurrent (พร้อมกัน)
	for _, path := range mdFiles {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			matches := searchInFile(filePath, searchWord, beforeLines, afterLines)

			// ป้องกัน race condition ตอนเพิ่มผลลัพธ์
			mu.Lock()
			allMatches = append(allMatches, matches...)
			mu.Unlock()
		}(path)
	}

	// รอให้ทุก goroutine เสร็จ
	wg.Wait()

	return allMatches
}

// FormatMatchesForAI formats matches into text for AI summarization
func formatMatchesForAI(matches []Match, query string) string {
	if len(matches) == 0 {
		return fmt.Sprintf("ไม่พบข้อมูลที่เกี่ยวข้องกับ '%s'", query)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("พบ %d ผลลัพธ์ที่เกี่ยวข้องกับคำค้นหา '%s':\n\n", len(matches), query))

	// จำกัดไม่เกิน 20 matches เพื่อไม่ให้ context ยาวเกินไป
	maxMatches := min(len(matches), 20)

	for i := 0; i < maxMatches; i++ {
		match := matches[i]
		builder.WriteString(fmt.Sprintf("--- ผลลัพธ์ที่ %d (จากไฟล์: %s, บรรทัด: %d) ---\n",
			i+1, filepath.Base(match.Filename), match.LineNum))

		for j, line := range match.Context {
			if j == match.MatchLine {
				builder.WriteString(fmt.Sprintf(">>> %s <<<\n", line))
			} else if strings.TrimSpace(line) != "" {
				builder.WriteString(fmt.Sprintf("    %s\n", line))
			}
		}
		builder.WriteString("\n")
	}

	if len(matches) > maxMatches {
		builder.WriteString(fmt.Sprintf("... และอีก %d ผลลัพธ์\n", len(matches)-maxMatches))
	}

	return builder.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
