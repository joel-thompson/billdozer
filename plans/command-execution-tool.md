# Command Execution Tool Implementation Guide

## What We're Building

A simple tool that lets the AI agent run whitelisted validation commands (lint, test, build) to validate code changes. Commands are configured per-project in `.agent-commands.yml` for security.

## Step 1: Create Configuration Types

Create `internal/config/commands.go`:

```go
package config

import (
    "fmt"
    "os"
    "gopkg.in/yaml.v3"
)

type CommandsConfig struct {
    Commands map[string]CommandSpec `yaml:"commands"`
}

type CommandSpec struct {
    Command        string `yaml:"command"`
    Description    string `yaml:"description"`
    TimeoutSeconds int    `yaml:"timeout_seconds"`
}

func LoadCommandsConfig(path string) (*CommandsConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config CommandsConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    return &config, nil
}
```

## Step 2: Create Command Tool

Create `internal/tools/command/execute.go`:

```go
package command

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"
    "time"
    
    "agent/internal/config"
    "agent/internal/schema"
    "agent/internal/tools"
)

// Error message constants
const (
    errMsgMissingParam     = "parameter %q is required"
    errMsgOperationFailed  = "failed to %s: %w"
    errMsgCommandNotFound  = "unknown command %q. Available commands: %s"
    errMsgEmptyCommand     = "empty command for %q"
    errMsgCommandFailed    = "command %q failed: %w"
)

type CommandInput struct {
    Name string `json:"name" jsonschema:"required" jsonschema_description:"Name of the command to execute from project configuration, or 'list' to show available commands"`
}

// Validate implements input validation
func (c *CommandInput) Validate() error {
    if c.Name == "" {
        return fmt.Errorf(errMsgMissingParam, "name")
    }
    return nil
}

type CommandTool struct {
    config *config.CommandsConfig
}

func (t CommandTool) Definition() tools.ToolDefinition {
    return tools.ToolDefinition{
        Name: "execute_command",
        Description: `Execute predefined commands for code validation (lint, test, build).

Usage Examples:
- {"name": "list"} // Show available commands
- {"name": "lint"} // Run linter
- {"name": "test"} // Run tests  
- {"name": "build"} // Build application

Security:
- Only commands defined in .agent-commands.yml can be executed
- Commands have timeouts to prevent hanging processes
- No arbitrary command execution allowed

Use this tool after making code changes to validate they work correctly.`,
        InputSchema: schema.GenerateSchema[CommandInput](),
    }
}

func (t CommandTool) Execute(ctx *tools.ToolContext, input json.RawMessage) (string, error) {
    commandInput, err := t.parseAndValidateInput(input)
    if err != nil {
        return "", err
    }
    
    // Load config each time to pick up changes
    config, err := config.LoadCommandsConfig(".agent-commands.yml")
    if err != nil {
        return "", fmt.Errorf("failed to load command configuration: %w", err)
    }
    
    // Handle list command
    if commandInput.Name == "list" {
        return t.listCommands(config), nil
    }
    
    // Execute specific command
    return t.executeCommand(config, commandInput.Name)
}

// Helper methods for better separation of concerns
func (t CommandTool) parseAndValidateInput(input json.RawMessage) (*CommandInput, error) {
    var commandInput CommandInput
    if err := json.Unmarshal(input, &commandInput); err != nil {
        return nil, fmt.Errorf("invalid JSON input: %w", err)
    }
    
    if err := commandInput.Validate(); err != nil {
        return nil, err
    }
    
    return &commandInput, nil
}

func (t CommandTool) listCommands(config *config.CommandsConfig) string {
    if len(config.Commands) == 0 {
        return "No commands available. Create .agent-commands.yml file with command definitions."
    }
    
    var result strings.Builder
    result.WriteString("Available commands:\n")
    for name, spec := range config.Commands {
        result.WriteString(fmt.Sprintf("- %s: %s\n", name, spec.Description))
    }
    return result.String()
}

func (t CommandTool) executeCommand(config *config.CommandsConfig, commandName string) (string, error) {
    spec, exists := config.Commands[commandName]
    if !exists {
        return "", fmt.Errorf(errMsgCommandNotFound, commandName, t.getCommandNames(config))
    }
    
    // Parse command and args
    parts := strings.Fields(spec.Command)
    if len(parts) == 0 {
        return "", fmt.Errorf(errMsgEmptyCommand, commandName)
    }
    
    // Create command with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 
        time.Duration(spec.TimeoutSeconds)*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        // Return output even on failure so agent can see error details
        return string(output), fmt.Errorf(errMsgCommandFailed, commandName, err)
    }
    
    return string(output), nil
}

func (t CommandTool) getCommandNames(config *config.CommandsConfig) string {
    var names []string
    for name := range config.Commands {
        names = append(names, name)
    }
    return strings.Join(names, ", ")
}

func init() {
    tools.DefaultRegistry.RegisterTool(CommandTool{})
}
```

## Step 3: Register Tool

Add import to `main.go`:

```go
// Import tool packages to register them
_ "agent/internal/tools/file"
_ "agent/internal/tools/command"  // Add this line
```

The `init()` function in the command package will automatically register the tool when imported.

## Step 4: Create Example Configuration

Create `.agent-commands.yml` in project root:

```yaml
commands:
  lint:
    command: "golangci-lint run"
    description: "Run Go linter to check code quality and style"
    timeout_seconds: 30
  
  test:
    command: "go test ./..."
    description: "Run all Go tests in the project"
    timeout_seconds: 60
  
  build:
    command: "go build -o bin/app main.go"
    description: "Build the Go application"
    timeout_seconds: 30
```

## Step 5: Test Implementation

Test these scenarios:

1. **List commands**: Call with `{"name": "list"}` - should show available commands
2. **Run lint**: Call with `{"name": "lint"}` - should run golangci-lint
3. **Run test**: Call with `{"name": "test"}` - should run go test
4. **Invalid command**: Call with `{"name": "xyz"}` - should show error with available commands
5. **Missing config**: Remove `.agent-commands.yml` - should handle gracefully

## Expected Agent Workflow

1. Agent makes code changes
2. Agent calls `execute_command` with `name: "list"` to see what's available
3. Agent gets: "Available commands:\n- lint: Run Go linter...\n- test: Run all tests..."
4. Agent calls `execute_command` with `name: "lint"` to check code quality
5. Agent gets lint results, fixes issues if needed
6. Agent calls `execute_command` with `name: "test"` to run tests
7. Agent gets test results, fixes issues if needed
8. Agent calls `execute_command` with `name: "build"` to ensure it compiles
9. Agent is confident the code works

## Security Features

- Only commands in `.agent-commands.yml` can run
- Commands have timeouts to prevent hanging
- No arbitrary command execution
- Commands run in project directory only
