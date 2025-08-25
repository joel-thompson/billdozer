package file

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"agent/internal/schema"
	"agent/internal/tools"
)

// Error message constants specific to delete operations
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
- Requires explicit user confirmation before deletion
- Validates file exists before deletion
- Clear error messages for missing files
- Does not delete directories (use with caution)`,
		InputSchema: schema.GenerateSchema[DeleteFileInput](),
	}
}

func (t DeleteFileTool) Execute(ctx *tools.ToolContext, input json.RawMessage) (string, error) {
	deleteInput, err := t.parseAndValidateInput(input)
	if err != nil {
		return "", err
	}

	if err := t.validateFileExists(deleteInput.Path); err != nil {
		return "", err
	}

	// Ask for user confirmation before deletion
	if !t.confirmDeletion(ctx, deleteInput.Path) {
		return "File deletion cancelled by user", nil
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

// confirmDeletion asks the user to confirm file deletion
func (t DeleteFileTool) confirmDeletion(ctx *tools.ToolContext, path string) bool {
	// Check if user input function is available
	if ctx.GetUserInput == nil {
		fmt.Printf("Warning: User input not available, proceeding with deletion\n")
		return true
	}

	// Ask for user confirmation
	fmt.Printf("⚠️ Billdozer wants to delete the file: \u001b[93m%s\u001b[0m\n", path)
	fmt.Printf("Do you want to proceed? (yes/y to confirm, anything else to cancel): ")

	response, ok := ctx.GetUserInput()
	if !ok {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "yes" || response == "y"
}

func init() {
	tools.DefaultRegistry.RegisterTool(DeleteFileTool{})
}
