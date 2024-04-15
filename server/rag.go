package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/gaspiman/cosine_similarity"
	"github.com/zakkor/server/llm"
)

func embedddd() {
	embedder := llm.Serve("ggml-sfr-embedding-mistral-q4_k_m.gguf", []string{"--embeddings"})

	embeddings, err := llm.EmbedFile(embedder, "./testdata/dune.txt")
	if err != nil {
		panic(err)
	}
	// Write embeddings to file, as json:
	embeddingsJSON, err := json.Marshal(embeddings)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("./testdata/dune.txt-embeddings.json", embeddingsJSON, 0644)
	if err != nil {
		panic(err)
	}
}

func rag(query string) string {
	embedder := llm.Serve("ggml-sfr-embedding-mistral-q4_k_m.gguf", []string{"--embeddings"})

	f, err := os.Open("./testdata/dune.txt-embeddings.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var embeddings []*llm.EmbedPair
	err = json.Unmarshal(data, &embeddings)
	if err != nil {
		panic(err)
	}

	queryEmbedding := llm.Embed(embedder, query)

	for _, embedding := range embeddings {
		if len(embedding.Embedding) == 0 {
			fmt.Println("empty embedding")
			continue
		}

		cosine, err := cosine_similarity.Cosine(queryEmbedding, embedding.Embedding)
		if err != nil {
			panic(err)
		}

		embedding.Similarity = cosine
	}

	// Sort embeddings by similarity score:
	sort.Slice(embeddings, func(i, j int) bool {
		return embeddings[i].Similarity > embeddings[j].Similarity
	})

	// Print top 10 most similar paragraphs:
	for i := 0; i < 10; i++ {
		fmt.Println(i, "- Similarity:", embeddings[i].Similarity)
		fmt.Println(embeddings[i].Content)
	}

	return embeddings[0].Content
}
