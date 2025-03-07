package toolfns

import (
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/byte-sat/llum-tools/tools"
)

// Note: Generated filename is significant. The init function for the generated file must run first.
//go:generate go run github.com/noonien/codoc/cmd/codoc@latest -out generated_toolfns.go -pkg toolfns .

var ToolGroups []*Group

func init() {
	ToolGroups = []*Group{
		//NewGroup("Browser",
		//	BrowserOpen,
		//	BrowserClick,
		//	BrowserType,
		//),
		//NewGroup("Web",
		//	WebSearch,
		//	WebNavigate,
		//),
		NewGroup("System",
			Shell,
		),
	}
}

type Group struct {
	Name string      `json:"name"`
	Repo *tools.Repo `json:"-"`
}

func NewGroup(name string, fns ...any) *Group {
	repo, err := tools.New(nil, fns...)
	if err != nil {
		log.Fatal(err)
	}
	return &Group{
		Name: name,
		Repo: repo,
	}
}

type ContentTypeResponse struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

// Executes the given bash command and returns the output of the command.
// command: The bash command to execute.
func Shell(command string) string {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error() + "\n" + string(out)
	}
	return string(out)
}

// Searches the web for the given query and returns a list of search results, containing a pair of the page title and the page URL, formatted like `title => url`, with one search result per line.
// You can use WebNavigate with the page URL to retrieve more detailed information about that page.
// query: The search query to use.
func WebSearch(query string) string {
	results := GoogleResults(query)
	return strings.Join(results, "\n\n")
}

// Navigates to the given URL and returns the text content of the page.
// url: The URL to navigate to.
func WebNavigate(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	pageText, err := HtmlToText(resp.Body)
	if err != nil {
		return err.Error()
	}

	return pageText
}
