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
	errMsgMissingParam    = "parameter %q is required"
	errMsgOperationFailed = "failed to %s: %w"
	errMsgCommandNotFound = "unknown command %q. Available commands: %s"
	errMsgEmptyCommand    = "empty command for %q"
	errMsgCommandFailed   = "command %q failed: %w"
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

type CommandTool struct{}

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
