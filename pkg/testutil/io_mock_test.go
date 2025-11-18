package testutil

import (
	"bufio"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockStdin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple input",
			input:    "yes",
			expected: "yes\n",
		},
		{
			name:     "Input with newline",
			input:    "no\n",
			expected: "no\n",
		},
		{
			name:     "Single character",
			input:    "y",
			expected: "y\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewMockStdin(tt.input)
			scanner := bufio.NewScanner(reader)

			require.True(t, scanner.Scan())
			assert.Equal(t, tt.expected[:len(tt.expected)-1], scanner.Text())
		})
	}
}

func TestMultiInputReader(t *testing.T) {
	inputs := []string{"first", "second", "third"}
	reader := NewMultiInputReader(inputs...)

	scanner := bufio.NewScanner(reader)

	// Read all inputs
	var results []string
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}

	assert.Equal(t, inputs, results)
}

func TestMultiInputReader_EOF(t *testing.T) {
	reader := NewMultiInputReader("only-one")

	buf := make([]byte, 100)

	// First read should succeed
	n, err := reader.Read(buf)
	assert.NoError(t, err)
	assert.Greater(t, n, 0)

	// Second read should return EOF
	_, err = reader.Read(buf)
	assert.ErrorIs(t, err, io.EOF)
}

func TestMockReadWriter(t *testing.T) {
	rw := NewMockReadWriter()

	// Test write
	input := "test input"
	rw.SetInput(input)

	// Test read
	buf := make([]byte, len(input))
	n, err := rw.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(input), n)
	assert.Equal(t, input, string(buf))
}

func TestMockReadWriter_GetOutput(t *testing.T) {
	rw := NewMockReadWriter()

	// Write some data
	output1 := "first line"
	n, err := rw.Write([]byte(output1))
	require.NoError(t, err)
	assert.Equal(t, len(output1), n)

	// Check output
	assert.Equal(t, output1, rw.GetOutput())

	// Write more data
	output2 := " second line"
	n, err = rw.Write([]byte(output2))
	require.NoError(t, err)
	assert.Equal(t, len(output2), n)

	// Check combined output
	assert.Equal(t, output1+output2, rw.GetOutput())
}
