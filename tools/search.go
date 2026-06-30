package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SearchTool struct{}

func (s *SearchTool) Name() string { return "search" }
func (s *SearchTool) Description() string {
	return "Search file contents with regex, glob, or literal matching."
}

func (s *SearchTool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern":     map[string]any{"type": "string", "description": "Search pattern (regex by default)."},
			"path":        map[string]any{"type": "string", "description": "Root directory to search."},
			"glob":        map[string]any{"type": "string", "description": "File glob filter (e.g., *.go)."},
			"literal":     map[string]any{"type": "boolean", "description": "Use literal string matching instead of regex."},
			"max_depth":   map[string]any{"type": "number", "description": "Max recursion depth (default 10)."},
			"max_results": map[string]any{"type": "number", "description": "Max matches to return (default 100)."},
		},
		"required": []string{"pattern"},
	}
}

type searchArgs struct {
	Pattern    string  `json:"pattern"`
	Path       string  `json:"path"`
	Glob       string  `json:"glob"`
	Literal    bool    `json:"literal"`
	MaxDepth   float64 `json:"max_depth"`
	MaxResults float64 `json:"max_results"`
}

type match struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

func (s *SearchTool) Execute(_ context.Context, raw json.RawMessage) Result {
	var args searchArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return ResultErr("invalid args: " + err.Error())
	}

	root := args.Path
	if root == "" {
		root = "."
	}

	maxDepth := 10
	if args.MaxDepth > 0 {
		maxDepth = int(args.MaxDepth)
	}

	maxResults := 100
	if args.MaxResults > 0 {
		maxResults = int(args.MaxResults)
	}

	var re *regexp.Regexp
	if args.Literal {
		re = regexp.MustCompile(regexp.QuoteMeta(args.Pattern))
	} else {
		var err error
		re, err = regexp.Compile(args.Pattern)
		if err != nil {
			return ResultErr("invalid regex: " + err.Error())
		}
	}

	var globRE *regexp.Regexp
	if args.Glob != "" {
		// Convert glob to regex
		g := regexp.QuoteMeta(args.Glob)
		g = strings.ReplaceAll(g, `\*`, ".*")
		g = strings.ReplaceAll(g, `\?`, ".")
		globRE = regexp.MustCompile("^" + g + "$")
	}

	var results []match
	count := 0

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || count >= maxResults {
			if count >= maxResults {
				return filepath.SkipAll
			}
			return nil
		}

		// Skip hidden dirs and common noise
		name := d.Name()
		if d.IsDir() && (name == ".git" || name == "node_modules" || name == "vendor" || name == "__pycache__" || (len(name) > 0 && name[0] == '.')) {
			return filepath.SkipDir
		}

		if d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			if depth := len(strings.Split(rel, string(os.PathSeparator))); depth > maxDepth && rel != "." {
				return filepath.SkipDir
			}
			return nil
		}

		if globRE != nil && !globRE.MatchString(name) {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		for i, line := range strings.Split(string(data), "\n") {
			if re.MatchString(line) {
				results = append(results, match{
					File: path,
					Line: i + 1,
					Text: strings.TrimSpace(line),
				})
				count++
				if count >= maxResults {
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return ResultErr(err.Error())
	}

	if len(results) == 0 {
		return ResultOk("no matches found")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d matches:\n\n", len(results)))
	for _, m := range results {
		sb.WriteString(fmt.Sprintf("%s:%d: %s\n", m.File, m.Line, m.Text))
	}
	if count >= maxResults {
		sb.WriteString("\n... (results truncated at limit)\n")
	}

	return ResultOk(sb.String())
}
