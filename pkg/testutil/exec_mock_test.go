package testutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecMocker_OnCommand(t *testing.T) {
	mocker := NewExecMocker()

	expectedResult := ExecResult{
		Stdout: "success output",
		Stderr: "",
		Error:  nil,
	}

	mocker.OnCommand("test-binary", expectedResult)

	stdout, stderr, err := mocker.Execute("test-binary", "arg1", "arg2")

	assert.Equal(t, expectedResult.Stdout, stdout)
	assert.Equal(t, expectedResult.Stderr, stderr)
	assert.Equal(t, expectedResult.Error, err)
}

func TestExecMocker_GetCalls(t *testing.T) {
	mocker := NewExecMocker()

	_, _, _ = mocker.Execute("cmd1", "arg1")
	_, _, _ = mocker.Execute("cmd2", "arg2", "arg3")

	calls := mocker.GetCalls()

	assert.Len(t, calls, 2)
	assert.Equal(t, "cmd1", calls[0].Binary)
	assert.Equal(t, []string{"arg1"}, calls[0].Args)
	assert.Equal(t, "cmd2", calls[1].Binary)
	assert.Equal(t, []string{"arg2", "arg3"}, calls[1].Args)
}

func TestExecMocker_WasCalled(t *testing.T) {
	mocker := NewExecMocker()

	_, _, _ = mocker.Execute("terraform", "version")
	_, _, _ = mocker.Execute("git", "status")

	assert.True(t, mocker.WasCalled("terraform"))
	assert.True(t, mocker.WasCalled("git"))
	assert.False(t, mocker.WasCalled("docker"))
}

func TestExecMocker_WasCalledWith(t *testing.T) {
	mocker := NewExecMocker()

	_, _, _ = mocker.Execute("terraform", "import", "aws_instance.web", "i-123")

	assert.True(t, mocker.WasCalledWith("terraform", "import", "aws_instance.web", "i-123"))
	assert.False(t, mocker.WasCalledWith("terraform", "import", "aws_s3_bucket.data", "bucket-456"))
}

func TestExecMocker_DefaultResult(t *testing.T) {
	mocker := NewExecMocker()

	defaultResult := ExecResult{
		Stdout: "default output",
		Stderr: "default error",
		Error:  fmt.Errorf("default error"),
	}
	mocker.SetDefaultResult(defaultResult)

	// Command with no specific mock should return default
	stdout, stderr, err := mocker.Execute("unknown-command")

	assert.Equal(t, defaultResult.Stdout, stdout)
	assert.Equal(t, defaultResult.Stderr, stderr)
	assert.Equal(t, defaultResult.Error, err)
}

func TestExecMocker_Reset(t *testing.T) {
	mocker := NewExecMocker()

	mocker.OnCommand("test", ExecResult{Stdout: "output"})
	_, _, _ = mocker.Execute("test")

	assert.Equal(t, 1, mocker.GetCallCount())

	mocker.Reset()

	assert.Equal(t, 0, mocker.GetCallCount())
	stdout, _, _ := mocker.Execute("test")
	assert.Empty(t, stdout) // Behavior was reset
}

func TestTerraformMocker_MockImportSuccess(t *testing.T) {
	mocker := NewTerraformMocker()

	mocker.MockImportSuccess("aws_instance", "web", "i-123")

	stdout, stderr, err := mocker.Execute("terraform", "import", "aws_instance.web", "i-123")

	assert.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Import successful")
	assert.Contains(t, stdout, "aws_instance.web")
}

func TestTerraformMocker_MockImportFailure(t *testing.T) {
	mocker := NewTerraformMocker()

	mocker.MockImportFailure("resource not found")

	stdout, stderr, err := mocker.Execute("terraform", "import", "aws_instance.web", "i-999")

	assert.Error(t, err)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "resource not found")
}

func TestTerraformMocker_MockVersionSuccess(t *testing.T) {
	mocker := NewTerraformMocker()

	mocker.MockVersionSuccess("1.6.0")

	stdout, stderr, err := mocker.Execute("terraform", "version")

	assert.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Terraform v1.6.0")
}

func TestExecCall_String(t *testing.T) {
	call := ExecCall{
		Binary: "terraform",
		Args:   []string{"import", "aws_instance.web", "i-123"},
	}

	expected := "terraform import aws_instance.web i-123"
	assert.Equal(t, expected, call.String())
}
