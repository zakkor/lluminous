package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zakkor/server/toolfns"
)

func main() {
	indexedSnippets, err := toolfns.IndexDirectory(os.Args[1])
	if err != nil {
		panic(err)
	}

	os.MkdirAll("./indexed", 0755)

	for i, snippet := range indexedSnippets {
		file, err := os.Create(fmt.Sprintf("./indexed/%d.json", i))
		if err != nil {
			panic(err)
		}
		defer file.Close()
		err = json.NewEncoder(file).Encode(snippet)
		if err != nil {
			panic(err)
		}
	}
}
