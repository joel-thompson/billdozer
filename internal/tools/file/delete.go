package file

import (
	"encoding/json"
	"fmt"
	"os"

	"agent/internal/schema"
	"agent/internal/tools"
)

// Error message constants (shared across all tools)
const (
	errMsgFileNotFound = "file not found: %s"
	errMsgIsDirectory  = "%s is a directory, not a file. Use directory tools for directory operations"
)

type DeleteFileInput struct {
	Path string `json:"path" jsonschema:"required" jsonschema_description:"File path to delete. Must be an existing file."`
}

// Validate implements input validation
func (d *DeleteFileInput) Validate() error {
	if d.Path == "" {
		return fmt.Errorf(errMsgMissingParam, "path")
	}
	return nil
}

type DeleteFileTool struct{}

func (t DeleteFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "delete_file",
		Description: `Delete a file from the filesystem.
		
Requirements:
- File must exist (will fail if file doesn't exist)
- Only deletes files, not directories  
- Cannot be undone

Safety:
- Validates file exists before deletion
- Clear error messages for missing files
- Does not delete directories (use with caution)`,
		InputSchema: schema.GenerateSchema[DeleteFileInput](),
	}
}

func (t DeleteFileTool) Execute(input json.RawMessage) (string, error) {
	deleteInput, err := t.parseAndValidateInput(input)
	if err != nil {
		return "", err
	}

	if err := t.validateFileExists(deleteInput.Path); err != nil {
		return "", err
	}

	if err := os.Remove(deleteInput.Path); err != nil {
		return "", fmt.Errorf(errMsgOperationFailed, "delete file", err)
	}

	return fmt.Sprintf("Successfully deleted file %s", deleteInput.Path), nil
}

// Helper methods for better separation of concerns
func (t DeleteFileTool) parseAndValidateInput(input json.RawMessage) (*DeleteFileInput, error) {
	var deleteInput DeleteFileInput
	if err := json.Unmarshal(input, &deleteInput); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	if err := deleteInput.Validate(); err != nil {
		return nil, err
	}

	return &deleteInput, nil
}

func (t DeleteFileTool) validateFileExists(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf(errMsgFileNotFound, path)
	}
	if err != nil {
		return fmt.Errorf(errMsgOperationFailed, "check file", err)
	}

	if info.IsDir() {
		return fmt.Errorf(errMsgIsDirectory, path)
	}

	return nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(DeleteFileTool{})
}
