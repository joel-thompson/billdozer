package tools

import (
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
)

// ToolContext provides runtime context for tool execution
type ToolContext struct {
	GetUserInput UserInputFunction
}

// ToolDefinition represents a tool that can be called by the agent
type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(ctx *ToolContext, input json.RawMessage) (string, error)
}

// Tool interface that all tools must implement
type Tool interface {
	Definition() ToolDefinition
	Execute(ctx *ToolContext, input json.RawMessage) (string, error)
}

// ToolAdapter adapts a Tool interface to a ToolDefinition
func ToolAdapter(tool Tool) ToolDefinition {
	def := tool.Definition()
	def.Function = tool.Execute
	return def
}

// UserInputFunction is a function type for getting user input
type UserInputFunction func() (string, bool)
