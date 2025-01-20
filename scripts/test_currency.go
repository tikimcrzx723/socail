package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type UpdatePostPayload struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

func updatePost(postID int64, p UpdatePostPayload, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("http://localhost:8080/v1/posts/%d", postID)

	b, _ := json.Marshal(p)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("error creating request:", err)
		return
	}

	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Update response status:", resp.Status)
}

func main() {
	var wg sync.WaitGroup

	postID := 5

	wg.Add(2)
	content := "NEW CONTENT FROM USER B"
	title := "NEW TITLE FROM USER A"

	go updatePost(int64(postID), UpdatePostPayload{Title: &title}, &wg)
	go updatePost(int64(postID), UpdatePostPayload{Content: &content}, &wg)

	wg.Wait()
}
