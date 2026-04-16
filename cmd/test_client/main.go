package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	body, err := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "secret123",
	})
	if err != nil {
		fmt.Println("JSON marshal error:", err)
		return
	}

	resp, err := http.Post(
		"http://localhost:6969/api/auth/register",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}
	fmt.Printf("Response: %s\n", string(respBody))
}
