package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GoogleResults(query string) []string {
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
		href, ok := s.Attr("href")
		if !ok {
			return
		}

		// Search results contain a <h3> tag
		if s.Has("h3").Length() == 0 {
			return
		}

		// Extract ?q= query parameter
		url, err := url.Parse(href)
		if err != nil {
			panic(err)
		}

		link := url.Query().Get("q")
		if link == "" || !strings.HasPrefix(link, "http") {
			return
		}
		text := s.Find("h3").Text()
		results = append(results, fmt.Sprintf("%s => %s", text, link))
	})

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
