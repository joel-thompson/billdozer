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
	Path    string `json:"path" jsonschema:"required" jsonschema_description:"File path to write content to (creates new file or overwrites existing). Example: 'src/main.go' or 'docs/readme.md'"`
	Content string `json:"content" jsonschema:"required" jsonschema_description:"Text content to write to the file. This parameter is REQUIRED - you must provide the actual content you want written to the file. Cannot be empty."`
}

// WriteFileTool implements file writing functionality
type WriteFileTool struct{}

// Definition returns the tool definition for the write file tool
func (t WriteFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "write_file",
		Description: `Write content to a file, creating it if it doesn't exist or overwriting if it does.

IMPORTANT: This tool requires BOTH a file path AND content to write.

Usage Examples:
- {"path": "config.yml", "content": "version: 1.0\nname: myapp"}
- {"path": "src/utils.go", "content": "package main\n\nfunc main() {\n  // code here\n}"}
- {"path": "docs/readme.md", "content": "# My Project\n\nThis is a readme file."}

Requirements:
- path: Must provide a file path (creates directories if needed)
- content: Must provide actual text content (cannot be empty)

Behavior:
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
		return "", fmt.Errorf("path parameter is required. Please provide a file path like 'src/main.go' or 'docs/readme.md'. Example: {\"path\": \"myfile.txt\", \"content\": \"your content here\"}")
	}

	if writeInput.Content == "" {
		return "", fmt.Errorf("content parameter is required and cannot be empty. You must provide the actual text content to write to the file. Example: {\"path\": \"myfile.txt\", \"content\": \"your content here\"}. If you want to create an empty file, use the create_file tool instead")
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
