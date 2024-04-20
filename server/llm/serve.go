package llm

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var serverCount int32 = 8079

type LlamaServer struct {
	Cmd       *exec.Cmd
	Addr      string
	ModelName string
}

func Serve(model string, options []string, llamaPath string) *LlamaServer {
	port := int(atomic.AddInt32(&serverCount, 1))
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

	return &LlamaServer{Cmd: cmd, Addr: fmt.Sprintf("http://127.0.0.1:%d", port), ModelName: filepath.Base(model)}
}

func (s *LlamaServer) Close() {
	atomic.AddInt32(&serverCount, -1)
	s.Cmd.Process.Kill()
}
