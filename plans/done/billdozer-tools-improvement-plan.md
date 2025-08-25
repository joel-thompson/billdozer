# Billdozer Tools Improvement Plan - Linear Implementation Guide

## Executive Summary

Replace problematic `create_file` and `write_file` tools with a single unified `write` tool (like Cursor), then add missing tools. This guide provides step-by-step implementation instructions.

## Step-by-Step Implementation

### STEP 1: Analyze Current State
**What we're replacing:**
- `internal/tools/file/create.go` - Creates empty files only, fails if exists
- `internal/tools/file/write.go` - Requires content, can't create empty files

**Root problem:** AI agents struggle to choose between create vs write, causing errors and decision fatigue.

### STEP 2: Define New Write Tool Specification

**New unified tool behavior:**
- **Empty content parameter** → Creates empty file (replaces create_file)
- **Content provided** → Creates file with content (replaces write_file)  
- **Always overwrites** → No "file exists" errors
- **Auto-creates directories** → Like current tools

```go
type WriteInput struct {
    Path    string `json:"path" jsonschema:"required" jsonschema_description:"File path to write to (creates new file or overwrites existing). Examples: 'src/main.go', 'docs/readme.md'"`
    Content string `json:"content" jsonschema_description:"Content to write to file. Leave empty to create an empty file (like touch command). Cannot be null, but can be empty string."`
}
```

### STEP 3: Implement New Write Tool

**3.1 - Delete old tools:**
- Delete `internal/tools/file/create.go` (no longer needed)

**3.2 - Replace write.go with new implementation:**
- Keep same file: `internal/tools/file/write.go`
- Replace entire contents with new unified tool

**Implementation Details:**
```go
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
```

**Key Improvements:**
- **Extracted helper methods** for better separation of concerns
- **Input validation interface** with `Validate()` method
- **Consistent error messages** using constants
- **Better error wrapping** with more context
- **Constants for permissions** instead of magic numbers

**3.3 - Update tool registration:**
- Remove `CreateFileTool` from init() function  
- Keep only `WriteFileTool` registration
- **Change tool name to `"write"`** to match Cursor exactly

**Implementation Details:**
```go
// Update Definition() method in WriteFileTool
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

// Update init() function - remove CreateFileTool registration
func init() {
    tools.DefaultRegistry.RegisterTool(WriteFileTool{}) // Only this line remains
}
```

### STEP 4: Test New Write Tool

**4.1 - Basic functionality test:**
```bash
go run main.go
# Test in AI chat:
# "create an empty file at test.txt" (should use empty content)
# "write hello world to test.txt" (should overwrite with content)
```

**4.2 - Verify both use cases work:**
- Empty file creation: `{"path": "empty.txt", "content": ""}`
- File with content: `{"path": "content.txt", "content": "Hello World"}`
- Directory creation: `{"path": "new/dir/file.txt", "content": "test"}`
- File overwriting: Write to existing file twice

**4.3 - Update documentation:**
- Update `README.md` to reflect single `write` tool
- Remove references to `create_file` and `write_file`
- Add examples of empty vs content usage

### STEP 5: Add Missing Tools (In Priority Order)

### STEP 5A: Add File Deletion Tool

**5A.1 - Create `internal/tools/file/delete.go`:**

**Implementation Details:**
```go
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
    errMsgFileNotFound   = "file not found: %s"
    errMsgIsDirectory    = "%s is a directory, not a file. Use directory tools for directory operations"
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
```

**Key Improvements:**
- **Extracted helper methods** for validation and parsing
- **Input validation interface** with `Validate()` method  
- **Consistent error constants** shared across tools
- **Better separation of concerns** in Execute method

### STEP 5B: Add File Search Tool

**5B.1 - Create `internal/tools/file/search.go`:**

**Implementation Details:**
```go
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

func (t GlobSearchTool) Execute(input json.RawMessage) (string, error) {
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
```

**Key Improvements:**
- **Named SearchResult type** instead of anonymous struct  
- **Input validation interface** with `Validate()` method
- **Extracted helper methods** for pattern building and search logic
- **Result type with String method** for flexible output formatting
- **Consistent error constants** and patterns

### STEP 5C: Enhance Read File Tool

**5C.1 - Update `internal/tools/file/read.go`:**

**Implementation Details:**
```go
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
    minLineNumber = 1
    minLimitValue = 1
    errMsgInvalidOffset = "offset must be >= %d (line numbers are 1-based)"
    errMsgInvalidLimit = "limit must be >= %d"
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
```

**Key Improvements:**
- **Cross-platform line handling** using `bufio.Scanner` instead of `strings.Split("\n")`
- **Named lineRange type** for better encapsulation of line calculations
- **Input validation interface** with `Validate()` method
- **Extracted helper methods** for parsing, validation, and line extraction
- **Constants for validation** instead of magic numbers
- **Better error messages** with consistent formatting
- **Edge case handling** for empty files and files without final newlines

**Additional Changes Needed:**
- Update the tool Description to document the new offset/limit parameters
- Update examples in the description to show offset/limit usage

## STEP 6: Final Integration & Testing

**6.1 - Create shared constants file (optional but recommended):**
Create `internal/tools/file/constants.go`:
```go
package file

// File permissions
const (
    DefaultFilePermissions = 0644
    DefaultDirPermissions  = 0755
)

// Validation constants  
const (
    MinLineNumber = 1
    MinLimitValue = 1
)

// Error message templates
const (
    ErrMsgMissingParam     = "parameter %q is required"
    ErrMsgOperationFailed  = "failed to %s: %w"
    ErrMsgFileNotFound     = "file not found: %s"
    ErrMsgIsDirectory      = "%s is a directory, not a file. Use directory tools for directory operations"
    ErrMsgInvalidPattern   = "invalid glob pattern '%s': %w"
    ErrMsgInvalidOffset    = "offset must be >= %d (line numbers are 1-based)"
    ErrMsgInvalidLimit     = "limit must be >= %d"
    ErrMsgOffsetTooLarge   = "offset %d exceeds file length (%d lines)"
)
```

**6.2 - Update imports in main.go (if needed):**
```go
// Ensure main.go imports the file package to register all tools
import (
    // ... existing imports ...
    
    // Import tool packages to register them
    _ "agent/internal/tools/file"  // This should register all file tools
)
```
**Note:** Since all tools are in the same `file` package, the existing import should automatically register all new tools via their `init()` functions.

**6.3 - Full system test:**
- Test all tools work together
- Verify AI agent can use all tools effectively
- Test edge cases and error conditions

**6.4 - Update documentation:**
- Update README.md with all new tools
- Document unified `write` tool and all new capabilities
- Add troubleshooting section

**6.5 - Performance validation:**
- Test with large files
- Test with many files
- Ensure reasonable performance

## Implementation Checklist

### Phase 1: Write Tool Unification
- [ ] Delete `internal/tools/file/create.go`
- [ ] Replace `internal/tools/file/write.go` with idiomatic unified version
- [ ] Implement input validation interface with `Validate()` method
- [ ] Extract helper methods for better separation of concerns
- [ ] Add constants for permissions and error messages
- [ ] Update tool registration (remove CreateFileTool, rename to "write")
- [ ] Test empty file creation
- [ ] Test file content writing  
- [ ] Test directory auto-creation
- [ ] Test file overwriting
- [ ] Update README.md

### Phase 2: Additional Tools
- [ ] Create shared constants file (`constants.go`) (optional but recommended)
- [ ] Implement idiomatic delete tool (`delete.go`) with helper methods
- [ ] Test file deletion functionality
- [ ] Implement idiomatic search tool (`search.go`) with named SearchResult type
- [ ] Test glob pattern searching
- [ ] Enhance read tool with cross-platform line handling and lineRange type
- [ ] Test large file reading with offset/limit
- [ ] Final integration testing
- [ ] Complete documentation update

## Success Criteria

**Phase 1 Complete When:**
- Single write tool handles both empty files and content
- AI agent uses write tool without confusion
- All file creation scenarios work correctly

**Phase 2 Complete When:**
- All Cursor file tool capabilities are available
- Tools work reliably in real scenarios  
- Documentation is comprehensive and accurate
- AI agent effectively uses complete toolset

## Quick Reference - New Tool Commands

**After implementation, AI agent will use:**
- `write` - Create empty files OR write content (unified, replaces create_file + write_file)
- `delete_file` - Delete files safely
- `glob_search` - Find files by pattern
- `read_file` - Read files (with optional offset/limit)
- `edit_file` - Single edit operations (unchanged)
- `list_files` - Directory listings (unchanged)
