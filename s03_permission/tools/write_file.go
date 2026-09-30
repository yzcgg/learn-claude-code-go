package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func newWriteFileTool(dir string) Tool {
	return Tool{
		Definition: anthropic.ToolParam{
			Name:        "write_file",
			Description: anthropic.String("Append content to an existing file. Fails if the file does not exist. Does not add a newline automatically."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"path":    map[string]string{"type": "string"},
					"content": map[string]string{"type": "string"},
				},
				Required: []string{"path", "content"},
			},
		},
		Handler: func(input map[string]any) (string, error) {
			path, ok := input["path"].(string)
			if !ok {
				return "", fmt.Errorf("write_file: path 必须是字符串")
			}
			content, ok := input["content"].(string)
			if !ok {
				return "", fmt.Errorf("write_file: content 必须是字符串")
			}
			return RunWriteFile(path, content, dir)
		},
	}
}

func RunWriteFile(path string, content string, dir string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path can not be empty")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}

	// Do not use O_CREATE: appending requires an existing file.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return "", err
	}
	n, writeErr := file.WriteString(content)
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return fmt.Sprintf("appended %d bytes to %s", n, path), nil
}
