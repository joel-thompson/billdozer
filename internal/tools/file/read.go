package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"agent/internal/schema"
	"agent/internal/tools"
)

// Constants for validation
const (
	minLineNumber        = 1
	minLimitValue        = 1
	errMsgInvalidOffset  = "offset must be >= %d (line numbers are 1-based)"
	errMsgInvalidLimit   = "limit must be >= %d"
	errMsgOffsetTooLarge = "offset %d exceeds file length (%d lines)"
)

type ReadFileInput struct {
	Path   string `json:"path" jsonschema:"required" jsonschema_description:"The relative path of a file in the working directory."`
	Offset *int   `json:"offset,omitempty" jsonschema_description:"Starting line number (1-based). If provided, only reads from this line onwards."`
	Limit  *int   `json:"limit,omitempty" jsonschema_description:"Maximum number of lines to read. If provided with offset, reads this many lines from the offset."`
}

// Validate implements input validation
func (r *ReadFileInput) Validate() error {
	if r.Path == "" {
		return fmt.Errorf(errMsgMissingParam, "path")
	}

	if r.Offset != nil && *r.Offset < minLineNumber {
		return fmt.Errorf(errMsgInvalidOffset, minLineNumber)
	}

	if r.Limit != nil && *r.Limit < minLimitValue {
		return fmt.Errorf(errMsgInvalidLimit, minLimitValue)
	}

	return nil
}

// lineRange represents a range of lines to read
type lineRange struct {
	start, end int
}

// calculateRange computes the line range based on offset and limit
func (lr *lineRange) calculateRange(totalLines int, offset, limit *int) error {
	lr.start = 0
	lr.end = totalLines

	if offset != nil {
		lr.start = *offset - 1 // Convert to 0-based
		if lr.start >= totalLines {
			return fmt.Errorf(errMsgOffsetTooLarge, *offset, totalLines)
		}
	}

	if limit != nil {
		requestedEnd := lr.start + *limit
		if requestedEnd < lr.end {
			lr.end = requestedEnd
		}
	}

	return nil
}

type ReadFileTool struct{}

func (t ReadFileTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "read_file",
		Description: `Read the contents of a file with optional line range support.

Usage Examples:
- {"path": "main.go"} // Read entire file
- {"path": "config.yml", "offset": 10} // Read from line 10 to end
- {"path": "data.txt", "offset": 5, "limit": 20} // Read lines 5-24 (20 lines starting from line 5)

Parameters:
- path: File path to read (required)
- offset: Starting line number (1-based, optional)
- limit: Maximum number of lines to read (optional)

Note: Line numbers are 1-based. Use this when you want to see what's inside a file.
Do not use this with directory names.`,
		InputSchema: schema.GenerateSchema[ReadFileInput](),
	}
}

func (t ReadFileTool) Execute(input json.RawMessage) (string, error) {
	readInput, err := t.parseAndValidateInput(input)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(readInput.Path)
	if err != nil {
		return "", err
	}

	// If no offset/limit specified, return full content (backward compatibility)
	if readInput.Offset == nil && readInput.Limit == nil {
		return string(content), nil
	}

	return t.extractLines(string(content), readInput)
}

// Helper methods for better separation of concerns
func (t ReadFileTool) parseAndValidateInput(input json.RawMessage) (*ReadFileInput, error) {
	var readInput ReadFileInput
	if err := json.Unmarshal(input, &readInput); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	if err := readInput.Validate(); err != nil {
		return nil, err
	}

	return &readInput, nil
}

// splitLines handles cross-platform line endings properly
func (t ReadFileTool) splitLines(content string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Handle empty files or files with no final newline
	if len(lines) == 0 && content != "" {
		lines = []string{content}
	}

	return lines
}

func (t ReadFileTool) extractLines(content string, input *ReadFileInput) (string, error) {
	lines := t.splitLines(content)
	totalLines := len(lines)

	var lr lineRange
	if err := lr.calculateRange(totalLines, input.Offset, input.Limit); err != nil {
		return "", err
	}

	selectedLines := lines[lr.start:lr.end]
	return strings.Join(selectedLines, "\n"), nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(ReadFileTool{})
}
