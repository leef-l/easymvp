package main

import (
	"bufio"
	"fmt"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/api/v3/workspace/projects/project-demo/events", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))
	fmt.Println("---")

	scanner := bufio.NewScanner(resp.Body)
	lines := 0
	for scanner.Scan() && lines < 10 {
		fmt.Println(scanner.Text())
		lines++
	}
}
