package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/byte-sat/llum-tools/schema"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/zakkor/server/llama"
	"github.com/zakkor/server/toolfns"
)

var (
	llamaPath = flag.String("llama", "", "Path to the llama.cpp directory. You only need this if you want to run local models using llama.cpp.")
	password  = flag.String("password", "", "Password for basic auth.")
)

const llamaPort = 8082

func main() {
	flag.Parse()

	var activeLlama *llama.ServerInstance

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}))
	r.Use(authMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Serve client:
	if embedStaticFiles {
		staticFS, err := fs.Sub(staticFiles, "dist-client")
		if err != nil {
			log.Fatal(err)
		}
		fileServer := http.FileServer(http.FS(staticFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			fileServer.ServeHTTP(w, r)
		})
	}

	th := &ToolHandler{Groups: toolfns.ToolGroups}
	r.Get("/tool_schema", th.ToolSchema)
	r.Post("/tool", th.InvokeTool)

	r.Get("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		models, err := llama.ListLocalModels(*llamaPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(map[string]any{"data": models}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/model", func(w http.ResponseWriter, r *http.Request) {
		modelMap := map[string]any{
			"model": "",
		}
		if activeLlama != nil {
			modelMap["model"] = activeLlama.ModelName
		}

		if err := json.NewEncoder(w).Encode(modelMap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	r.Post("/model", func(w http.ResponseWriter, r *http.Request) {
		var newModel map[string]any
		if err := json.NewDecoder(r.Body).Decode(&newModel); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if activeLlama != nil {
			activeLlama.Kill()
		}
		var err error
		activeLlama, err = llama.Serve(*llamaPath, llamaPort, newModel["model"].(string), []string{"-c", "4096", "-ngl", "1", "-t", "8", "-tb", "12", "-b", "4096"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("Running at http://localhost:8081")
	httpServer := &http.Server{Addr: ":8081", Handler: r}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Signal handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // Block until a signal is received.

	// Graceful shutdown
	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *password != "" && r.Header.Get("Authorization") != ("Basic "+*password) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type ToolHandler struct {
	Groups []*toolfns.Group
}

func (tr *ToolHandler) ToolSchema(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	type encodedGroup struct {
		Name   string            `json:"name"`
		Schema []schema.Function `json:"schema"`
	}

	var encodedGroups []encodedGroup
	for _, group := range tr.Groups {
		encodedGroups = append(encodedGroups, encodedGroup{
			Name:   group.Name,
			Schema: group.Repo.Schema(),
		})
	}
	err := enc.Encode(encodedGroups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (tr *ToolHandler) InvokeTool(w http.ResponseWriter, r *http.Request) {
	var call struct {
		ChatID string         `json:"chat_id"`
		Name   string         `json:"name"`
		Args   map[string]any `json:"arguments"`
	}
	if err := json.NewDecoder(io.TeeReader(r.Body, os.Stdout)).Decode(&call); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	for _, group := range tr.Groups {
		var out any
		out, err = group.Repo.Invoke(nil, call.Name, call.Args)
		if err != nil {
			continue
		}
		json.NewEncoder(w).Encode(out)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
