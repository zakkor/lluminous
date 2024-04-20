package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zakkor/server/funcreader"
)

var Tools = NewToolMap(
	javascript_interpreter,
	write_file,
	run_go_code,
)

// Evaluates the given JavaScript expression and returns the result of the evaluation.
// - expression (string): The JavaScript code to execute.
func javascript_interpreter(args Arguments) string {
	expression := args.String("expression")
	cmd := exec.Command("node", "-p", expression)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// Compiles and runs the given file containing Go source code, returning the output of the program, or the compile errors if compilation was not successful.
// - file (string): The path to the Go source code file to compile.
func run_go_code(args Arguments) string {
	file := args.String("file")
	cmd := exec.Command("go", "run", file)
	cmd.Dir = "/Users/ed/src/lluminous/server/sandbox"
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// Writes the given content to the file at the given path.
// - path (string): The path to the file to write to.
// - content (string): The content to write to the file.
func write_file(args Arguments) string {
	path := args.String("path")
	content := args.String("content")
	f, err := os.OpenFile(filepath.Join("/Users/ed/src/lluminous/server/sandbox", path), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err.Error()
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return err.Error()
	}
	return ""
}

func NewToolMap(fs ...Fn) map[string]Tool {
	tools := make(map[string]Tool)
	for _, f := range fs {
		tool := Tool{
			Schema: NewSchema(f),
			Fn:     f,
		}
		tools[tool.Schema.Function.Name] = tool
	}
	return tools
}

type Tool struct {
	Schema Schema
	Fn     Fn
}
type Fn func(Arguments) string
type Arguments map[string]any

func (args Arguments) String(name string) string { return args[name].(string) }
func (args Arguments) Int(name string) int       { return args[name].(int) }

type Schema struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  struct {
			Type       string `json:"type"`
			Properties map[string]struct {
				Type        string `json:"type"`
				Description string `json:"description"`
			} `json:"properties"`
		} `json:"parameters"`
	} `json:"function"`
}

func NewSchema(f func(Arguments) string) Schema {
	schema := Schema{}
	schema.Function.Name = strings.ToLower(funcreader.FuncName(f))

	// Get description and extract function description as well as arguments description
	desc := funcreader.FuncDescription(f)
	// Arguments begin at first "-", description is everything before that first "-"
	splitDesc := strings.Split(desc, "-")
	schema.Type = "function"
	schema.Function.Description = strings.TrimSpace(splitDesc[0])
	// Arguments are everything after the first "-"
	schema.Function.Parameters.Type = "object"
	schema.Function.Parameters.Properties = make(map[string]struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	})

	if len(splitDesc) < 2 {
		return schema
	}

	// Extract argument names and types
	args := strings.Split(splitDesc[1], "-")

	// Get func arguments types and names:
	// An argument line looks like this:
	// - person (string): The person to say hello to.
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		name := strings.Split(arg, " ")[0]
		type_ := strings.Split(strings.Split(arg, "(")[1], ")")[0]
		desc := strings.Split(arg, ":")[1][1:]

		schema.Function.Parameters.Properties[name] = struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		}{
			Type:        type_,
			Description: desc,
		}
	}

	return schema
}
