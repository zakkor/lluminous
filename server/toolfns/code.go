package toolfns

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/svelte"
)

const VoyageAIURL = "https://api.voyageai.com/v1/embeddings"

type IndexedSnippet struct {
	SourcePath             string        `json:"sourcePath"`
	SourceLineStart        int           `json:"sourceLineStart"`
	SourceLineEnd          int           `json:"sourceLineEnd"`
	SourceCodeSnippet      string        `json:"sourceCodeSnippet"`
	LocationEmbedding      [1536]float64 `json:"locationEmbedding"`
	FunctionalityEmbedding [1536]float64 `json:"functionalityEmbedding"`
}

type IndexedSnippetNoEmbedding struct {
	SourcePath        string `json:"sourcePath"`
	SourceLineStart   int    `json:"sourceLineStart"`
	SourceLineEnd     int    `json:"sourceLineEnd"`
	SourceCodeSnippet string `json:"sourceCodeSnippet"`
}

// CodeSnippetSearch searches for relevant snippets of code matching the search query.
// locationQuery: A search query which should describe the location (in the filesystem) of the code you are looking for. If no location data can be inferred from the user's query, leave this blank.
// functionalityQuery: A search query which should describe the functionality of the code you are looking for.
func CodeSnippetSearch(locationQuery, functionalityQuery string) (IndexedSnippetNoEmbedding, error) {
	var locationEmbedding [1536]float64
	if locationQuery != "" {
		var err error
		locationEmbedding, err = EmbeddingVoyage(locationQuery)
		if err != nil {
			panic(err)
		}
	}

	functionalityEmbedding, err := EmbeddingVoyage(functionalityQuery)
	if err != nil {
		panic(err)
	}

	infos, err := os.ReadDir("./toolfns/cmd/embed/indexed")
	if err != nil {
		return IndexedSnippetNoEmbedding{}, err
	}

	var bestMatchDistance float64
	var bestMatchSnippet *IndexedSnippet

	for _, info := range infos {
		file, err := os.Open(filepath.Join("./toolfns/cmd/embed/indexed", info.Name()))
		if err != nil {
			panic(err)
		}
		defer file.Close()

		var snippet IndexedSnippet
		err = json.NewDecoder(file).Decode(&snippet)
		if err != nil {
			panic(err)
		}

		var locationDistance float64
		if locationQuery != "" {
			locationDistance, err = Cosine(snippet.LocationEmbedding[:], locationEmbedding[:])
			if err != nil {
				panic(err)
			}
		}

		functionalityDistance, err := Cosine(snippet.FunctionalityEmbedding[:], functionalityEmbedding[:])
		if err != nil {
			panic(err)
		}

		var combinedDistance float64
		if locationQuery == "" {
			combinedDistance = functionalityDistance
		} else {
			locationWeight := 0.1
			functionalityWeight := 0.9
			combinedDistance = (locationWeight*locationDistance + functionalityWeight*functionalityDistance) / (locationWeight + functionalityWeight)
		}

		if combinedDistance > bestMatchDistance {
			bestMatchDistance = combinedDistance
			bestMatchSnippet = &snippet
		}
	}

	return IndexedSnippetNoEmbedding{
		SourcePath:        bestMatchSnippet.SourcePath,
		SourceLineStart:   bestMatchSnippet.SourceLineStart,
		SourceLineEnd:     bestMatchSnippet.SourceLineEnd,
		SourceCodeSnippet: bestMatchSnippet.SourceCodeSnippet,
	}, nil
}

// CodeSnippetEdit updates the code at the given path, replacing the code that previously was between the sourceLineStart and sourceLineEnd positions.
// path: The path to the file.
// sourceLineStart: The line number to start code replacement from. Use the exact sourceLineStart previously obtained through CodeSearch.
// sourceLineEnd: The line number to end code replacement at. Use the exact sourceLineEnd previously obtained through CodeSearch
// updatedCodeSnippet: The updated code snippet to insert between sourceLineStart and sourceLineEnd. Rewrite the entire code snippet previously obtained through CodeSearch.
func CodeSnippetEdit(path string, sourceLineStart, sourceLineEnd int, updatedCodeSnippet string) error {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Read the file contents
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Validate line numbers
	if sourceLineStart < 0 || sourceLineEnd > len(lines) || sourceLineStart > sourceLineEnd {
		return fmt.Errorf("invalid line numbers")
	}

	// Replace the specified lines with the new code
	newLines := append(lines[:sourceLineStart], strings.Split(updatedCodeSnippet, "\n")...)
	newLines = append(newLines, lines[sourceLineEnd+1:]...)

	// Open the file for writing
	file, err = os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Write the updated contents back to the file
	writer := bufio.NewWriter(file)
	for _, line := range newLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %v", err)
	}

	return nil
}

// CodeOpenFile retrieves the contents of the file at the given path.
// path: The path to the file.
func CodeOpenFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

func IndexFile(path string) ([]IndexedSnippet, error) {
	sourceCode, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse Svelte
	svelteParser := sitter.NewParser()
	svelteParser.SetLanguage(svelte.GetLanguage())

	svelteTree, err := svelteParser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, err
	}

	svelteRoot := svelteTree.RootNode()

	// // Find JavaScript ranges
	// jsRange, err := findJavaScriptRanges(svelteRoot)
	// if err == nil {
	// 	// Parse JavaScript
	// 	jsParser := sitter.NewParser()
	// 	jsParser.SetLanguage(javascript.GetLanguage())
	// 	jsParser.SetIncludedRanges([]sitter.Range{jsRange})
	// 	jsTree, err := jsParser.ParseCtx(context.Background(), nil, sourceCode)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	jsRoot := jsTree.RootNode()
	// fmt.Println("JavaScript script tag AST:", jsRoot.String())
	// } else {
	// fmt.Println("No script tag found")
	// }

	locationEmbedding, err := EmbeddingVoyage(path)
	if err != nil {
		return nil, err
	}

	// Iterate through top-level nodes and embed their text contents
	var indexedSnippets []IndexedSnippet
	for i := 0; i < int(svelteRoot.ChildCount()); i++ {
		child := svelteRoot.Child(i)

		sourceCodeSnippet := child.Content(sourceCode)

		functionalityEmbedding, err := EmbeddingVoyage(sourceCodeSnippet)
		if err != nil {
			return nil, err
		}

		indexedSnippets = append(indexedSnippets, IndexedSnippet{
			SourcePath:             path,
			SourceLineStart:        int(child.StartPoint().Row),
			SourceLineEnd:          int(child.EndPoint().Row),
			SourceCodeSnippet:      sourceCodeSnippet,
			LocationEmbedding:      locationEmbedding,
			FunctionalityEmbedding: functionalityEmbedding,
		})
	}

	return indexedSnippets, nil
}

func IndexDirectory(path string) ([]IndexedSnippet, error) {
	var indexedSnippets []IndexedSnippet

	// If path is a file, index only it and return early
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fileInfo.IsDir() {
		return IndexFile(path)
	}

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".svelte") {
			return nil
		}

		snippets, err := IndexFile(path)
		if err != nil {
			return err
		}

		indexedSnippets = append(indexedSnippets, snippets...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return indexedSnippets, nil
}

var errNoScript = errors.New("no script tag found")

func findJavaScriptRanges(node *sitter.Node) (sitter.Range, error) {
	iter := sitter.NewIterator(node, sitter.BFSMode)
	for {
		n, err := iter.Next()
		if err != nil {
			// No script was found
			return sitter.Range{}, errNoScript
		}

		if n.Type() != "script_element" {
			continue
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			if child.Type() == "raw_text" {
				return child.Range(), nil
			}
		}
	}
}

func EmbeddingVoyage(text string) ([1536]float64, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"input": []string{text},
		"model": "voyage-code-2",
	})
	if err != nil {
		return [1536]float64{}, err
	}

	// Create a new request
	req, err := http.NewRequest("POST", VoyageAIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return [1536]float64{}, err
	}

	VOYAGEAI_API_KEY := os.Getenv("VOYAGEAI_API_KEY")

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+VOYAGEAI_API_KEY)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return [1536]float64{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return [1536]float64{}, errors.New("http response status not OK")
	}

	type embeddingVoyageResponse struct {
		Object string `json:"object"`
		Data   []struct {
			Object    string        `json:"object"`
			Embedding [1536]float64 `json:"embedding"`
			Index     int           `json:"index"`
		} `json:"data"`
		Model string `json:"model"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	embeddingResp := embeddingVoyageResponse{}
	err = json.NewDecoder(resp.Body).Decode(&embeddingResp)
	if err != nil {
		return [1536]float64{}, err
	}

	return embeddingResp.Data[0].Embedding, nil
}

func Cosine(a []float64, b []float64) (cosine float64, err error) {
	count := 0
	length_a := len(a)
	length_b := len(b)
	if length_a > length_b {
		count = length_a
	} else {
		count = length_b
	}
	sumA := 0.0
	s1 := 0.0
	s2 := 0.0
	for k := 0; k < count; k++ {
		if k >= length_a {
			s2 += math.Pow(b[k], 2)
			continue
		}
		if k >= length_b {
			s1 += math.Pow(a[k], 2)
			continue
		}
		sumA += a[k] * b[k]
		s1 += math.Pow(a[k], 2)
		s2 += math.Pow(b[k], 2)
	}
	if s1 == 0 || s2 == 0 {
		return 0.0, errors.New("vectors should not be null (all zeros)")
	}
	return sumA / (math.Sqrt(s1) * math.Sqrt(s2)), nil
}
