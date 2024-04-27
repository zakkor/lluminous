package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type TokenizeRequest struct {
	Content string `json:"content"`
}

type TokenizeResponse struct {
	Tokens []int `json:"tokens"`
}

func TokenizeCount(srv *LlamaServer, content string) int {
	tokReq := TokenizeRequest{
		Content: content,
	}

	jsonData, err := json.Marshal(tokReq)
	if err != nil {
		panic(err)
	}

	// Create a new request using http
	req, err := http.NewRequest("POST", srv.Addr+"/tokenize", bytes.NewReader(jsonData))
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

	var tokResp TokenizeResponse
	err = json.Unmarshal(body, &tokResp)
	if err != nil {
		panic(err)
	}

	return len(tokResp.Tokens)
}
