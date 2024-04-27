package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type EmbedRequest struct {
	Content string `json:"content"`
}

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func Embed(srv *LlamaServer, content string) []float64 {
	jsonData, err := json.Marshal(EmbedRequest{
		Content: content,
	})
	if err != nil {
		panic(err)
	}

	// Create a new request using http
	req, err := http.NewRequest("POST", srv.Addr+"/embedding", bytes.NewReader(jsonData))
	if err != nil {
		panic(err)
	}

	// Add headers to the request
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var embedResp EmbedResponse
	err = json.Unmarshal(body, &embedResp)
	if err != nil {
		panic(err)
	}

	return embedResp.Embedding
}

type EmbedPair struct {
	Content    string    `json:"content"`
	Embedding  []float64 `json:"embedding"`
	Similarity float64   `json:"-"`
}

func EmbedFile(srv *LlamaServer, path string) ([]EmbedPair, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Split into paragraphs:
	paragraphs := strings.Split(string(data), "\n\n")
	embeddings := make([]EmbedPair, len(paragraphs))
	for i, para := range paragraphs {
		embedding := Embed(srv, para)
		if len(embedding) == 0 {
			fmt.Println("empty embedding, paragraph length", len(para))
		} else {
			fmt.Println("was ok, para length length", len(para))
		}
		embeddings[i] = EmbedPair{Content: para, Embedding: embedding}
	}

	return embeddings, nil
}
