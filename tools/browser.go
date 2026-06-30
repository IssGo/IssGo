package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type BrowserTool struct{}

func (b *BrowserTool) Name() string { return "browser" }
func (b *BrowserTool) Description() string {
	return "Headless browser for navigation, screenshots, content extraction, and JS execution."
}

func (b *BrowserTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url":      map[string]any{"type": "string", "description": "URL to navigate to."},
			"action":   map[string]any{"type": "string", "enum": []string{"navigate", "screenshot", "content", "text", "title", "click", "eval"}, "description": "Browser action."},
			"selector": map[string]any{"type": "string", "description": "CSS selector for content/click/eval."},
			"script":   map[string]any{"type": "string", "description": "JS to evaluate (for eval action)."},
			"wait_ms":  map[string]any{"type": "number", "description": "Extra wait time in ms."},
		},
		"required": []string{"url", "action"},
	}
}

type browserArgs struct {
	URL      string  `json:"url"`
	Action   string  `json:"action"`
	Selector string  `json:"selector"`
	Script   string  `json:"script"`
	WaitMS   float64 `json:"wait_ms"`
}

func (b *BrowserTool) Execute(ctx context.Context, raw json.RawMessage) Result {
	var args browserArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return ResultErr("invalid args: " + err.Error())
	}

	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	tabCtx, tabCancel := chromedp.NewContext(allocCtx)
	defer tabCancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(tabCtx, 45*time.Second)
	defer timeoutCancel()

	wait := time.Duration(args.WaitMS) * time.Millisecond

	switch args.Action {
	case "navigate", "title":
		var title string
		tasks := chromedp.Tasks{chromedp.Navigate(args.URL)}
		if args.WaitMS > 0 {
			tasks = append(tasks, chromedp.Sleep(wait))
		}
		tasks = append(tasks, chromedp.Title(&title))
		if err := chromedp.Run(timeoutCtx, tasks); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(fmt.Sprintf("Page title: %s\nURL: %s", title, args.URL))

	case "screenshot":
		var buf []byte
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(args.URL),
			chromedp.Sleep(wait),
			chromedp.CaptureScreenshot(&buf),
		); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOkWithMeta(fmt.Sprintf("screenshot captured (%d bytes)", len(buf)), map[string]any{
			"size": len(buf),
		})

	case "content":
		if args.Selector == "" {
			args.Selector = "body"
		}
		var html string
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(args.URL),
			chromedp.Sleep(wait),
			chromedp.OuterHTML(args.Selector, &html),
		); err != nil {
			return ResultErr(err.Error())
		}
		if len(html) > 50000 {
			html = html[:50000] + "\n... (truncated)"
		}
		return ResultOk(html)

	case "text":
		if args.Selector == "" {
			args.Selector = "body"
		}
		var text string
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(args.URL),
			chromedp.Sleep(wait),
			chromedp.Text(args.Selector, &text),
		); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(strings.TrimSpace(text))

	case "click":
		if args.Selector == "" {
			return ResultErr("selector required for click action")
		}
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(args.URL),
			chromedp.Sleep(wait),
			chromedp.Click(args.Selector),
			chromedp.Sleep(500*time.Millisecond),
		); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk("clicked " + args.Selector)

	case "eval":
		if args.Script == "" {
			return ResultErr("script required for eval action")
		}
		var result string
		if err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(args.URL),
			chromedp.Sleep(wait),
			chromedp.Evaluate(args.Script, &result),
		); err != nil {
			return ResultErr(err.Error())
		}
		return ResultOk(result)

	default:
		return ResultErr("unknown action: " + args.Action)
	}
}
