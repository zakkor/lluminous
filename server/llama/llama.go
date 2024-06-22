package llama

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ServerInstance struct {
	Cmd       *exec.Cmd
	Addr      string
	ModelName string
}

func Serve(llamaPath string, port int, model string, options []string) (*ServerInstance, error) {
	path := filepath.Join(llamaPath, "server")

	// Note: `--log-format text` is vital, because `json` mode does not flush after every write.
	options = append(options, "--port", strconv.Itoa(port), "-m", filepath.Join(llamaPath, "models", model), "--log-format", "text")

	cmd := exec.Command(path, options...)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	// Wait for "HTTP server listening" to be written to stderr
	for {
		b := buf.String()
		if strings.Contains(string(b), "HTTP server listening") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Printf("Started llama.cpp server at %s, hosting %s\n", addr, model)
	return &ServerInstance{Cmd: cmd, Addr: addr, ModelName: filepath.Base(model)}, nil
}

func (s *ServerInstance) Kill() {
	s.Cmd.Process.Kill()
}

type ModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func ListLocalModels(llamaPath string) ([]ModelInfo, error) {
	// Open models directory and get a list of file names inside it
	dir, err := os.Open(filepath.Join(llamaPath, "models"))
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	files, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	// Get only files ending in .gguf
	models := []ModelInfo{}
	for _, file := range files {
		if !strings.HasPrefix(file, "ggml-vocab") && filepath.Ext(file) == ".gguf" {
			models = append(models, ModelInfo{
				ID:   file,
				Name: strings.TrimSuffix(file, ".gguf"),
			})
		}
	}

	return models, nil
}
