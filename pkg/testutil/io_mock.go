package testutil

import (
	"bytes"
	"io"
)

// MockStdin creates a mock stdin reader with predefined input
// This allows testing of interactive prompts without actual user input
func NewMockStdin(input string) io.Reader {
	// Add newline to simulate Enter key
	if input != "" && input[len(input)-1] != '\n' {
		input += "\n"
	}
	return bytes.NewBufferString(input)
}

// MultiInputReader allows multiple sequential inputs for multiple prompts
// Useful for testing flows with multiple user interactions
type MultiInputReader struct {
	inputs []string
	index  int
}

// NewMultiInputReader creates a reader that returns different inputs for each Read call
func NewMultiInputReader(inputs ...string) *MultiInputReader {
	return &MultiInputReader{
		inputs: inputs,
		index:  0,
	}
}

// Read implements io.Reader interface
// Returns the next input from the list, automatically adding newlines
func (m *MultiInputReader) Read(p []byte) (n int, err error) {
	if m.index >= len(m.inputs) {
		return 0, io.EOF
	}

	input := m.inputs[m.index]
	if input != "" && input[len(input)-1] != '\n' {
		input += "\n"
	}
	m.index++

	n = copy(p, input)
	return n, nil
}

// MockReadWriter provides both read and write capabilities for testing
type MockReadWriter struct {
	*bytes.Buffer
}

// NewMockReadWriter creates a new mock read-writer
func NewMockReadWriter() *MockReadWriter {
	return &MockReadWriter{
		Buffer: &bytes.Buffer{},
	}
}

// SetInput sets the input to be read
func (m *MockReadWriter) SetInput(input string) {
	m.Buffer.Reset()
	m.Buffer.WriteString(input)
}

// GetOutput returns what was written
func (m *MockReadWriter) GetOutput() string {
	return m.Buffer.String()
}
