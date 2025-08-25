# Agent CLI

A modular AI agent CLI built in Go that can execute tools through conversation with Claude. The architecture makes it easy to add new tools without modifying existing code.

## Quick Start

### Prerequisites

- Go 1.24+ installed
- Anthropic API key set as environment variable `ANTHROPIC_API_KEY`

### Running the CLI

1. Navigate to the project directory and install dependencies with `go mod tidy`
2. Set your Anthropic API key as an environment variable
3. Run the application with `go run main.go`

The CLI starts an interactive conversation with Claude. Claude automatically uses available tools when appropriate for tasks like reading files, writing files, searching for files by patterns, deleting files, listing directories, or editing text files.

## Why This Architecture

This design prioritizes maintainability and extensibility:

- **Zero Configuration** - New tools register automatically without manual setup
- **Complete Isolation** - Each tool is self-contained and independently testable
- **Logical Organization** - Related tools are grouped in packages
- **Type Safety** - Go's compiler catches errors at build time
- **Effortless Scaling** - Add dozens of tools without increasing complexity

The auto-registration system eliminates the most common source of maintenance overhead: manually updating tool lists across multiple files.

## Project Structure

The modular architecture separates concerns clearly:

- **main.go** - CLI entry point and orchestration
- **internal/agent/** - Conversation management and Claude integration  
- **internal/schema/** - JSON schema generation utilities
- **internal/tools/** - Tool interfaces, registry, and implementations
  - **types.go** - Common interfaces and type definitions
  - **registry.go** - Automatic tool registration system  
  - **file/** - File operation tools (read, list, write, delete_file, glob_search, edit)
  - **[other packages]** - Additional tool categories as needed

## Available Tools

### File Operations

- **`write`** - Unified file creation and writing tool
  - Creates empty files when content is omitted: `{"path": "empty.txt", "content": ""}`
  - Writes content to files: `{"path": "config.yml", "content": "version: 1.0\nname: myapp"}`
  - Always overwrites existing files (eliminates "file exists" errors)
  - Auto-creates parent directories as needed

- **`read_file`** - Enhanced file reading with line range support
  - Read entire files: `{"path": "main.go"}`
  - Read from specific line: `{"path": "config.yml", "offset": 10}`
  - Read line ranges: `{"path": "data.txt", "offset": 5, "limit": 20}`
  - Cross-platform line ending support

- **`delete_file`** - Safe file deletion
  - Deletes existing files: `{"path": "unwanted_file.txt"}`
  - Validates file exists before deletion
  - Only deletes files, not directories
  - Clear error messages for safety

- **`glob_search`** - Pattern-based file searching
  - Find files by pattern: `{"pattern": "*.go"}`
  - Search in specific directories: `{"pattern": "test_*.txt", "path": "tests"}`
  - Supports glob patterns (`*`, `?`, `[abc]`)
  - Returns structured results with match count

- **`list_files`** - Directory listing (existing tool)

- **`edit_file`** - Single edit operations (existing tool)

## Working With Tools

### Adding New Tools

1. **Create package** - Make a new directory under `internal/tools/` for your tool category
2. **Implement tool** - Create a struct with `Definition()` and `Execute()` methods
3. **Define input** - Create an input struct with JSON schema tags for parameters  
4. **Auto-register** - Add `init()` function that calls `tools.DefaultRegistry.RegisterTool()`
5. **Import package** - Add import to `main.go` with `_` prefix to trigger registration

### Modifying Existing Tools

1. Navigate to the tool file in its package directory
2. Modify the implementation in the `Execute` method  
3. Update input schema if you change parameters
4. No other files need changes - the registry handles everything automatically

Tools are completely self-contained, so changes only affect the individual tool file.

### Testing Tools

Create test files alongside tool implementations. Test the `Execute` method directly with mock JSON input to verify behavior without depending on the full agent system. Each tool can be tested in complete isolation.

### Best Practices

- Use descriptive JSON schema descriptions for better Claude integration
- Add `jsonschema:"required"` tags for mandatory parameters
- Include concrete usage examples in tool descriptions  
- Provide actionable error messages with examples of correct usage
- Group related tools in the same package
- Handle errors gracefully with clear error messages
- Use the `schema.GenerateSchema[T]()` helper for input validation
- Test tools individually before integrating

## Contributing

1. Follow the tool patterns described above
2. Add tests for new tools
3. Update this README if you add new tool categories
4. Ensure tools follow the established conventions
