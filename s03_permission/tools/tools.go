package tools

import "github.com/anthropics/anthropic-sdk-go"

type ToolHandler func(input map[string]any) (string, error)

// Tool keeps a tool's API definition and implementation together.
type Tool struct {
	Definition anthropic.ToolParam
	Handler    ToolHandler
}

// NewTools is the single registration point for tools, in API presentation order.
func NewTools(dir string) []Tool {
	return []Tool{
		// newBashTool(dir),
		newReadFileTool(dir),
		newWriteFileTool(dir),
		newCreateFileTool(dir),
	}
}
