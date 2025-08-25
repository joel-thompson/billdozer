# Implementation TODO

## Steps to Refactor main.go

### 1. Create `tools.go`
- [ ] Extract `ToolDefinition` struct
- [ ] Move all tool input structs (ReadFileInput, ListFilesInput, EditFileInput)
- [ ] Move all tool implementations (ReadFile, ListFiles, EditFile)
- [ ] Move tool definitions (ReadFileDefinition, ListFilesDefinition, EditFileDefinition)
- [ ] Move `GenerateSchema` function
- [ ] Move helper functions (createNewFile)

### 2. Create `agent.go`
- [ ] Extract `Agent` struct
- [ ] Move `NewAgent` constructor
- [ ] Move `Run` method
- [ ] Move `executeTool` method
- [ ] Move `runInference` method

### 3. Refactor `main.go`
- [ ] Keep only main function
- [ ] Keep imports needed for main
- [ ] Keep tool list assembly
- [ ] Keep agent initialization
- [ ] Remove all extracted code

### 4. Verify Implementation
- [ ] Check that all files compile
- [ ] Test that functionality remains identical
- [ ] Verify clean file structure