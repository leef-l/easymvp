package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func main() {
	for i := 0; i < 3; i++ {
		fmt.Printf("\n=== Request %d ===\n", i)
		body := bytes.NewReader([]byte(`{"content":"hello stream test ` + fmt.Sprintf("%d", i) + `"}`))
		req, _ := http.NewRequest("POST", "http://127.0.0.1:8000/api/v3/projects/project_20260427000331.805806000_q/architect-chat/messages/stream", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")
		client := &http.Client{Timeout: 120 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		scanner := bufio.NewScanner(resp.Body)
		const maxCapacity = 1024 * 1024
		buf := make([]byte, 4096)
		scanner.Buffer(buf, maxCapacity)
		start := time.Now()
		count := 0
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				count++
				fmt.Printf("[%4.1fs] %s\n", time.Since(start).Seconds(), line[:min(120, len(line))])
			}
		}
		resp.Body.Close()
		if err := scanner.Err(); err != nil {
			fmt.Println("scanner error:", err)
		}
		fmt.Printf("Total: %d events in %.1fs\n", count, time.Since(start).Seconds())
		time.Sleep(1 * time.Second)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
