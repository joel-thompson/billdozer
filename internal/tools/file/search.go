package file

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"agent/internal/schema"
	"agent/internal/tools"
)

// Error constants
const (
	errMsgInvalidPattern = "invalid glob pattern '%s': %w"
)

type GlobSearchInput struct {
	Pattern string `json:"pattern" jsonschema:"required" jsonschema_description:"Glob pattern to search for. Examples: '*.go', 'test_*.txt', 'src/**/*.js'"`
	Path    string `json:"path,omitempty" jsonschema_description:"Base directory to search in (defaults to current directory if not provided)"`
}

// Validate implements input validation
func (g *GlobSearchInput) Validate() error {
	if g.Pattern == "" {
		return fmt.Errorf(errMsgMissingParam, "pattern")
	}
	return nil
}

// SearchResult represents the structured result of a glob search
type SearchResult struct {
	Pattern string   `json:"pattern"`
	Matches []string `json:"matches"`
	Count   int      `json:"count"`
}

// String returns a formatted string representation
func (sr *SearchResult) String() string {
	if sr.Count == 0 {
		return fmt.Sprintf("No files found matching pattern '%s'", sr.Pattern)
	}

	jsonResult, err := json.MarshalIndent(sr, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting results: %v", err)
	}

	return string(jsonResult)
}

type GlobSearchTool struct{}

func (t GlobSearchTool) Definition() tools.ToolDefinition {
	return tools.ToolDefinition{
		Name: "glob_search",
		Description: `Find files matching a glob pattern.
		
Usage Examples:
- {"pattern": "*.go"} // Find all .go files in current directory
- {"pattern": "test_*.txt", "path": "tests"} // Find test files in tests directory  

Supported Patterns:
- * matches any sequence of characters
- ? matches any single character
- [abc] matches any character in the set
- Use forward slashes for paths on all platforms

Note: Recursive patterns (**) support depends on Go's filepath.Glob implementation`,
		InputSchema: schema.GenerateSchema[GlobSearchInput](),
	}
}

func (t GlobSearchTool) Execute(ctx *tools.ToolContext, input json.RawMessage) (string, error) {
	searchInput, err := t.parseAndValidateInput(input)
	if err != nil {
		return "", err
	}

	result, err := t.performSearch(searchInput)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// Helper methods for better separation of concerns
func (t GlobSearchTool) parseAndValidateInput(input json.RawMessage) (*GlobSearchInput, error) {
	var searchInput GlobSearchInput
	if err := json.Unmarshal(input, &searchInput); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	if err := searchInput.Validate(); err != nil {
		return nil, err
	}

	return &searchInput, nil
}

func (t GlobSearchTool) buildSearchPattern(input *GlobSearchInput) string {
	if input.Path == "" {
		return input.Pattern
	}
	return filepath.Join(input.Path, input.Pattern)
}

func (t GlobSearchTool) performSearch(input *GlobSearchInput) (*SearchResult, error) {
	searchPattern := t.buildSearchPattern(input)

	matches, err := filepath.Glob(searchPattern)
	if err != nil {
		return nil, fmt.Errorf(errMsgInvalidPattern, searchPattern, err)
	}

	return &SearchResult{
		Pattern: searchPattern,
		Matches: matches,
		Count:   len(matches),
	}, nil
}

func init() {
	tools.DefaultRegistry.RegisterTool(GlobSearchTool{})
}
