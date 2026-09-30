package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func newReadFileTool(dir string) Tool {
	return Tool{
		Definition: anthropic.ToolParam{
			Name:        "read_file",
			Description: anthropic.String("read file"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"path": map[string]string{"type": "string"},
				},
				Required: []string{"path"},
			},
		},
		Handler: func(input map[string]any) (string, error) {
			path, ok := input["path"].(string)
			if !ok {
				return "", fmt.Errorf("read_file: path 必须是字符串")
			}
			return RunReadFile(path, dir)
		},
	}
}

// tool 02: read file
func RunReadFile(path string, dir string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path can not be empty")
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
