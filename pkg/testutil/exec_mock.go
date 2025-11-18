package testutil

import (
	"fmt"
	"strings"
	"sync"
)

// ExecCall represents a single command execution
type ExecCall struct {
	Binary string
	Args   []string
}

// String returns a string representation of the command
func (e ExecCall) String() string {
	return fmt.Sprintf("%s %s", e.Binary, strings.Join(e.Args, " "))
}

// ExecResult represents the result of a command execution
type ExecResult struct {
	Stdout string
	Stderr string
	Error  error
}

// ExecMocker mocks command execution for testing
// It records all commands executed and returns predefined results
type ExecMocker struct {
	mu       sync.Mutex
	calls    []ExecCall
	behavior map[string]ExecResult
	// Default behavior when no specific mock is set
	defaultResult ExecResult
}

// NewExecMocker creates a new command execution mocker
func NewExecMocker() *ExecMocker {
	return &ExecMocker{
		calls:    make([]ExecCall, 0),
		behavior: make(map[string]ExecResult),
		defaultResult: ExecResult{
			Stdout: "",
			Stderr: "",
			Error:  nil,
		},
	}
}

// OnCommand sets the expected result for a specific command
// The command string should match the binary name
func (e *ExecMocker) OnCommand(binary string, result ExecResult) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.behavior[binary] = result
}

// OnCommandWithArgs sets the expected result for a command with specific arguments
func (e *ExecMocker) OnCommandWithArgs(binary string, args []string, result ExecResult) {
	e.mu.Lock()
	defer e.mu.Unlock()
	key := fmt.Sprintf("%s %s", binary, strings.Join(args, " "))
	e.behavior[key] = result
}

// SetDefaultResult sets the default result for commands without specific mocks
func (e *ExecMocker) SetDefaultResult(result ExecResult) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.defaultResult = result
}

// Execute simulates command execution and records the call
func (e *ExecMocker) Execute(binary string, args ...string) (string, string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	call := ExecCall{Binary: binary, Args: args}
	e.calls = append(e.calls, call)

	// Try exact match first (binary + args)
	key := call.String()
	if result, ok := e.behavior[key]; ok {
		return result.Stdout, result.Stderr, result.Error
	}

	// Try binary-only match
	if result, ok := e.behavior[binary]; ok {
		return result.Stdout, result.Stderr, result.Error
	}

	// Return default result
	return e.defaultResult.Stdout, e.defaultResult.Stderr, e.defaultResult.Error
}

// GetCalls returns all recorded command calls
func (e *ExecMocker) GetCalls() []ExecCall {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]ExecCall{}, e.calls...)
}

// GetCallCount returns the number of times commands were executed
func (e *ExecMocker) GetCallCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.calls)
}

// WasCalled checks if a specific command was called
func (e *ExecMocker) WasCalled(binary string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, call := range e.calls {
		if call.Binary == binary {
			return true
		}
	}
	return false
}

// WasCalledWith checks if a command was called with specific arguments
func (e *ExecMocker) WasCalledWith(binary string, args ...string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, call := range e.calls {
		if call.Binary == binary && stringSliceEqual(call.Args, args) {
			return true
		}
	}
	return false
}

// Reset clears all recorded calls and behaviors
func (e *ExecMocker) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls = make([]ExecCall, 0)
	e.behavior = make(map[string]ExecResult)
}

// Helper function to compare string slices
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TerraformMocker provides convenience methods for mocking terraform commands
type TerraformMocker struct {
	*ExecMocker
}

// NewTerraformMocker creates a mocker specifically for terraform commands
func NewTerraformMocker() *TerraformMocker {
	return &TerraformMocker{
		ExecMocker: NewExecMocker(),
	}
}

// MockImportSuccess mocks a successful terraform import
func (t *TerraformMocker) MockImportSuccess(resourceType, resourceName, resourceID string) {
	stdout := fmt.Sprintf("Import successful!\nResource: %s.%s (%s)", resourceType, resourceName, resourceID)
	t.OnCommand("terraform", ExecResult{
		Stdout: stdout,
		Stderr: "",
		Error:  nil,
	})
}

// MockImportFailure mocks a failed terraform import
func (t *TerraformMocker) MockImportFailure(errorMsg string) {
	t.OnCommand("terraform", ExecResult{
		Stdout: "",
		Stderr: errorMsg,
		Error:  fmt.Errorf("terraform import failed"),
	})
}

// MockVersionSuccess mocks a successful terraform version command
func (t *TerraformMocker) MockVersionSuccess(version string) {
	t.OnCommandWithArgs("terraform", []string{"version"}, ExecResult{
		Stdout: fmt.Sprintf("Terraform v%s", version),
		Stderr: "",
		Error:  nil,
	})
}
