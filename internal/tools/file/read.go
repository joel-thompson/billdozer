package file

import (
	"encoding/json"
	"os"

	"agent/internal/schema"
	"agent/internal/tools"
)

// ReadFileInput represents the input parameters for reading a file
type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

// ReadFileTool implements the file reading functionality
type ReadFileTool struct{}

// Definition returns the tool definition for the read file tool
func (t ReadFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name:        "read_file",
		Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
		InputSchema: schema.GenerateSchema[ReadFileInput](),
	}
}

// Execute performs the file reading operation
func (t ReadFileTool) Execute(input json.RawMessage) (string, error) {
	var readFileInput ReadFileInput
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(ReadFileTool{})
}
