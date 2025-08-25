package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"agent/internal/agent"
	"agent/internal/tools"
	"github.com/anthropics/anthropic-sdk-go"

	// Import tool packages to register them
	_ "agent/internal/tools/file"
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

	// Get all registered tools from the registry
	registeredTools := tools.DefaultRegistry.GetAll()

	// Initialize and start agent
	agentInstance := agent.NewAgent(&client, getUserMessage, registeredTools)
	err := agentInstance.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
