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
		Description: `Make edits to a text file.

Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other.

If the file specified with path doesn't exist, it will be created.`,
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

	if editFileInput.OldStr == editFileInput.NewStr {
		return "", fmt.Errorf("old_str and new_str must be different")
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) && editFileInput.OldStr == "" {
			return t.createNewFile(editFileInput.Path, editFileInput.NewStr)
		}
		return "", err
	}

	oldContent := string(content)

	// Special case: if old_str is empty and file content is empty, treat as "add content to empty file"
	if editFileInput.OldStr == "" && oldContent == "" {
		newContent := editFileInput.NewStr
		err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
		if err != nil {
			return "", err
		}
		return "OK", nil
	}

	// Normal replacement
	newContent := strings.Replace(oldContent, editFileInput.OldStr, editFileInput.NewStr, -1)

	if oldContent == newContent && editFileInput.OldStr != "" {
		return "", fmt.Errorf("old_str not found in file")
	}

	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return "", err
	}

	return "OK", nil
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
