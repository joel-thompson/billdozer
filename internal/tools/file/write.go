package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"agent/internal/schema"
	"agent/internal/tools"
)

// Constants for better maintainability
const (
	defaultFilePermissions = 0644
	defaultDirPermissions  = 0755
	errMsgMissingParam     = "parameter %q is required"
	errMsgOperationFailed  = "failed to %s: %w"
)

// WriteFileInput with validation interface
type WriteFileInput struct {
	Path    string `json:"path" jsonschema:"required" jsonschema_description:"File path to write to (creates new file or overwrites existing). Examples: 'src/main.go', 'docs/readme.md'"`
	Content string `json:"content" jsonschema_description:"Content to write to file. Leave empty to create an empty file (like touch command). Cannot be null, but can be empty string."`
}

// Validate implements input validation
func (w *WriteFileInput) Validate() error {
	if w.Path == "" {
		return fmt.Errorf(errMsgMissingParam, "path")
	}
	return nil
}

type WriteFileTool struct{}

func (t WriteFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "write", // Changed from "write_file"
		Description: `Write content to a file OR create an empty file.
		
IMPORTANT: This unified tool replaces both create_file and write_file.

Usage Examples:
- {"path": "empty.txt", "content": ""} // Creates empty file
- {"path": "config.yml", "content": "version: 1.0\nname: myapp"} // Creates file with content
- {"path": "new/dir/file.txt", "content": "test"} // Creates directories as needed

Behavior:
- Empty content creates empty file (like touch command)
- With content creates file with that content  
- Always overwrites existing files
- Creates parent directories automatically`,
		InputSchema: schema.GenerateSchema[WriteFileInput](),
	}
}

func (t WriteFileTool) Execute(input json.RawMessage) (string, error) {
	writeInput, err := t.parseAndValidateInput(input)
	if err != nil {
		return "", err
	}

	if err := t.ensureDirectoryExists(writeInput.Path); err != nil {
		return "", err
	}

	return t.writeFile(writeInput.Path, writeInput.Content)
}

// Helper methods for better separation of concerns
func (t WriteFileTool) parseAndValidateInput(input json.RawMessage) (*WriteFileInput, error) {
	var writeInput WriteFileInput
	if err := json.Unmarshal(input, &writeInput); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	if err := writeInput.Validate(); err != nil {
		return nil, err
	}

	return &writeInput, nil
}

func (t WriteFileTool) ensureDirectoryExists(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, defaultDirPermissions); err != nil {
			return fmt.Errorf(errMsgOperationFailed, "create directory", err)
		}
	}
	return nil
}

func (t WriteFileTool) writeFile(path, content string) (string, error) {
	if content == "" {
		// Create empty file (replaces create_file functionality)
		file, err := os.Create(path)
		if err != nil {
			return "", fmt.Errorf(errMsgOperationFailed, "create empty file", err)
		}
		file.Close()
		return fmt.Sprintf("Created empty file %s", path), nil
	}

	// Write content to file
	if err := os.WriteFile(path, []byte(content), defaultFilePermissions); err != nil {
		return "", fmt.Errorf(errMsgOperationFailed, "write file", err)
	}

	return fmt.Sprintf("Successfully wrote content to file %s", path), nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(WriteFileTool{}) // Only this line remains - CreateFileTool removed
}
