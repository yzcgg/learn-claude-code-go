package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

var (
	localConfig Config
	WORK_DIR    = "/Users/yangzhicheng/Desktop/yzc/codes/learn-claude-code-go"
)

func main() {
	localConfig = LoadConfig()
	client := anthropic.NewClient(
		option.WithBaseURL(localConfig.BaseUrl),
		option.WithAPIKey(localConfig.APIKey),
	)
	agentLoop(context.Background(), client, []anthropic.MessageParam{
		anthropic.NewUserMessage(
			anthropic.NewTextBlock("创建一个名为hello.go的文件，输出Hello World Go。"),
		),
	})
}

func getToolList() []anthropic.ToolUnionParam {
	tools := make([]anthropic.ToolUnionParam, 0)
	// add bash tool
	tools = append(tools, anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        "bash",
			Description: anthropic.String("run a bash command"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Type: "object",
				Properties: map[string]any{
					"command": map[string]string{
						"type": "string",
					},
				},
				Required: []string{"command"},
			},
		},
	})
	return tools
}

func getSystemPrompt(path string) []anthropic.TextBlockParam {
	return []anthropic.TextBlockParam{
		{Text: fmt.Sprintf("You are a coding agent at %s. Use bash to solve tasks. Act, don't explain.", path)},
	}
}

func agentLoop(ctx context.Context, client anthropic.Client, messages []anthropic.MessageParam) {
	for {
		resp, err := client.Messages.New(
			ctx,
			anthropic.MessageNewParams{
				Model:     localConfig.Model,
				MaxTokens: 8000,
				Messages:  messages,
				Tools:     getToolList(),
				System:    getSystemPrompt(WORK_DIR),
			},
		)
		if err != nil {
			log.Printf("LLM call error, err = %v\n", err)
			return
		}
		messages = append(messages, resp.ToParam())

		// log.Printf("resp = %+v\n", resp)
		if resp.StopReason != "tool_use" {
			log.Printf("end loop.\n")
			return
		}

		results := make([]anthropic.ContentBlockParamUnion, 0)
		for idx, block := range resp.Content {
			log.Printf("id = %v, block = %+v", idx, block)
			if toolUse, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
				var input map[string]any
				err := json.Unmarshal(toolUse.Input, &input)
				if err != nil {
					log.Fatalf("parse block input failed, err = %v\n", err)
					return
				}
				output, err := runBash(input["command"].(string))
				if err != nil {
					log.Fatalf("run bash failed, err = %v\n", err)
					return
				}
				log.Printf("output = %v\n", output)

				results = append(results, anthropic.NewToolResultBlock(
					toolUse.ID,
					output,
					false,
				))
			}
		}
		messages = append(messages, anthropic.NewUserMessage(
			results...,
		))
	}
}

func runBash(cmdStr string) (string, error) {
	log.Println(cmdStr)
	dangerous := []string{"rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"}
	if slices.Contains(dangerous, cmdStr) {
		return "", nil
	}
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = WORK_DIR
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func parseCmd(cmd string) []string {
	return strings.Split(cmd, " ")
}

type Config struct {
	APIKey  string
	BaseUrl string
	Model   string
}

func LoadConfig() Config {
	return Config{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseUrl: os.Getenv("OPENAI_API_BASE_URL"),
		Model:   os.Getenv("OPENAI_API_MODEL"),
	}
}
