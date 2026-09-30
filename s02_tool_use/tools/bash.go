package tools

import (
	"fmt"
	"log"
	"os/exec"
	"slices"

	"github.com/anthropics/anthropic-sdk-go"
)

func newBashTool(dir string) Tool {
	return Tool{
		Definition: anthropic.ToolParam{
			Name:        "bash",
			Description: anthropic.String("run a bash command"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"command": map[string]string{"type": "string"},
				},
				Required: []string{"command"},
			},
		},
		Handler: func(input map[string]any) (string, error) {
			command, ok := input["command"].(string)
			if !ok {
				return "", fmt.Errorf("bash: command 必须是字符串")
			}
			return RunBash(command, dir)
		},
	}
}

func RunBash(cmdStr string, dir string) (string, error) {
	log.Println(cmdStr)
	dangerous := []string{"rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"}
	if slices.Contains(dangerous, cmdStr) {
		return "", nil
	}
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}
