package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/yzc/s03/config"
	"github.com/yzc/s03/permission"
	"github.com/yzc/s03/tools"
)

var (
	WORK_DIR   = config.LocalConfig.LccWorkDir + "/s02_tool_use"
	testPrompt = []string{
		"给我在 /etc 下创建 d.txt, 写内容为 '我不是鸡，我是鸽鸽'",
	}
)

func main() {
	log.Printf("WORK_DIR = %s\n", WORK_DIR)
	client := anthropic.NewClient(option.WithBaseURL(config.LocalConfig.BaseUrl), option.WithAPIKey(config.LocalConfig.APIKey))
	agentLoop(context.Background(), client, []anthropic.MessageParam{
		anthropic.NewUserMessage(
			anthropic.NewTextBlock(testPrompt[0]),
		),
	})
}

func getToolList(registeredTools []tools.Tool) []anthropic.ToolUnionParam {
	definitions := make([]anthropic.ToolUnionParam, 0, len(registeredTools))
	for i := range registeredTools {
		definitions = append(definitions, anthropic.ToolUnionParam{
			OfTool: &registeredTools[i].Definition,
		})
	}
	return definitions
}

func getSystemPrompt(path string) []anthropic.TextBlockParam {
	return []anthropic.TextBlockParam{
		{Text: fmt.Sprintf("You are a coding agent at %s. Use bash to solve tasks. Act, don't explain.", path)},
	}
}

func agentLoop(ctx context.Context, client anthropic.Client, messages []anthropic.MessageParam) {
	registeredTools := tools.NewTools(WORK_DIR)
	toolList := getToolList(registeredTools)
	toolHandlers := make(map[string]tools.ToolHandler, len(registeredTools))
	for _, tool := range registeredTools {
		toolHandlers[tool.Definition.Name] = tool.Handler
	}

	for {
		resp, err := client.Messages.New(
			ctx,
			anthropic.MessageNewParams{
				Model:     config.LocalConfig.Model,
				MaxTokens: 8000,
				Messages:  messages,
				Tools:     toolList,
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
		for _, block := range resp.Content {
			if toolUse, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
				log.Printf("\033[36m> %s\033[0m", toolUse.Name)

				if !permission.CheckPermission(toolUse, WORK_DIR) {
					results = append(results, anthropic.NewToolResultBlock(
						toolUse.ID,
						"Permission denied. The user denied this action. Do not retry it or attempt an alternative to bypass the denial.",
						true,
					))
					continue
				}

				var input map[string]any
				err := json.Unmarshal(toolUse.Input, &input)
				if err != nil {
					log.Fatalf("parse block input failed, err = %v\n", err)
					return
				}
				handler, ok := toolHandlers[toolUse.Name]
				if !ok {
					log.Fatalf("unknown tool: %v\n", toolUse.Name)
					return
				}
				output, err := handler(input)
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
