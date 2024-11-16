// generated @ 2024-11-07T18:04:41+02:00 by gendoc
package toolfns

import "github.com/noonien/codoc"

func init() {
	codoc.Register(codoc.Package{
		ID:   "github.com/zakkor/server/toolfns",
		Name: "toolfns",
		Doc:  "generated @ 2024-08-01T16:13:44+03:00 by gendoc",
		Functions: map[string]codoc.Function{
			"BrowserClick": {
				Name: "BrowserClick",
				Doc:  "BrowserClick clicks on the element with the given label and returns a screenshot of the page after clicking.\nlabel: The label of the element to click, which was present in a previous screenshot. The labels have a bright green background, and are sequentially numbered, like 1, 2, 3... etc, a unique number for each clickable element.",
				Args: []string{
					"label",
				},
			},
			"BrowserOpen": {
				Name: "BrowserOpen",
				Doc:  "BrowserOpen opens the given URL in a headful browser and returns a screenshot of the page. The screenshot will contain labels for clickable elements, which can be clicked using the BrowserClick function.\nurl: The URL to open.",
				Args: []string{
					"url",
				},
			},
			"BrowserType": {
				Name: "BrowserType",
				Doc:  "BrowserType types some text into an input with the given label and returns a screenshot of the page after typing.\nlabel: The label of the element to click, which was present in a previous screenshot. The labels have a bright green background, and are sequentially numbered, like 1, 2, 3... etc, a unique number for each clickable element.\ntext: The text to type into the input.",
				Args: []string{
					"label",
					"text",
				},
			},
			"CodeOpenFile": {
				Name: "CodeOpenFile",
				Doc:  "CodeOpenFile retrieves the contents of the file at the given path.\npath: The path to the file.",
				Args: []string{
					"path",
				},
			},
			"CodeSnippetEdit": {
				Name: "CodeSnippetEdit",
				Doc:  "CodeSnippetEdit updates the code at the given path, replacing the code that previously was between the sourceLineStart and sourceLineEnd positions.\npath: The path to the file.\nsourceLineStart: The line number to start code replacement from. Use the exact sourceLineStart previously obtained through CodeSearch.\nsourceLineEnd: The line number to end code replacement at. Use the exact sourceLineEnd previously obtained through CodeSearch\nupdatedCodeSnippet: The updated code snippet to insert between sourceLineStart and sourceLineEnd. Rewrite the entire code snippet previously obtained through CodeSearch.",
				Args: []string{
					"path",
					"sourceLineStart",
					"sourceLineEnd",
					"updatedCodeSnippet",
				},
			},
			"CodeSnippetSearch": {
				Name: "CodeSnippetSearch",
				Doc:  "CodeSnippetSearch searches for relevant snippets of code matching the search query.\nlocationQuery: A search query which should describe the location (in the filesystem) of the code you are looking for. If no location data can be inferred from the user's query, leave this blank.\nfunctionalityQuery: A search query which should describe the functionality of the code you are looking for.",
				Args: []string{
					"locationQuery",
					"functionalityQuery",
				},
			},
			"Cosine": {
				Name: "Cosine",
				Args: []string{
					"a",
					"b",
				},
				Results: []string{
					"cosine",
					"err",
				},
			},
			"EmbeddingVoyage": {
				Name: "EmbeddingVoyage",
				Args: []string{
					"text",
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
			"IndexDirectory": {
				Name: "IndexDirectory",
				Args: []string{
					"path",
				},
			},
			"IndexFile": {
				Name: "IndexFile",
				Args: []string{
					"path",
				},
			},
			"NewGroup": {
				Name: "NewGroup",
				Args: []string{
					"name",
					"fns",
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
			"findJavaScriptRanges": {
				Name: "findJavaScriptRanges",
				Args: []string{
					"node",
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
			"Group": {
				Name: "Group",
			},
			"IndexedSnippet": {
				Name: "IndexedSnippet",
			},
			"IndexedSnippetNoEmbedding": {
				Name: "IndexedSnippetNoEmbedding",
			},
		},
	})
}
