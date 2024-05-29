package llm

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const port = 8082

type LlamaServer struct {
	Cmd       *exec.Cmd
	Addr      string
	ModelName string
}

func Serve(llamaPath, model string, options []string) *LlamaServer {
	path := filepath.Join(llamaPath, "server")

	// Note: `--log-format text` is vital, because `json` mode does not flush after every write.
	options = append(options, "--port", strconv.Itoa(port), "-m", filepath.Join(llamaPath, "models", model), "--log-format", "text")

	cmd := exec.Command(path, options...)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	err := cmd.Start()
	if err != nil {
		panic(err)
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
	fmt.Printf("Started llama.cpp server at %s\n", addr)
	return &LlamaServer{Cmd: cmd, Addr: addr, ModelName: filepath.Base(model)}
}

func (s *LlamaServer) Close() {
	s.Cmd.Process.Kill()
}
