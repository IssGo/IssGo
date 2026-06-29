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

func (w *WebTool) Name() string        { return "web" }
func (w *WebTool) Description() string { return "Make HTTP requests to fetch or send data." }

func (w *WebTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to request.",
			},
			"method": map[string]any{
				"type":        "string",
				"enum":        []string{"GET", "POST", "PUT", "DELETE"},
				"description": "HTTP method.",
			},
			"headers": map[string]any{
				"type":        "object",
				"description": "HTTP headers as key-value pairs.",
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Request body (JSON string or plain text).",
			},
		},
		"required": []string{"url"},
	}
}

type webArgs struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func (w *WebTool) Execute(ctx context.Context, raw json.RawMessage) Result {
	var args webArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("invalid args: %v", err)}
	}

	if args.Method == "" {
		args.Method = "GET"
	}

	client := resty.New().SetTimeout(30 * time.Second)

	req := client.R().
		SetContext(ctx).
		SetHeaders(args.Headers)

	if args.Body != "" {
		req.SetBody(args.Body)
	}

	var respErr error
	var resp *resty.Response

	switch strings.ToUpper(args.Method) {
	case "GET":
		resp, respErr = req.Get(args.URL)
	case "POST":
		resp, respErr = req.Post(args.URL)
	case "PUT":
		resp, respErr = req.Put(args.URL)
	case "DELETE":
		resp, respErr = req.Delete(args.URL)
	default:
		return Result{Success: false, Error: "unsupported method: " + args.Method}
	}

	if respErr != nil {
		return Result{Success: false, Error: respErr.Error()}
	}

	output := fmt.Sprintf("Status: %d\nBody: %s", resp.StatusCode(), string(resp.Body()))
	return Result{Success: resp.IsSuccess(), Output: output}
}
