package tools

import (
	"flag"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zakkor/server/funcreader"
)

var sandboxPath = flag.String("sandbox", "", "Path to the sandbox directory.")

var Tools = NewToolMap(
	WebSearch,
	WebNavigate,
	WebNavigateChrome,
	Shell,
	ShellSandboxed,
	CodeInterpreter,
	GoCompiler,
	ReadFile,
	WriteFile,
	ListDirectory,
	MailSearch,
	MailRead,
	EnterMaze,
	Move,
)

var maze [][]int = [][]int{
	{1, 3, 1, 1, 1},
	{1, 0, 0, 0, 1},
	{1, 0, 1, 1, 1},
	{1, 0, 0, 0, 1},
	{1, 1, 1, 2, 1},
}
var y int = 4
var x int = 3

func describeCell(value int) string {
	switch value {
	case 0:
		return "a walkable space"
	case 1:
		return "a wall"
	case 2:
		return "the maze entrance"
	case 3:
		return "the maze exit"
	}
	return "unknown"
}

func describeLocation() string {
	desc := "You are on " + describeCell(maze[y][x]) + "."
	if y > 0 {
		desc += "\nTo the north is " + describeCell(maze[y-1][x]) + "."
	}
	if y < len(maze)-1 {
		desc += "\nTo the south is " + describeCell(maze[y+1][x]) + "."
	}
	if x > 0 {
		desc += "\nTo the west is " + describeCell(maze[y][x-1]) + "."
	}
	if x < len(maze[0])-1 {
		desc += "\nTo the east is " + describeCell(maze[y][x+1]) + "."
	}
	return desc
}

// Enters the maze and returns the initial description of what is around you. Continue to navigate the maze using Move. Only call this function once.
func EnterMaze(args Arguments) string {
	return "You have entered the maze.\n" + describeLocation()
}

// Navigate through the maze. After each move you'll receive a new description of what is around you after you moved. You cannot move into walls. If you navigate in the direction of the exit, you will finish the maze.
// - direction (string): The direction to move in. Can be "north", "south", "east", or "west".
func Move(args Arguments) string {
	direction := args.String("direction")
	switch direction {
	case "north":
		if y > 0 && maze[y-1][x] != 1 {
			y--
		} else {
			return "You cannot move in that direction."
		}
	case "south":
		if y < len(maze)-1 && maze[y+1][x] != 1 {
			y++
		} else {
			return "You cannot move in that direction."
		}
	case "west":
		if x > 0 && maze[y][x-1] != 1 {
			x--
		} else {
			return "You cannot move in that direction."
		}
	case "east":
		if x < len(maze[0])-1 && maze[y][x+1] != 1 {
			x++
		} else {
			return "You cannot move in that direction."
		}
	}
	desc := describeLocation()
	if maze[y][x] == 3 {
		return "You have finished the maze!"
	}
	return desc
}

// Searches the mail app for emails whose subject field contains the search query, and returns a list of email subjects, one per line.
// The search results contain the email subject, a separator, followed by the email ID. You can use MailRead with the email ID to retrieve more detailed information about that email.
// - subjectContains (string): The search query to use.
func MailSearch(args Arguments) string {
	subjectContains := args.String("subjectContains")
	return run("osascript", "./tools/searchmail.scpt", subjectContains)
}

// Reads the email with the given ID and returns the email content.
// - id (string): The ID of the email to read.
func MailRead(args Arguments) string {
	id := args.String("id")
	return run("osascript", "./tools/getemail.scpt", id)
}

// Searches the web for the given query and returns a list of search results, containing a pair of the page title and the page URL, formatted like `title => url`, with one search result per line.
// You can use WebNavigate with the page URL to retrieve more detailed information about that page.
// - query (string): The search query to use.
func WebSearch(args Arguments) string {
	query := args.String("query")
	results := GoogleResults(query)
	return strings.Join(results, "\n\n")
}

// Navigates to the given URL and returns the text content of the page.
// - url (string): The URL to navigate to.
func WebNavigate(args Arguments) string {
	url := args.String("url")

	resp, err := http.Get(url)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	pageText, err := htmlToText(resp.Body)
	if err != nil {
		return err.Error()
	}

	return pageText
}

// Navigates to the given URL and returns the text content of the page.
// - url (string): The URL to navigate to.
func WebNavigateChrome(args Arguments) string {
	url := args.String("url")

	pageHtml := fetchPage(url)

	pageHtmlReader := strings.NewReader(pageHtml)

	pageText, err := htmlToText(pageHtmlReader)
	if err != nil {
		return err.Error()
	}

	return pageText
}

// Lists the files and directories in the given directory.
// - path (string): The path to the directory to list.
func ListDirectory(args Arguments) string {
	path := args.String("path")
	if *sandboxPath == "" {
		return "Sandbox path not set."
	}
	files, err := os.ReadDir(filepath.Join(*sandboxPath, path))
	if err != nil {
		return err.Error()
	}
	var names []string
	for _, file := range files {
		names = append(names, file.Name())
	}
	return strings.Join(names, "\n")
}

// Reads the content of the file at the given path and returns it.
// - path (string): The path to the file to read.
func ReadFile(args Arguments) string {
	path := args.String("path")
	if *sandboxPath == "" {
		return "Sandbox path not set."
	}
	content, err := os.ReadFile(filepath.Join(*sandboxPath, path))
	if err != nil {
		return err.Error()
	}
	return string(content)
}

// Evaluates the given JavaScript expression and returns the result of the evaluation.
// - expression (string): The JavaScript code to execute.
func CodeInterpreter(args Arguments) string {
	expression := args.String("expression")

	return runSandboxed("node", "-p", expression)
}

// Executes the given bash command and returns the output of the command.
// - command (string): The bash command to execute.
func ShellSandboxed(args Arguments) string {
	return runSandboxed("bash", "-c", args.String("command"))
}

// Executes the given bash command and returns the output of the command.
// - command (string): The bash command to execute.
func Shell(args Arguments) string {
	if *sandboxPath == "" {
		return "Sandbox path not set."
	}
	cmd := exec.Command("bash", "-c", args.String("command"))
	cmd.Dir = *sandboxPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error() + "\n" + string(out)
	}
	return string(out)
}

// Compiles and runs the given Go source code, returning the output of the program, or the compile errors if compilation was not successful.
// - source (string): The Go source code to compile and run.
func GoCompiler(args Arguments) string {
	source := args.String("source")

	if *sandboxPath == "" {
		return "Sandbox path not set."
	}
	file := filepath.Join(*sandboxPath, "sandbox.go")
	err := os.WriteFile(file, []byte(source), 0644)
	if err != nil {
		return err.Error()
	}

	build := exec.Command("go", "build", "-o", "sandbox", "sandbox.go")
	build.Dir = *sandboxPath
	out, err := build.CombinedOutput()
	if err != nil {
		return err.Error() + "\n" + string(out)
	}

	return runSandboxed("./sandbox")
}

// Writes the given content to the file at the given path. Don't forget to include the file contents.
// - path (string): The path to the file to write to.
// - content (string): The content to write to the file.
func WriteFile(args Arguments) string {
	path := args.String("path")
	content := args.String("content")
	if *sandboxPath == "" {
		return "Sandbox path not set."
	}
	f, err := os.OpenFile(filepath.Join(*sandboxPath, path), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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

func run(args ...string) string {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error() + "\n" + string(out)
	}
	return string(out)
}

func runSandboxed(args ...string) string {
	if *sandboxPath == "" {
		return "Sandbox path not set."
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err.Error()
	}
	cmd := exec.Command("sandbox-exec", append([]string{"-f", cwd + "/sandbox.sb"}, args...)...)
	cmd.Dir = *sandboxPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error() + "\n" + string(out)
	}
	return string(out)
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
	schema.Function.Name = funcreader.FuncName(f)

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
