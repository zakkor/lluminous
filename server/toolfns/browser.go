package toolfns

import (
	"encoding/base64"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

var page playwright.Page

var labelScript = `
let i = 1;
document.querySelectorAll('a, button, input').forEach((el) => {
	el.setAttribute('data-label', i);
	i++;
});`
var labelStyle = `
a, button, input {
		position: relative !important;
		margin-top: 16px !important;
}
a::before, button::before, input::before {
		content: attr(data-label) !important;
		position: absolute !important;
		top: -30px !important;
		left: 50% !important;
		transform: translateX(-50%) !important;
		background-color: #00ffd2 !important;
		color: black !important;
		padding: 2px 6px !important;
		border-radius: 4px !important;
		font-size: 16px !important;
		font-weight: bold !important;
		white-space: nowrap !important;
		opacity: 1 !important;
}`

// BrowserOpen opens the given URL in a headful browser and returns a screenshot of the page. The screenshot will contain labels for clickable elements, which can be clicked using the BrowserClick function.
// url: The URL to open.
func BrowserOpen(url string) (ContentTypeResponse, error) {
	pw, err := playwright.Run()
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not launch Playwright: %w", err)
	}

	// browser, err := pw.Firefox.LaunchPersistentContext(
	// 	"/Users/ed/Library/Application Support/Firefox/Profiles/nmt1omde.default-release",
	// 	playwright.BrowserTypeLaunchPersistentContextOptions{
	// 		Headless: playwright.Bool(false),
	// 	},
	// )

	browser, err := pw.Chromium.LaunchPersistentContext(
		"/Users/ed/Library/Application Support/Google/Chrome/",
		playwright.BrowserTypeLaunchPersistentContextOptions{
			// Channel:  playwright.String("chrome-canary"),
			// Headless: playwright.Bool(false),
		},
	)
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not launch Chromium: %w", err)
	}
	page, err = browser.NewPage()
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not create page: %w", err)
	}

	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not goto: %w", err)
	}

	_, err = page.AddScriptTag(playwright.PageAddScriptTagOptions{
		Content: &labelScript,
	})
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not insert script: %w", err)
	}
	ss, err := page.Screenshot(playwright.PageScreenshotOptions{
		Scale: playwright.ScreenshotScaleCss,
		Style: &labelStyle,
	})
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not create screenshot: %w", err)
	}

	// Base64 encode the screenshot
	ssb := base64.StdEncoding.EncodeToString(ss)

	return ContentTypeResponse{
		ContentType: "image/png",
		Content:     "data:image/png;base64," + ssb,
	}, nil
}

// BrowserClick clicks on the element with the given label and returns a screenshot of the page after clicking.
// label: The label of the element to click, which was present in a previous screenshot. The labels have a bright green background, and are sequentially numbered, like 1, 2, 3... etc, a unique number for each clickable element.
func BrowserClick(label string) (ContentTypeResponse, error) {
	// Click on the element with the given label
	err := page.Locator("[data-label='" + label + "']").Click()
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not click: %w", err)
	}

	_, err = page.AddScriptTag(playwright.PageAddScriptTagOptions{
		Content: &labelScript,
	})
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not insert script: %w", err)
	}

	// Take a screenshot after clicking
	ss, err := page.Screenshot(playwright.PageScreenshotOptions{
		Scale: playwright.ScreenshotScaleCss,
		Style: &labelStyle,
	})
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not create screenshot: %w", err)
	}

	// Base64 encode the screenshot
	ssb := base64.StdEncoding.EncodeToString(ss)

	return ContentTypeResponse{
		ContentType: "image/png",
		Content:     "data:image/png;base64," + ssb,
	}, nil
}

// BrowserType types some text into an input with the given label and returns a screenshot of the page after typing.
// label: The label of the element to click, which was present in a previous screenshot. The labels have a bright green background, and are sequentially numbered, like 1, 2, 3... etc, a unique number for each clickable element.
// text: The text to type into the input.
func BrowserType(label, text string) (ContentTypeResponse, error) {
	// Click on the element with the given label
	err := page.Locator("[data-label='" + label + "']").Fill(text)
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not fill: %w", err)
	}

	_, err = page.AddScriptTag(playwright.PageAddScriptTagOptions{
		Content: &labelScript,
	})
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not insert script: %w", err)
	}

	// Take a screenshot after clicking
	ss, err := page.Screenshot(playwright.PageScreenshotOptions{
		Scale: playwright.ScreenshotScaleCss,
		Style: &labelStyle,
	})
	if err != nil {
		return ContentTypeResponse{}, fmt.Errorf("could not create screenshot: %w", err)
	}

	// Base64 encode the screenshot
	ssb := base64.StdEncoding.EncodeToString(ss)

	return ContentTypeResponse{
		ContentType: "image/png",
		Content:     "data:image/png;base64," + ssb,
	}, nil
}

// if _, err = page.Screenshot(playwright.PageScreenshotOptions{
// 	Path: playwright.String("foo.png"),
// }); err != nil {
// 	log.Fatalf("could not create screenshot: %v", err)
// }
// if err = browser.Close(); err != nil {
// 	log.Fatalf("could not close browser: %v", err)
// }
// if err = pw.Stop(); err != nil {
// 	log.Fatalf("could not stop Playwright: %v", err)
// }
