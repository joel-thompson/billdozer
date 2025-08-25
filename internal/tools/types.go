package tools

import (
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
)

// ToolDefinition represents a tool that can be called by the agent
type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

// Tool interface that all tools must implement
type Tool interface {
	Definition() ToolDefinition
	Execute(input json.RawMessage) (string, error)
}

// ToolAdapter adapts a Tool interface to a ToolDefinition
func ToolAdapter(tool Tool) ToolDefinition {
	def := tool.Definition()
	def.Function = tool.Execute
	return def
}
