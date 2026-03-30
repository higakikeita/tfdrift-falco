package vcs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetBaseBranchSHA_Success tests successful retrieval of base branch SHA
func TestGetBaseBranchSHA_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/ref/heads/main" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(refResponse{
				Object: struct {
					SHA string `json:"sha"`
				}{SHA: "abc123def456"},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	sha, err := client.getBaseBranchSHA(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "abc123def456", sha)
}

func TestGetBaseBranchSHA_NotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	sha, err := client.getBaseBranchSHA(context.Background())

	assert.Error(t, err)
	assert.Empty(t, sha)
	assert.Contains(t, err.Error(), "github api error")
	assert.Contains(t, err.Error(), "404")
}

func TestCreateBranch_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/refs" && r.Method == http.MethodPost {
			var req createRefRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			assert.Equal(t, "refs/heads/feature-branch", req.Ref)
			assert.Equal(t, "abc123", req.SHA)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"ref": req.Ref,
				"sha": req.SHA,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	err := client.createBranch(context.Background(), "feature-branch", "abc123")

	require.NoError(t, err)
}

func TestCreateBranch_AlreadyExists(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/refs" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"message":"Reference already exists"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	err := client.createBranch(context.Background(), "existing-branch", "abc123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "github api error")
	assert.Contains(t, err.Error(), "422")
}

func TestCreateTree_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/trees" && r.Method == http.MethodPost {
			var req treeRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			assert.Len(t, req.Tree, 2)
			assert.Equal(t, "file1.txt", req.Tree[0].Path)
			assert.Equal(t, "100644", req.Tree[0].Mode)
			assert.Equal(t, "blob", req.Tree[0].Type)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(treeResponse{SHA: "tree123"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}

	sha, err := client.createTree(context.Background(), files, "base123")

	require.NoError(t, err)
	assert.Equal(t, "tree123", sha)
}

func TestCreateTree_EmptyFiles(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/trees" && r.Method == http.MethodPost {
			var req treeRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Len(t, req.Tree, 0)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(treeResponse{SHA: "tree123"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	sha, err := client.createTree(context.Background(), map[string]string{}, "base123")

	require.NoError(t, err)
	assert.Equal(t, "tree123", sha)
}

func TestCreateCommit_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/commits" && r.Method == http.MethodPost {
			var req commitRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			assert.Equal(t, "Fix: update terraform", req.Message)
			assert.Equal(t, "tree123", req.Tree)
			assert.Len(t, req.Parents, 1)
			assert.Equal(t, "parent123", req.Parents[0])
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(commitResponse{SHA: "commit456"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	sha, err := client.createCommit(context.Background(), "Fix: update terraform", "tree123", []string{"parent123"})

	require.NoError(t, err)
	assert.Equal(t, "commit456", sha)
}

func TestUpdateRef_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/refs/heads/feature" && r.Method == http.MethodPatch {
			var req updateRefRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			assert.Equal(t, "newsha123", req.SHA)
			assert.False(t, req.Force)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"ref": "refs/heads/feature",
				"sha": req.SHA,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	err := client.updateRef(context.Background(), "feature", "newsha123")

	require.NoError(t, err)
}

func TestUpdateRef_NotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	err := client.updateRef(context.Background(), "nonexistent-branch", "sha123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "github api error")
}

func TestCreatePullRequest_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/pulls" && r.Method == http.MethodPost {
			var req prRequestBody
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			assert.Equal(t, "Fix: Security issue", req.Title)
			assert.Equal(t, "Details of the fix", req.Body)
			assert.Equal(t, "fix-security", req.Head)
			assert.Equal(t, "main", req.Base)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(prResponse{
				ID:      123,
				Number:  42,
				HTMLURL: "https://github.com/owner/repo/pull/42",
				State:   "open",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Fix: Security issue",
		Body:       "Details of the fix",
		BranchName: "fix-security",
		CommitMsg:  "Fix security issue",
		Files:      map[string]string{},
	}

	result, err := client.createPullRequest(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, 42, result.Number)
	assert.Contains(t, result.URL, "/pull/42")
}

func TestDoRequest_WithContext(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := client.doRequest(ctx, http.MethodGet, mockServer.URL, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestDoRequest_ContextCancellation(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.doRequest(ctx, http.MethodGet, mockServer.URL, nil)

	assert.Error(t, err)
}

func TestDoRequest_WithBody(t *testing.T) {
	var receivedBody []byte

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	bodyData := []byte(`{"test": "data"}`)
	resp, err := client.doRequest(context.Background(), http.MethodPost, mockServer.URL, bodyData)

	require.NoError(t, err)
	assert.Equal(t, bodyData, receivedBody)
	resp.Body.Close()
}

func TestCommitFiles_Success(t *testing.T) {
	callCount := 0

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get base branch SHA
		if r.URL.Path == "/repos/owner/repo/git/ref/heads/main" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(refResponse{
				Object: struct {
					SHA string `json:"sha"`
				}{SHA: "base123"},
			})
			return
		}

		// Create tree
		if r.URL.Path == "/repos/owner/repo/git/trees" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(treeResponse{SHA: "tree123"})
			return
		}

		// Create commit
		if r.URL.Path == "/repos/owner/repo/git/commits" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(commitResponse{SHA: "commit123"})
			return
		}

		// Update ref
		if r.Method == http.MethodPatch {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"sha": "commit123"})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	files := map[string]string{
		"main.tf": "resource \"aws_instance\" \"example\" {}",
	}

	err := client.commitFiles(context.Background(), "feature", "Fix: Update terraform", files, "base123")

	require.NoError(t, err)
	_ = callCount // Use variable to avoid unused warning
}

func TestNewGitHubClient_HTTPClientConfig(t *testing.T) {
	client := NewGitHubClient("owner", "repo", "main", "token")

	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestCreatePR_MultipleFilesCommit(t *testing.T) {
	mockServer := createExtendedMockGitHubServer(t)
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "test-token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Multi-file fix",
		Body:       "This PR fixes multiple files",
		BranchName: "fix/multi-file-123",
		CommitMsg:  "Fix multiple resources",
		Files: map[string]string{
			"terraform/main.tf":      "resource \"aws_instance\" \"web\" {}",
			"terraform/outputs.tf":   "output \"instance_id\" {}",
			"terraform/variables.tf": "variable \"instance_type\" {}",
		},
	}

	result, err := client.CreatePR(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.Number, 0)
}

func TestCreatePR_CompleteWorkflow(t *testing.T) {
	mockServer := createExtendedMockGitHubServer(t)
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "test-token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Auto-remediation: Fix VPC configuration",
		Body:       "This PR auto-remediates VPC configuration drift detected by TFDrift-Falco",
		BranchName: "remediation/drift-aws_vpc-abc123de",
		CommitMsg:  "fix: auto-remediation for aws_vpc drift\n\nRemediation for VPC configuration mismatch",
		Files: map[string]string{
			"remediation/aws_vpc_vpc.tf": `
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "main-vpc"
  }
}`,
		},
	}

	result, err := client.CreatePR(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Number)
	assert.Contains(t, result.URL, "/pull/")
}

func TestCreateTree_ParseError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/trees" && r.Method == http.MethodPost {
			// Return invalid JSON
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`invalid json`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	_, err := client.createTree(context.Background(), map[string]string{"file.txt": "content"}, "base123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode tree response")
}

func TestCreateCommit_ParseError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/commits" && r.Method == http.MethodPost {
			// Return invalid JSON
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`invalid json`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	_, err := client.createCommit(context.Background(), "msg", "tree123", []string{"parent123"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode commit response")
}

func TestCreatePullRequest_ParseError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/pulls" && r.Method == http.MethodPost {
			// Return invalid JSON
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`invalid json`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Test",
		Body:       "Test",
		BranchName: "test",
		CommitMsg:  "Test",
		Files:      map[string]string{},
	}

	_, err := client.createPullRequest(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode pr response")
}

// Helper function for extended mock GitHub server
func createExtendedMockGitHubServer(t *testing.T) *httptest.Server {
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
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(treeResponse{SHA: "tree123"})
			return
		}

		// Create commit
		if r.URL.Path == "/repos/owner/repo/git/commits" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(commitResponse{SHA: "commit123"})
			return
		}

		// Update branch reference
		if r.Method == http.MethodPatch {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"sha": "commit123",
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
