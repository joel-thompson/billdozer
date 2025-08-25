package file

import (
	"encoding/json"
	"os"
	"path/filepath"

	"agent/internal/schema"
	"agent/internal/tools"
)

// ListFilesInput represents the input parameters for listing files
type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

// ListFilesTool implements the file listing functionality
type ListFilesTool struct{}

// Definition returns the tool definition for the list files tool
func (t ListFilesTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name:        "list_files",
		Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
		InputSchema: schema.GenerateSchema[ListFilesInput](),
	}
}

// Execute performs the file listing operation
func (t ListFilesTool) Execute(ctx *tools.ToolContext, input json.RawMessage) (string, error) {
	var listFilesInput ListFilesInput
	err := json.Unmarshal(input, &listFilesInput)
	if err != nil {
		return "", err
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(ListFilesTool{})
}
