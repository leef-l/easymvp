package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	body := bytes.NewReader([]byte(`{"content":"hello, stream test"}`))
	req, _ := http.NewRequest("POST", "http://127.0.0.1:8000/api/v3/projects/project_20260427000331.805806000_q/architect-chat/messages/stream", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer to handle large SSE lines
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, 4096)
	scanner.Buffer(buf, maxCapacity)
	start := time.Now()
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			count++
			fmt.Printf("[%5.1fs] len=%d %s\n", time.Since(start).Seconds(), len(line), line[:min(200, len(line))])
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("scanner error:", err)
	}
	fmt.Printf("\nTotal events: %d, elapsed: %.1fs\n", count, time.Since(start).Seconds())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
