package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/zakkor/server/llm"
)

const llamaPath = "/Users/ed/src/llama.cpp"

func main() {
	var activeLlama *llm.LlamaServer

	// Start HTTP server
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
	}))

	router.Get("/models", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(map[string]any{
			"models": listLocalModels(),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(data)
	})

	router.Get("/model", func(w http.ResponseWriter, r *http.Request) {
		if activeLlama == nil {

			data, err := json.Marshal(map[string]any{
				"model": "",
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(data)
			return
		}
		data, err := json.Marshal(map[string]any{
			"model": activeLlama.ModelName,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})
	router.Post("/model", func(w http.ResponseWriter, r *http.Request) {
		var newModel map[string]any
		if err := json.NewDecoder(r.Body).Decode(&newModel); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if activeLlama != nil {
			activeLlama.Close()
		}
		activeLlama = llm.Serve(newModel["model"].(string), []string{"-c", "4096", "-ngl", "1", "-t", "8", "-tb", "12", "-b", "4096"})

	})

	router.Post("/tokenize_count", func(w http.ResponseWriter, r *http.Request) {
		var content map[string]any
		if err := json.NewDecoder(r.Body).Decode(&content); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if activeLlama == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		count := llm.TokenizeCount(activeLlama, content["content"].(string))

		fmt.Fprintf(w, "%d", count)
	})

	router.Get("/tool_schema", func(w http.ResponseWriter, r *http.Request) {
		schemas := []FunctionSchema{
			NewFunctionSchema(Ls),
		}
		schemasJSON, err := json.Marshal(schemas)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(schemasJSON)
	})

	router.Post("/tool", func(w http.ResponseWriter, r *http.Request) {
		// Get `name` and `arguments`
		var toolcall map[string]any
		if err := json.NewDecoder(r.Body).Decode(&toolcall); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Call the tool
		name := toolcall["name"].(string)
		arguments := toolcall["arguments"]
		tool, ok := tools[name]
		if !ok {
			panic("unknown tool: " + name)
		}
		content := tool(arguments)
		contentJSON, err := json.Marshal(content)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(contentJSON)
	})

	httpServer := &http.Server{Addr: ":8081", Handler: router}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Signal handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // Block until a signal is received.

	// Graceful shutdown
	if err := httpServer.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}

func listLocalModels() []string {
	// Open models directory and get a list of file names inside it
	dir, err := os.Open(filepath.Join(llamaPath, "models"))
	if err != nil {
		panic(err)
	}
	defer dir.Close()

	files, err := dir.Readdirnames(0)
	if err != nil {
		panic(err)
	}

	// Get only files ending in .gguf
	models := []string{}
	for _, file := range files {
		if !strings.HasPrefix(file, "ggml-vocab") && filepath.Ext(file) == ".gguf" {
			models = append(models, file)
		}
	}

	return models
}
