package tools

import (
	"sync"
)

// Registry manages tool registration and retrieval
type Registry struct {
	tools []ToolDefinition
	mutex sync.RWMutex
}

// Register adds a tool to the registry
func (r *Registry) Register(tool ToolDefinition) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.tools = append(r.tools, tool)
}

// RegisterTool adds a Tool interface implementation to the registry
func (r *Registry) RegisterTool(tool Tool) {
	r.Register(ToolAdapter(tool))
}

// GetAll returns all registered tools
func (r *Registry) GetAll() []ToolDefinition {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	// Return a copy to prevent external modification
	result := make([]ToolDefinition, len(r.tools))
	copy(result, r.tools)
	return result
}

// GetByName returns a tool by its name, or nil if not found
func (r *Registry) GetByName(name string) *ToolDefinition {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, tool := range r.tools {
		if tool.Name == name {
			// Return a copy to prevent external modification
			toolCopy := tool
			return &toolCopy
		}
	}
	return nil
}

// Clear removes all tools from the registry (useful for testing)
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.tools = r.tools[:0]
}

// DefaultRegistry is the global registry instance
var DefaultRegistry = &Registry{}
