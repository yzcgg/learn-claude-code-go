package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func newCreateFileTool(dir string) Tool {
	return Tool{
		Definition: anthropic.ToolParam{
			Name:        "create_file",
			Description: anthropic.String("Create a new empty file. Fails if the path already exists or the parent directory does not exist."),
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
				return "", fmt.Errorf("create_file: path 必须是字符串")
			}
			return RunCreateFile(path, dir)
		},
	}
}

func RunCreateFile(path string, dir string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path can not be empty")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}

	// O_EXCL prevents an existing file from being overwritten.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return fmt.Sprintf("created %s", path), nil
}
