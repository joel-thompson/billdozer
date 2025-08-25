package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"agent/internal/schema"
	"agent/internal/tools"
)

// EditFileInput represents the input parameters for editing a file
type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

// EditFileTool implements the file editing functionality
type EditFileTool struct{}

// Definition returns the tool definition for the edit file tool
func (t EditFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "edit_file",
		Description: `Make edits to a text file or create a new file.

For editing existing files:
- Replaces 'old_str' with 'new_str' in the given file
- 'old_str' and 'new_str' MUST be different from each other
- 'old_str' must exist exactly once in the file

For creating new files:
- Use an empty 'old_str' or common placeholders like: FILE_DOES_NOT_EXIST, NEW_FILE, PLACEHOLDER, empty, CREATE_FILE
- 'new_str' will be the content of the new file
- Directory structure will be created if needed`,
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
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check if we're trying to create a new file
	if t.isFileCreationRequest(editFileInput.OldStr) {
		// For file creation, allow same old_str and new_str since old_str is just a placeholder
		return t.createNewFile(editFileInput.Path, editFileInput.NewStr)
	}

	// For normal edits, old_str and new_str must be different
	if editFileInput.OldStr == editFileInput.NewStr {
		return "", fmt.Errorf("old_str and new_str must be different")
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist and we're not in file creation mode
			return "", fmt.Errorf("file does not exist. To create a new file, use an empty old_str or one of these placeholders: FILE_DOES_NOT_EXIST, NEW_FILE, PLACEHOLDER")
		}
		return "", err
	}

	oldContent := string(content)

	// Special case: if old_str is empty and we want to add content to empty file
	if editFileInput.OldStr == "" {
		if oldContent == "" {
			// Add content to empty file
			err = os.WriteFile(editFileInput.Path, []byte(editFileInput.NewStr), 0644)
			if err != nil {
				return "", err
			}
			return "OK", nil
		} else {
			return "", fmt.Errorf("old_str cannot be empty when file has existing content. Please specify the text to replace")
		}
	}

	// Normal replacement
	newContent := strings.Replace(oldContent, editFileInput.OldStr, editFileInput.NewStr, -1)

	if oldContent == newContent {
		return "", fmt.Errorf("old_str '%s' not found in file", editFileInput.OldStr)
	}

	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

// isFileCreationRequest checks if the old_str indicates a request to create a new file
func (t EditFileTool) isFileCreationRequest(oldStr string) bool {
	// Common placeholder values that indicate file creation intent
	creationPlaceholders := []string{
		"",                    // Empty string (original behavior)
		"FILE_DOES_NOT_EXIST", // Explicit placeholder
		"NEW_FILE",            // Clear intent
		"PLACEHOLDER",         // Generic placeholder
		"empty",               // Common natural language
		"CREATE_FILE",         // Another explicit option
	}

	for _, placeholder := range creationPlaceholders {
		if oldStr == placeholder {
			return true
		}
	}
	return false
}

// createNewFile creates a new file with the given content
func (t EditFileTool) createNewFile(filePath, content string) (string, error) {
	dir := path.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(EditFileTool{})
}
