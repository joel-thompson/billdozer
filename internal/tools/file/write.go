package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"agent/internal/schema"
	"agent/internal/tools"
)

// WriteFileInput represents the input parameters for writing to a file
type WriteFileInput struct {
	Path    string `json:"path" jsonschema_description:"File path to write content to (creates new file or overwrites existing)"`
	Content string `json:"content" jsonschema_description:"Text content to write (cannot be empty)"`
}

// WriteFileTool implements file writing functionality
type WriteFileTool struct{}

// Definition returns the tool definition for the write file tool
func (t WriteFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "write_file",
		Description: `Write content to a file, creating it if it doesn't exist or overwriting if it does.

- Always requires content to write
- Creates directories in the path if they don't exist
- Overwrites existing files completely
- Use create_file if you want to create an empty file
- Use edit_file if you want to modify specific parts of an existing file`,
		InputSchema: schema.GenerateSchema[WriteFileInput](),
	}
}

// Execute writes content to the specified file
func (t WriteFileTool) Execute(input json.RawMessage) (string, error) {
	var writeInput WriteFileInput
	err := json.Unmarshal(input, &writeInput)
	if err != nil {
		return "", err
	}

	if writeInput.Path == "" {
		return "", fmt.Errorf("path cannot be empty. Provide a file path to write to")
	}

	if writeInput.Content == "" {
		return "", fmt.Errorf("content cannot be empty. Use create_file to create empty files")
	}

	// Create directory if needed
	dir := filepath.Dir(writeInput.Path)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write the file (create or overwrite)
	err = os.WriteFile(writeInput.Path, []byte(writeInput.Content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote content to file %s", writeInput.Path), nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(WriteFileTool{})
}
