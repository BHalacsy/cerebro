package anthropic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

type Tool struct {
	Name        string
	Description string
	InputSchema anthropic.ToolInputSchemaParam
	Run         func(ctx context.Context, input json.RawMessage) (string, error)
}

type Agent struct {
	client anthropic.Client
	model  anthropic.Model
	system string
	tools  []Tool
}

func NewAgent(model anthropic.Model, system string, tools []Tool) *Agent {
	client := anthropic.NewClient()
	return &Agent{client: client, model: model, system: system, tools: tools}
}

func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
	}

	toolParams := make([]anthropic.ToolUnionParam, len(a.tools))
	for i, t := range a.tools {
		toolParams[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: anthropic.String(t.Description),
				InputSchema: t.InputSchema,
			},
		}
	}

	params := anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: 4096,
		Messages:  messages,
	}
	if a.system != "" {
		params.System = []anthropic.TextBlockParam{{Text: a.system}}
	}
	if len(toolParams) > 0 {
		params.Tools = toolParams
	}

	for {
		msg, err := a.client.Messages.New(ctx, params)
		if err != nil {
			return "", fmt.Errorf("anthropic: %w", err)
		}

		messages = append(messages, msg.ToParam())
		params.Messages = messages

		if msg.StopReason == anthropic.StopReasonEndTurn {
			return extractText(msg), nil
		}

		results := a.handleToolUse(ctx, msg)
		if len(results) == 0 {
			return extractText(msg), nil
		}

		messages = append(messages, anthropic.NewUserMessage(results...))
		params.Messages = messages
	}
}

func (a *Agent) handleToolUse(ctx context.Context, msg *anthropic.Message) []anthropic.ContentBlockParamUnion {
	toolMap := make(map[string]Tool, len(a.tools))
	for _, t := range a.tools {
		toolMap[t.Name] = t
	}

	var results []anthropic.ContentBlockParamUnion
	for _, block := range msg.Content {
		tu, ok := block.AsAny().(anthropic.ToolUseBlock)
		if !ok {
			continue
		}

		tool, exists := toolMap[tu.Name]
		if !exists {
			results = append(results, anthropic.NewToolResultBlock(tu.ID, fmt.Sprintf("unknown tool: %s", tu.Name), true))
			continue
		}

		raw, _ := json.Marshal(tu.Input)
		output, err := tool.Run(ctx, raw)
		if err != nil {
			results = append(results, anthropic.NewToolResultBlock(tu.ID, err.Error(), true))
			continue
		}

		results = append(results, anthropic.NewToolResultBlock(tu.ID, output, false))
	}
	return results
}

func extractText(msg *anthropic.Message) string {
	for _, block := range msg.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			return tb.Text
		}
	}
	return ""
}
