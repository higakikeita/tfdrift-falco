package testutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Additional edge case and coverage tests for assertions

func TestAssertNotContains_NotPresent(t *testing.T) {
	str := "hello world"
	substr := "goodbye"
	AssertNotContains(t, str, substr, "should not contain substring")
}

func TestAssertLen_CorrectLen(t *testing.T) {
	slice := []string{"a", "b", "c"}
	AssertLen(t, slice, 3, "length should be 3")
}

func TestAssertNotEmpty_NotEmptySlice(t *testing.T) {
	slice := []string{"item"}
	AssertNotEmpty(t, slice, "should not be empty")
}

func TestAssertTrue_TrueValue(t *testing.T) {
	AssertTrue(t, true, "should be true")
}

func TestAssertFalse_FalseValue(t *testing.T) {
	AssertFalse(t, false, "should be false")
}

// Additional exec_mock tests for edge cases

func TestExecMocker_MultipleCommandsWithDifferentResults(t *testing.T) {
	mocker := NewExecMocker()

	mocker.OnCommand("cmd1", ExecResult{Stdout: "output1", Stderr: "", Error: nil})
	mocker.OnCommand("cmd2", ExecResult{Stdout: "output2", Stderr: "", Error: nil})

	out1, _, _ := mocker.Execute("cmd1")
	out2, _, _ := mocker.Execute("cmd2")

	assert.Equal(t, "output1", out1)
	assert.Equal(t, "output2", out2)
}

func TestExecMocker_ExactMatchVsBinaryMatch(t *testing.T) {
	mocker := NewExecMocker()

	// Set default for binary
	mocker.OnCommand("terraform", ExecResult{Stdout: "default", Stderr: "", Error: nil})
	// Set specific args match
	mocker.OnCommandWithArgs("terraform", []string{"version"}, ExecResult{Stdout: "v1.5.0", Stderr: "", Error: nil})

	// Exact match should be preferred
	out, _, _ := mocker.Execute("terraform", "version")
	assert.Equal(t, "v1.5.0", out)

	// Different args should use default
	out, _, _ = mocker.Execute("terraform", "plan")
	assert.Equal(t, "default", out)
}

func TestExecMocker_WasNotCalled(t *testing.T) {
	mocker := NewExecMocker()
	mocker.Execute("cmd1")

	assert.True(t, mocker.WasCalled("cmd1"))
	assert.False(t, mocker.WasCalled("cmd2"))
}

func TestExecMocker_WasCalledWith_PartialMatch(t *testing.T) {
	mocker := NewExecMocker()
	mocker.Execute("terraform", "plan", "-out=tf.plan")

	assert.True(t, mocker.WasCalledWith("terraform", "plan", "-out=tf.plan"))
	assert.False(t, mocker.WasCalledWith("terraform", "apply"))
}

func TestExecMocker_GetCallCount_Multiple(t *testing.T) {
	mocker := NewExecMocker()
	mocker.Execute("cmd1")
	mocker.Execute("cmd2")
	mocker.Execute("cmd3")

	assert.Equal(t, 3, mocker.GetCallCount())
}

func TestExecMocker_StderrCapture(t *testing.T) {
	mocker := NewExecMocker()
	mocker.OnCommand("cmd", ExecResult{
		Stdout: "output",
		Stderr: "error message",
		Error:  fmt.Errorf("failed"),
	})

	_, stderr, err := mocker.Execute("cmd")
	assert.Equal(t, "error message", stderr)
	assert.Error(t, err)
}

func TestExecMocker_EmptyBinaryName(t *testing.T) {
	mocker := NewExecMocker()
	mocker.SetDefaultResult(ExecResult{Stdout: "ok", Stderr: "", Error: nil})

	_, _, err := mocker.Execute("")
	assert.NoError(t, err)
}

// Additional fixtures tests

func TestCreateTempDir_CleanupRemovesDir(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "test-")
	assert.DirExists(t, dir)

	cleanup()
	assert.NoDirExists(t, dir)
}

func TestWriteTestFile_VerifyContent(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "test-")
	defer cleanup()

	testContent := "test content with special chars: !@#$%"
	path := WriteTestFile(t, dir, "test.txt", testContent)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(data))
}

func TestWriteTestFile_MultipleFiles(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "test-")
	defer cleanup()

	path1 := WriteTestFile(t, dir, "file1.txt", "content1")
	path2 := WriteTestFile(t, dir, "file2.txt", "content2")

	assert.FileExists(t, path1)
	assert.FileExists(t, path2)
}

func TestCreateTestConfig_ProvidersEnabled(t *testing.T) {
	cfg := CreateTestConfig()
	assert.True(t, cfg.Providers.AWS.Enabled)
	assert.Equal(t, "local", cfg.Providers.AWS.State.Backend)
}

func TestCreateTestConfig_FalcoEnabled(t *testing.T) {
	cfg := CreateTestConfig()
	assert.True(t, cfg.Falco.Enabled)
	assert.Equal(t, "localhost", cfg.Falco.Hostname)
	assert.Equal(t, uint16(5060), cfg.Falco.Port)
}

func TestCreateTestStateFile_ValidJSON(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "test-")
	defer cleanup()

	path := CreateTestStateFile(t, dir)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	// Basic JSON validation
	assert.Contains(t, string(data), "\"version\":")
	assert.Contains(t, string(data), "\"terraform_version\":")
}

func TestCreateTestEvent_AllFieldsSet(t *testing.T) {
	event := CreateTestEvent("aws_instance", "i-123", "StartInstances")

	assert.Equal(t, "aws", event.Provider)
	assert.Equal(t, "aws_instance", event.ResourceType)
	assert.Equal(t, "i-123", event.ResourceID)
	assert.Equal(t, "StartInstances", event.EventName)
	assert.Equal(t, "IAMUser", event.UserIdentity.Type)
	assert.NotNil(t, event.Changes)
}

func TestCreateTestDriftAlert_Metadata(t *testing.T) {
	alert := CreateTestDriftAlert()

	assert.Equal(t, "high", alert.Severity)
	assert.Equal(t, "aws_instance", alert.ResourceType)
	assert.Equal(t, "web", alert.ResourceName)
	assert.Equal(t, "drift", alert.AlertType)
	assert.Equal(t, "t3.micro", alert.OldValue)
	assert.Equal(t, "t3.small", alert.NewValue)
}

func TestCreateTestResource_Attributes(t *testing.T) {
	resource := CreateTestResource("aws_instance", "web", "i-456")

	assert.Equal(t, "aws_instance", resource.Type)
	assert.Equal(t, "web", resource.Name)
	assert.Equal(t, "i-456", resource.Attributes["id"])
	assert.Equal(t, "t3.micro", resource.Attributes["instance_type"])
	assert.Equal(t, "ami-12345678", resource.Attributes["ami"])
}

func TestCreateTestUnmanagedAlert_Properties(t *testing.T) {
	alert := CreateTestUnmanagedAlert()

	assert.Equal(t, "medium", alert.Severity)
	assert.Equal(t, "aws_s3_bucket", alert.ResourceType)
	assert.Equal(t, "unmanaged-bucket-123", alert.ResourceID)
	assert.Contains(t, alert.Changes, "bucket_name")
}

// Additional io_mock tests

func TestNewMockStdin_MultipleReads(t *testing.T) {
	input := "test input"
	reader := NewMockStdin(input)

	// First read
	data1 := make([]byte, 20)
	n1, err1 := reader.Read(data1)
	assert.NoError(t, err1)
	assert.Greater(t, n1, 0)

	// Second read should return EOF
	data2 := make([]byte, 20)
	_, err2 := reader.Read(data2)
	assert.Equal(t, io.EOF, err2)
}

func TestMultiInputReader_SequentialInputs(t *testing.T) {
	reader := NewMultiInputReader("first", "second", "third")

	data := make([]byte, 100)

	n1, _ := reader.Read(data)
	assert.Contains(t, string(data[:n1]), "first")

	n3, _ := reader.Read(data)
	assert.Contains(t, string(data[:n3]), "second")

	n4, _ := reader.Read(data)
	assert.Contains(t, string(data[:n4]), "third")

	// Next read should be EOF
	_, err := reader.Read(data)
	assert.Equal(t, io.EOF, err)
}

func TestMockReadWriter_ReadWrite(t *testing.T) {
	rw := NewMockReadWriter()

	// Write data
	written, err := rw.Write([]byte("write data"))
	assert.NoError(t, err)
	assert.Greater(t, written, 0)

	// Get output
	output := rw.GetOutput()
	assert.Equal(t, "write data", output)

	// Set and read input
	rw.SetInput("read data")
	data := make([]byte, 100)
	n, err := rw.Read(data)
	assert.NoError(t, err)
	assert.Greater(t, n, 0)
}

func TestMockReadWriter_Reset(t *testing.T) {
	rw := NewMockReadWriter()

	rw.Write([]byte("data"))
	assert.NotEmpty(t, rw.GetOutput())

	rw.SetInput("")
	assert.Equal(t, "", rw.GetOutput())
}

// Additional mock_falco tests

func TestMockFalcoClient_ConnectThenStream(t *testing.T) {
	client := NewMockFalcoClient()
	event := CreateTestEvent("aws_instance", "i-123", "RunInstances")
	client.AddEvent(event)

	err := client.Connect(nil)
	assert.NoError(t, err)
	assert.True(t, client.IsConnected())
}

func TestMockFalcoClient_SetStreamError(t *testing.T) {
	client := NewMockFalcoClient()
	client.SetStreamError(fmt.Errorf("stream failed"))

	event := CreateTestEvent("aws_instance", "i-123", "RunInstances")
	client.AddEvent(event)

	_ = client.Connect(nil)
	eventChan := make(chan *types.Event)
	err := client.StreamEvents(nil, eventChan)
	assert.Error(t, err)
}

func TestMockFalcoClient_MultipleConnections(t *testing.T) {
	client := NewMockFalcoClient()

	_ = client.Connect(nil)
	assert.True(t, client.IsConnected())

	client.Disconnect()
	assert.False(t, client.IsConnected())

	_ = client.Connect(nil)
	assert.True(t, client.IsConnected())
}

func TestMockFalcoClient_GetProcessedEventCount(t *testing.T) {
	client := NewMockFalcoClient()
	event1 := CreateTestEvent("aws_instance", "i-1", "RunInstances")
	event2 := CreateTestEvent("aws_s3_bucket", "bucket", "CreateBucket")

	client.AddEvents([]*types.Event{event1, event2})

	// Before streaming
	assert.Equal(t, 0, client.GetProcessedEventCount())
}

// Additional mock_http tests

func TestMockHTTPServer_CustomStatusCode(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	server.SetStatusCode(http.StatusCreated)
	resp, err := http.Get(server.URL())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestMockHTTPServer_CustomResponseBody(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	customBody := `{"error": "not found"}`
	server.SetResponseBody(customBody)

	resp, err := http.Get(server.URL())
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, customBody, string(body))
}

func TestMockHTTPServer_RequestWithHeaders(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL(), nil)
	req.Header.Set("X-Custom-Header", "custom-value")

	client := &http.Client{}
	client.Do(req)

	lastReq := server.GetLastRequest()
	assert.Equal(t, "custom-value", lastReq.Header.Get("X-Custom-Header"))
}

func TestMockHTTPServer_PostWithJSON(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	jsonBody := `{"key": "value"}`
	resp, err := http.Post(server.URL(), "application/json", bytes.NewBufferString(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, jsonBody, server.GetLastRequestBody())
}

func TestMockHTTPServer_MultipleStatusCodeChanges(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	server.SetStatusCode(http.StatusOK)
	resp1, _ := http.Get(server.URL())
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	resp1.Body.Close()

	server.SetStatusCode(http.StatusInternalServerError)
	resp2, _ := http.Get(server.URL())
	assert.Equal(t, http.StatusInternalServerError, resp2.StatusCode)
	resp2.Body.Close()
}

func TestMockHTTPServer_EmptyRequestBody(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	http.Get(server.URL())
	assert.Equal(t, "", server.GetLastRequestBody())
}

func TestMockHTTPServer_URL_NotEmpty(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	assert.NotEmpty(t, server.URL())
	assert.Contains(t, server.URL(), "http")
}

func TestMockHTTPServer_GetLastRequest_NoRequests(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	assert.Nil(t, server.GetLastRequest())
}

func TestMockHTTPServer_GetLastRequestBody_NoRequests(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	assert.Equal(t, "", server.GetLastRequestBody())
}

func TestMockHTTPServer_ResetClearsRequests(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	http.Get(server.URL())
	http.Get(server.URL())
	assert.Equal(t, 2, server.GetRequestCount())

	server.Reset()
	assert.Equal(t, 0, server.GetRequestCount())
}

// Concurrency tests

func TestMockFalcoClient_ConcurrentOps(t *testing.T) {
	client := NewMockFalcoClient()
	done := make(chan bool)

	// Add events concurrently
	for i := 0; i < 5; i++ {
		go func(id int) {
			event := CreateTestEvent("aws_instance", fmt.Sprintf("i-%d", id), "RunInstances")
			client.AddEvent(event)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	assert.Equal(t, 5, client.GetEventCount())
}

func TestMockHTTPServer_ConcurrentRequests(t *testing.T) {
	server := NewMockHTTPServer()
	defer server.Close()

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			http.Get(server.URL())
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, 10, server.GetRequestCount())
}
