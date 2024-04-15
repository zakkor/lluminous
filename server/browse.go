package main

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// func browse(question string) {
// 	// Generate a Google search query for the question
// 	queryPrompt := NewPrompt().
// 		System("You are a helpful assistant. Turn the user's question into a search query that can be used on Google. Answer with the query directly, and nothing else.").
// 		User("what are some crazy japanese traditions for christmas?").
// 		Assistant("crazy japanese christmas traditions").
// 		User(question)

// 	query := queryPrompt.Complete()

// 	fmt.Println("searching for \"" + query + "\"...")
// 	results := Google(query)
// 	for _, result := range results {
// 		fmt.Println("browsing", result, "...")
// 		resp, err := http.Get(result)
// 		if err != nil {
// 			panic(err)
// 		}
// 		defer resp.Body.Close()

// 		pageText, err := htmlToText(resp.Body)
// 		if err != nil {
// 			panic(err)
// 		}

// 		os.WriteFile("page.txt", []byte(pageText), 0644)

// 		questionPrompt := NewPrompt().
// 			System("You are a helpful assistant. Answer the user's question based on the provided context which is a web page. Base your answer exclusively based on the contents of the page.").
// 			User("Question: How does nuclear fusion work?\nContext: Dogs and cats do not get along together well.").
// 			Assistant("The page does not contain any information about nuclear fusion.").
// 			User("Question: " + question + "\nContext: " + pageText)

// 		tokenCount := 1 + tokenize(questionPrompt.String()) // BOS
// 		fmt.Println("completion with page context has", tokenCount, "tokens")
// 		if tokenCount > 7000 {
// 			fmt.Println("page is too large, skipping")
// 			continue
// 		}

// 		answer := questionPrompt.Complete()

// 		fmt.Println(answer)

// 		// Check if fulfilled
// 		fulfilledPrompt := NewPrompt().
// 			System("You are a helpful assistant. Determine if the question has been resolved by the given answer below.").
// 			User("Question: What is PPPeter's real name?" + "\nAnswer: The name of Peter has not been mentioned in the provided context. Could you please provide more information or context about PPeter so I can help find their name?").
// 			Assistant("no").
// 			User("Question: " + question + "\nAnswer: " + answer)

// 		fulfilled := fulfilledPrompt.CompleteOpts(CompleteOptions{
// 			Grammar: `root ::= "yes" | "no"`,
// 		})

// 		if fulfilled == "yes" {
// 			fmt.Println("finished")
// 			break
// 		}
// 	}
// }

func Google(query string) []string {
	// Do a HTTP GET request to Google
	resp, err := http.Get("https://www.google.com/search?rls=en&q=" + url.QueryEscape(query) + "&ie=UTF-8&oe=UTF-8")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	// Find all links
	var results []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok && strings.HasPrefix(href, "/url?q=") && !strings.Contains(href, "google") {
			// Extract ?q= query parameter
			url, err := url.Parse(href)
			if err != nil {
				panic(err)
			}
			results = append(results, url.Query().Get("q"))
		}
	})

	// Return first result for now
	return results
}

func htmlToText(htmlr io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(htmlr)
	if err != nil {
		return "", err
	}
	doc.Find("*").Filter("head, style, iframe, noscript, script, nav, img, header, footer, button, a").Remove()
	return removeDuplicateWhitespace(doc.Text()), nil
}

func removeDuplicateWhitespace(input string) string {
	// Lastly, trim space to remove leading and trailing whitespace.
	input = strings.TrimSpace(input)

	input = regexp.MustCompile(` {2,}`).ReplaceAllString(input, " ")
	input = regexp.MustCompile(`\t{2,}`).ReplaceAllString(input, "\t")
	input = regexp.MustCompile(`(\r\n){2,}`).ReplaceAllString(input, "\n")
	input = regexp.MustCompile(`(\n\r){2,}`).ReplaceAllString(input, "\n")
	input = regexp.MustCompile(`( \n){2,}`).ReplaceAllString(input, "\n")
	input = regexp.MustCompile(`(\n ){2,}`).ReplaceAllString(input, "\n")
	input = regexp.MustCompile(`\n{2,}`).ReplaceAllString(input, "\n")
	input = regexp.MustCompile(`\r{2,}`).ReplaceAllString(input, "\n")

	return input
}
