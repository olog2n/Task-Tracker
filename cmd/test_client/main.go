package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"tracker/internal/model"
)

func main() {
	task := model.Task{Title: "Тест через Go", Author: "Me", Description: "Проверка"}
	body, _ := json.Marshal(task)

	resp, err := http.Post("http://localhost:6969/api/tasks", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)

}
