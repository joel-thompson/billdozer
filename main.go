package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

// main is the application entry point
func main() {
	client := anthropic.NewClient()

	// Set up user input scanner
	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	// Assemble available tools
	tools := []ToolDefinition{ReadFileDefinition, ListFilesDefinition, EditFileDefinition}

	// Initialize and start agent
	agent := NewAgent(&client, getUserMessage, tools)
	err := agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
