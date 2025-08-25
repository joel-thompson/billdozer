# Command Execution Tool Plan

## Overview

Add a command execution tool that allows the AI agent to run whitelisted commands (linting, tests, build, etc.) with per-project configuration.

## Goals

- Enable AI agent to validate code it creates by running tests, linters, build commands
- Maintain security through a whitelist approach - agent cannot run arbitrary commands  
- Support per-project configuration rather than hardcoded commands
- Include descriptions for each command to help the agent understand when to use them

## Architecture

### Configuration File

Create a `.agent-commands.yml` file in project root:

```yaml
version: "1.0"
commands:
  lint:
    command: "golangci-lint run"
    description: "Run Go linter to check code quality and style"
    working_directory: "."
    timeout: 30s
  
  test:
    command: "go test ./..."
    description: "Run all Go tests in the project"
    working_directory: "."
    timeout: 60s
  
  build:
    command: "go build -o bin/app main.go"
    description: "Build the Go application"
    working_directory: "."
    timeout: 30s
  
  tidy:
    command: "go mod tidy"
    description: "Clean up Go module dependencies"
    working_directory: "."
    timeout: 10s
```

### Implementation Structure

1. **Config Package** (`internal/config/`)
   - `commands.go` - Parse `.agent-commands.yml`
   - `types.go` - Configuration struct definitions

2. **Command Tool** (`internal/tools/command/`)
   - `execute.go` - Command execution tool implementation
   - Input validation, timeout handling, security checks

3. **Tool Registration**
   - Auto-register command tool in `main.go`
   - Tool discovers and loads project config automatically

## Implementation Details

### Configuration Schema

```go
type CommandsConfig struct {
    Version  string                 `yaml:"version"`
    Commands map[string]CommandSpec `yaml:"commands"`
}

type CommandSpec struct {
    Command          string        `yaml:"command"`
    Description      string        `yaml:"description"`
    WorkingDirectory string        `yaml:"working_directory"`
    Timeout          time.Duration `yaml:"timeout"`
}
```

### Tool Input

```go
type CommandInput struct {
    Name string `json:"name" jsonschema_description:"Name of the command to execute from project configuration"`
}
```

### Security Features

- Only commands defined in `.agent-commands.yml` can be executed
- Configurable timeouts prevent runaway processes
- Working directory is restricted to project scope
- Command output is captured and returned to agent

### Error Handling

- Missing config file: Inform agent no commands are available
- Invalid command name: List available commands  
- Command failure: Return exit code and error output
- Timeout: Terminate process and return timeout error

## Usage Flow

1. Agent needs to validate code (e.g., after making changes)
2. Agent calls `execute_command` tool with command name (e.g., "test")
3. Tool reads `.agent-commands.yml` from project root
4. Tool validates command exists and executes it with configured settings
5. Tool returns command output and exit status to agent
6. Agent can interpret results and take appropriate action

## Benefits

- **Security**: Whitelisted commands only, no arbitrary execution
- **Flexibility**: Each project defines its own relevant commands  
- **Context**: Command descriptions help agent choose appropriate tools
- **Reliability**: Timeouts and error handling prevent issues
- **Maintainability**: Configuration separate from code

## Next Steps

1. Implement configuration parsing in `internal/config/`
2. Create command execution tool in `internal/tools/command/`
3. Add tool registration to `main.go`
4. Create example `.agent-commands.yml` for this project
5. Test with common Go development workflows
