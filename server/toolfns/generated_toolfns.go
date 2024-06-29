// generated @ 2024-06-29T15:27:26+03:00 by gendoc
package toolfns

import "github.com/noonien/codoc"

func init() {
	codoc.Register(codoc.Package{
		ID:   "github.com/zakkor/server/toolfns",
		Name: "toolfns",
		Doc:  "generated @ 2024-06-26T23:36:24+03:00 by gendoc",
		Functions: map[string]codoc.Function{
			"DisplayImage": {
				Name: "DisplayImage",
				Doc:  "Opens a specialized UI element that displays the image to the user so they can view it. You yourself cannot see this image, so do not try to provide any additional information about it.\nurl: URL of the image to display",
				Args: []string{
					"url",
				},
			},
			"GoogleResults": {
				Name: "GoogleResults",
				Args: []string{
					"query",
				},
			},
			"HtmlToText": {
				Name: "HtmlToText",
				Args: []string{
					"htmlr",
				},
			},
			"Shell": {
				Name: "Shell",
				Doc:  "Executes the given bash command and returns the output of the command.\ncommand: The bash command to execute.",
				Args: []string{
					"command",
				},
			},
			"WebNavigate": {
				Name: "WebNavigate",
				Doc:  "Navigates to the given URL and returns the text content of the page.\nurl: The URL to navigate to.",
				Args: []string{
					"url",
				},
			},
			"WebSearch": {
				Name: "WebSearch",
				Doc:  "Searches the web for the given query and returns a list of search results, containing a pair of the page title and the page URL, formatted like `title => url`, with one search result per line.\nYou can use WebNavigate with the page URL to retrieve more detailed information about that page.\nquery: The search query to use.",
				Args: []string{
					"query",
				},
			},
			"init": {
				Name: "init",
			},
			"removeDuplicateWhitespace": {
				Name: "removeDuplicateWhitespace",
				Args: []string{
					"input",
				},
			},
		},
		Structs: map[string]codoc.Struct{
			"ContentTypeResponse": {
				Name: "ContentTypeResponse",
			},
		},
	})
}
