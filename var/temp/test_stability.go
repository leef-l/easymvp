package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func main() {
	const total = 20
	success := 0
	partial := 0
	fail := 0

	for i := 0; i < total; i++ {
		body := bytes.NewReader([]byte(fmt.Sprintf(`{"content":"stability test %d"}`, i)))
		req, _ := http.NewRequest("POST", "http://127.0.0.1:8000/api/v3/projects/project_20260427091545.263045500_e2/architect-chat/messages/stream", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")
		client := &http.Client{Timeout: 180 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[%2d] ERROR: %v\n", i, err)
			fail++
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		const maxCapacity = 1024 * 1024
		buf := make([]byte, 4096)
		scanner.Buffer(buf, maxCapacity)

		var events []string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				events = append(events, line)
			}
		}
		resp.Body.Close()
		if err := scanner.Err(); err != nil {
			fmt.Printf("[%2d] SCANNER_ERROR: %v | events=%d\n", i, err, len(events))
			fail++
			continue
		}

		lastEvent := ""
		if len(events) > 0 {
			lastEvent = events[len(events)-1]
		}
		hasDone := strings.Contains(lastEvent, `"execution.done"`) || strings.Contains(lastEvent, `"done":true`)
		isFallback := hasDone && len(events) < 5

		if hasDone {
			if isFallback {
				fmt.Printf("[%2d] SUCCESS(fallback): %d events\n", i, len(events))
			} else {
				fmt.Printf("[%2d] SUCCESS(stream):  %d events\n", i, len(events))
			}
			success++
		} else if len(events) > 0 {
			fmt.Printf("[%2d] PARTIAL: %d events (last=%s)\n", i, len(events), lastEvent[:min(80, len(lastEvent))])
			partial++
		} else {
			fmt.Printf("[%2d] EMPTY: no events\n", i)
			fail++
		}
		time.Sleep(4 * time.Second)
	}

	fmt.Printf("\n========== STABILITY REPORT ==========\n")
	fmt.Printf("Total:    %d\n", total)
	fmt.Printf("Success:  %d (%.0f%%)\n", success, float64(success)/float64(total)*100)
	fmt.Printf("Partial:  %d (%.0f%%)\n", partial, float64(partial)/float64(total)*100)
	fmt.Printf("Fail:     %d (%.0f%%)\n", fail, float64(fail)/float64(total)*100)
	fmt.Printf("Stable:   %s\n", func() string {
		if success == total {
			return "PASS"
		}
		if success+partial == total {
			return "MARGINAL"
		}
		return "UNSTABLE"
	}())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
