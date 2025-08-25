package file

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"agent/internal/schema"
	"agent/internal/tools"
)

// EditFileInput represents the input parameters for editing a file
type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"Path to existing file to edit"`
	OldStr string `json:"old_str" jsonschema_description:"Exact text to find and replace (must appear exactly once)"`
	NewStr string `json:"new_str" jsonschema_description:"Replacement text (must differ from old_str)"`
}

// EditFileTool implements the file editing functionality
type EditFileTool struct{}

// Definition returns the tool definition for the edit file tool
func (t EditFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "edit_file",
		Description: `Edit an existing text file by replacing text.

- File must already exist (use create_file or write_file for new files)
- Replaces 'old_str' with 'new_str' in the given file
- 'old_str' must exist exactly once in the file
- 'old_str' and 'new_str' must be different`,
		InputSchema: schema.GenerateSchema[EditFileInput](),
	}
}

// Execute performs the file editing operation
func (t EditFileTool) Execute(input json.RawMessage) (string, error) {
	var editFileInput EditFileInput
	err := json.Unmarshal(input, &editFileInput)
	if err != nil {
		return "", err
	}

	if editFileInput.Path == "" {
		return "", fmt.Errorf("path cannot be empty. Provide a file path to edit")
	}

	if editFileInput.OldStr == "" {
		return "", fmt.Errorf("old_str cannot be empty. Use create_file or write_file for new files")
	}

	if editFileInput.OldStr == editFileInput.NewStr {
		return "", fmt.Errorf("old_str and new_str must be different")
	}

	// Read existing file
	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file does not exist. Use create_file or write_file for new files")
		}
		return "", err
	}

	// Check if file is binary to prevent corruption
	if isBinary(content) {
		return "", fmt.Errorf("cannot edit binary file %s. Use write_file to replace binary files entirely", editFileInput.Path)
	}

	oldContent := string(content)

	// Check that old_str exists exactly once
	count := strings.Count(oldContent, editFileInput.OldStr)
	if count == 0 {
		return "", fmt.Errorf("old_str '%s' not found in file", editFileInput.OldStr)
	}
	if count > 1 {
		return "", fmt.Errorf("old_str '%s' found %d times in file, must exist exactly once", editFileInput.OldStr, count)
	}

	// Perform replacement
	newContent := strings.Replace(oldContent, editFileInput.OldStr, editFileInput.NewStr, 1)

	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully edited file %s", editFileInput.Path), nil
}

// isBinary detects if a file contains binary data to prevent text editing corruption
func isBinary(data []byte) bool {
	// Simple heuristic: if file contains null bytes in first 512 bytes, it's likely binary
	checkLen := 512
	if len(data) < checkLen {
		checkLen = len(data)
	}
	return bytes.IndexByte(data[:checkLen], 0) != -1
}

func init() {
	tools.DefaultRegistry.RegisterTool(EditFileTool{})
}
