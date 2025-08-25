package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"agent/internal/schema"
	"agent/internal/tools"
)

// CreateFileInput represents the input parameters for creating a file
type CreateFileInput struct {
	Path string `json:"path" jsonschema_description:"File path for the new empty file (directories will be created as needed)"`
}

// CreateFileTool implements file creation functionality
type CreateFileTool struct{}

// Definition returns the tool definition for the create file tool
func (t CreateFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "create_file",
		Description: `Create an empty file (like 'touch' command).

- Creates an empty file at the specified path
- Fails if the file already exists  
- Creates directories in the path if they don't exist
- Use write_file if you want to create a file with content`,
		InputSchema: schema.GenerateSchema[CreateFileInput](),
	}
}

// Execute creates a new empty file
func (t CreateFileTool) Execute(input json.RawMessage) (string, error) {
	var createInput CreateFileInput
	err := json.Unmarshal(input, &createInput)
	if err != nil {
		return "", err
	}

	if createInput.Path == "" {
		return "", fmt.Errorf("path cannot be empty. Provide a file path to create")
	}

	// Create directory if needed
	dir := filepath.Dir(createInput.Path)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Atomically create the file (fails if it already exists)
	file, err := os.OpenFile(createInput.Path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("file %s already exists. Use write_file to overwrite existing files", createInput.Path)
		}
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	file.Close()

	return fmt.Sprintf("Successfully created empty file %s", createInput.Path), nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(CreateFileTool{})
}
