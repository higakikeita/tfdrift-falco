package vcs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGitHubClient tests the initialization of GitHubClient.
func TestNewGitHubClient(t *testing.T) {
	client := NewGitHubClient("owner", "repo", "main", "token123")

	assert.NotNil(t, client)
	assert.Equal(t, "owner", client.owner)
	assert.Equal(t, "repo", client.repo)
	assert.Equal(t, "main", client.baseBranch)
	assert.Equal(t, "token123", client.token)
	assert.Equal(t, defaultAPIBase, client.apiBase)
	assert.NotNil(t, client.httpClient)
}

// TestCreatePR_Success tests successful PR creation with mocked API.
func TestCreatePR_Success(t *testing.T) {
	// Create a mock server that simulates GitHub API responses
	mockServer := createMockGitHubServer(t)
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "test-token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Fix: Auto-remediation for security group",
		Body:       "This PR auto-remediates a security group configuration drift.",
		BranchName: "fix/security-group-123",
		CommitMsg:  "Auto-remediate security group",
		Files: map[string]string{
			"terraform/main.tf": "resource \"aws_security_group\" \"example\" {\n  # Fixed config\n}",
		},
	}

	result, err := client.CreatePR(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Number)
	assert.Contains(t, result.URL, "/pull/")
}

// TestCreatePR_MissingTitle tests PR creation with missing title.
func TestCreatePR_MissingTitle(t *testing.T) {
	client := NewGitHubClient("owner", "repo", "main", "token")

	req := &PRRequest{
		Title:      "",
		Body:       "Body text",
		BranchName: "fix/something",
		CommitMsg:  "Fix something",
		Files:      map[string]string{"file.txt": "content"},
	}

	result, err := client.CreatePR(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "title is required")
}

// TestCreatePR_MissingBranchName tests PR creation with missing branch name.
func TestCreatePR_MissingBranchName(t *testing.T) {
	client := NewGitHubClient("owner", "repo", "main", "token")

	req := &PRRequest{
		Title:      "Fix something",
		Body:       "Body text",
		BranchName: "",
		CommitMsg:  "Fix something",
		Files:      map[string]string{"file.txt": "content"},
	}

	result, err := client.CreatePR(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "branch name is required")
}

// TestCreatePR_MissingCommitMsg tests PR creation with missing commit message.
func TestCreatePR_MissingCommitMsg(t *testing.T) {
	client := NewGitHubClient("owner", "repo", "main", "token")

	req := &PRRequest{
		Title:      "Fix something",
		Body:       "Body text",
		BranchName: "fix/something",
		CommitMsg:  "",
		Files:      map[string]string{"file.txt": "content"},
	}

	result, err := client.CreatePR(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "commit message is required")
}

// TestCreatePR_NilRequest tests PR creation with nil request.
func TestCreatePR_NilRequest(t *testing.T) {
	client := NewGitHubClient("owner", "repo", "main", "token")

	result, err := client.CreatePR(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "pr request cannot be nil")
}

// TestCreatePR_EmptyFiles tests PR creation with empty files map.
func TestCreatePR_EmptyFiles(t *testing.T) {
	mockServer := createMockGitHubServer(t)
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "test-token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Documentation update",
		Body:       "Update documentation only",
		BranchName: "docs/update",
		CommitMsg:  "Update docs",
		Files:      map[string]string{}, // Empty files
	}

	result, err := client.CreatePR(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Number)
}

// TestCreatePR_MultipleFiles tests PR creation with multiple files.
func TestCreatePR_MultipleFiles(t *testing.T) {
	mockServer := createMockGitHubServer(t)
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "test-token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Multi-file fix",
		Body:       "Fixes across multiple files",
		BranchName: "fix/multi-file",
		CommitMsg:  "Fix multiple issues",
		Files: map[string]string{
			"file1.tf": "# Fixed config 1",
			"file2.tf": "# Fixed config 2",
			"file3.tf": "# Fixed config 3",
		},
	}

	result, err := client.CreatePR(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Number)
}

// TestCreatePR_LongFileName tests PR creation with long file names.
func TestCreatePR_LongFileName(t *testing.T) {
	mockServer := createMockGitHubServer(t)
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "test-token")
	client.apiBase = mockServer.URL

	longFileName := "terraform/aws/security/vpc/security_groups/definitions/auto_remediation/example_very_long_resource_name.tf"
	req := &PRRequest{
		Title:      "Long file name test",
		Body:       "Testing long file paths",
		BranchName: "test/long-names",
		CommitMsg:  "Test long file names",
		Files: map[string]string{
			longFileName: "# Config",
		},
	}

	result, err := client.CreatePR(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestCreatePR_BranchCreationFails tests PR creation when branch creation fails.
func TestCreatePR_BranchCreationFails(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/ref/heads/main" {
			// Return base branch SHA
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(refResponse{Object: struct {
				SHA string `json:"sha"`
			}{SHA: "abc123"}})
			return
		}

		if r.URL.Path == "/repos/owner/repo/git/refs" && r.Method == http.MethodPost {
			// Branch creation fails
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"message":"Reference already exists"}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "test-token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Test PR",
		Body:       "Test",
		BranchName: "existing-branch",
		CommitMsg:  "Test",
		Files:      map[string]string{"file.txt": "content"},
	}

	result, err := client.CreatePR(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create branch")
}

// TestCreatePR_InvalidToken tests PR creation with invalid token.
func TestCreatePR_InvalidToken(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message":"Bad credentials"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "invalid-token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Test PR",
		Body:       "Test",
		BranchName: "test-branch",
		CommitMsg:  "Test",
		Files:      map[string]string{"file.txt": "content"},
	}

	result, err := client.CreatePR(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestDoRequest_AuthorizationHeader verifies proper authorization header format.
func TestDoRequest_AuthorizationHeader(t *testing.T) {
	var capturedAuth string

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "my-secret-token")
	client.apiBase = mockServer.URL

	_, err := client.doRequest(context.Background(), http.MethodGet, mockServer.URL, nil)

	require.NoError(t, err)
	assert.Equal(t, "Bearer my-secret-token", capturedAuth)
}

// TestDoRequest_ContentTypeHeader verifies proper content type header.
func TestDoRequest_ContentTypeHeader(t *testing.T) {
	var capturedContentType string

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	_, err := client.doRequest(context.Background(), http.MethodPost, mockServer.URL, []byte(`{}`))

	require.NoError(t, err)
	assert.Equal(t, "application/json", capturedContentType)
}

// TestDoRequest_AcceptHeader verifies proper accept header.
func TestDoRequest_AcceptHeader(t *testing.T) {
	var capturedAccept string

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	_, err := client.doRequest(context.Background(), http.MethodGet, mockServer.URL, nil)

	require.NoError(t, err)
	assert.Equal(t, "application/vnd.github.v3+json", capturedAccept)
}

// Helper function to create a mock GitHub API server for complete PR workflow testing.
func createMockGitHubServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get base branch SHA
		if r.URL.Path == "/repos/owner/repo/git/ref/heads/main" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(refResponse{
				Object: struct {
					SHA string `json:"sha"`
				}{SHA: "abc123def456"},
			})
			return
		}

		// Create new branch
		if r.URL.Path == "/repos/owner/repo/git/refs" && r.Method == http.MethodPost {
			var req createRefRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ref": req.Ref,
				"sha": req.SHA,
			})
			return
		}

		// Create tree
		if r.URL.Path == "/repos/owner/repo/git/trees" && r.Method == http.MethodPost {
			var req treeRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(treeResponse{SHA: "tree123"})
			return
		}

		// Create commit
		if r.URL.Path == "/repos/owner/repo/git/commits" && r.Method == http.MethodPost {
			var req commitRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(commitResponse{SHA: "commit123"})
			return
		}

		// Update branch reference
		if r.URL.Path == "/repos/owner/repo/git/refs/heads/fix/security-group-123" && r.Method == http.MethodPatch {
			var req updateRefRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ref": "refs/heads/fix/security-group-123",
				"sha": req.SHA,
			})
			return
		}

		// Handle generic branch update
		if r.Method == http.MethodPatch {
			var req updateRefRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"sha": req.SHA,
			})
			return
		}

		// Create pull request
		if r.URL.Path == "/repos/owner/repo/pulls" && r.Method == http.MethodPost {
			var req prRequestBody
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(prResponse{
				ID:      123,
				Number:  1,
				HTMLURL: "https://github.com/owner/repo/pull/1",
				State:   "open",
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
	}))
}
