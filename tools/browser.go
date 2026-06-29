package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

type BrowserTool struct{}

func (b *BrowserTool) Name() string        { return "browser" }
func (b *BrowserTool) Description() string { return "Navigate and extract content from web pages using a headless browser." }

func (b *BrowserTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to navigate to.",
			},
			"action": map[string]any{
				"type":        "string",
				"enum":        []string{"navigate", "screenshot", "content"},
				"description": "Browser action: navigate to URL, take screenshot, or get page content.",
			},
			"selector": map[string]any{
				"type":        "string",
				"description": "CSS selector for extracting specific content.",
			},
		},
		"required": []string{"url", "action"},
	}
}

type browserArgs struct {
	URL      string `json:"url"`
	Action   string `json:"action"`
	Selector string `json:"selector"`
}

func (b *BrowserTool) Execute(ctx context.Context, raw json.RawMessage) Result {
	var args browserArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("invalid args: %v", err)}
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	defer allocCancel()

	tabCtx, tabCancel := chromedp.NewContext(allocCtx)
	defer tabCancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(tabCtx, 30*time.Second)
	defer timeoutCancel()

	switch args.Action {
	case "navigate":
		var title string
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(args.URL),
			chromedp.Title(&title),
		); err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		return Result{Success: true, Output: fmt.Sprintf("navigated to %s, title: %s", args.URL, title)}

	case "screenshot":
		var buf []byte
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(args.URL),
			chromedp.CaptureScreenshot(&buf),
		); err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		return Result{Success: true, Output: fmt.Sprintf("screenshot captured (%d bytes)", len(buf))}

	case "content":
		if args.Selector == "" {
			args.Selector = "body"
		}
		var html string
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(args.URL),
			chromedp.OuterHTML(args.Selector, &html),
		); err != nil {
			return Result{Success: false, Error: err.Error()}
		}
		return Result{Success: true, Output: html}

	default:
		return Result{Success: false, Error: "unknown action: " + args.Action}
	}
}
