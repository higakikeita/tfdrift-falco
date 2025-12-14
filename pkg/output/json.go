package output

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// JSONOutput writes drift events as NDJSON (newline-delimited JSON) to a writer
type JSONOutput struct {
	writer io.Writer
	mu     sync.Mutex
}

// NewJSONOutput creates a new JSON output writer
// By default, writes to stdout
func NewJSONOutput() *JSONOutput {
	return &JSONOutput{
		writer: os.Stdout,
	}
}

// NewJSONOutputWithWriter creates a new JSON output with a custom writer
func NewJSONOutputWithWriter(w io.Writer) *JSONOutput {
	return &JSONOutput{
		writer: w,
	}
}

// Write writes a drift event as a single JSON line
func (j *JSONOutput) Write(event *types.DriftEvent) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	jsonStr, err := event.ToJSONString()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Write JSON line (NDJSON format)
	_, err = fmt.Fprintln(j.writer, jsonStr)
	if err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

// Close closes the output (if the writer is closable)
func (j *JSONOutput) Close() error {
	if closer, ok := j.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
