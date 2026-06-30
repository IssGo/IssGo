package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type WebTool struct{}

func (w *WebTool) Name() string { return "web" }
func (w *WebTool) Description() string {
	return "Make HTTP requests with full support for headers, body, and auth."
}

func (w *WebTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url":         map[string]any{"type": "string", "description": "The URL to request."},
			"method":      map[string]any{"type": "string", "enum": []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}, "description": "HTTP method."},
			"headers":     map[string]any{"type": "object", "description": "HTTP headers."},
			"body":        map[string]any{"type": "string", "description": "Request body."},
			"auth":        map[string]any{"type": "object", "description": "Basic auth {username, password}."},
			"timeout_sec": map[string]any{"type": "number", "description": "Timeout in seconds (default 30)."},
		},
		"required": []string{"url"},
	}
}

type webArgs struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Auth    struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
	TimeoutSec float64 `json:"timeout_sec"`
}

func (w *WebTool) Execute(ctx context.Context, raw json.RawMessage) Result {
	var args webArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return ResultErr("invalid args: " + err.Error())
	}

	if args.Method == "" {
		args.Method = "GET"
	}

	timeout := 30 * time.Second
	if args.TimeoutSec > 0 && args.TimeoutSec <= 120 {
		timeout = time.Duration(args.TimeoutSec) * time.Second
	}

	client := resty.New().SetTimeout(timeout)
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(5))
	req := client.R().SetContext(ctx).SetHeaders(args.Headers)

	if args.Auth.Username != "" {
		req.SetBasicAuth(args.Auth.Username, args.Auth.Password)
	}

	if args.Body != "" {
		req.SetBody(args.Body)
	}

	var (
		resp *resty.Response
		err  error
	)

	switch strings.ToUpper(args.Method) {
	case "GET":
		resp, err = req.Get(args.URL)
	case "POST":
		resp, err = req.Post(args.URL)
	case "PUT":
		resp, err = req.Put(args.URL)
	case "DELETE":
		resp, err = req.Delete(args.URL)
	case "PATCH":
		resp, err = req.Patch(args.URL)
	case "HEAD":
		resp, err = req.Head(args.URL)
	case "OPTIONS":
		resp, err = req.Options(args.URL)
	default:
		return ResultErr("unsupported method: " + args.Method)
	}

	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}

	output := fmt.Sprintf("Status: %d %s\nHeaders:\n", resp.StatusCode(), resp.Status())
	for k, v := range resp.Header() {
		output += fmt.Sprintf("  %s: %s\n", k, strings.Join(v, ", "))
	}
	body := string(resp.Body())
	if len(body) > 20000 {
		body = body[:20000] + "\n... (truncated)"
	}
	output += "\nBody:\n" + body

	return Result{Success: resp.IsSuccess(), Output: output}
}
