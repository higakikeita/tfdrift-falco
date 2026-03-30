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

// TestGetBaseBranchSHA_DecodingError tests handling of invalid JSON response
func TestGetBaseBranchSHA_DecodingError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	sha, err := client.getBaseBranchSHA(context.Background())

	assert.Error(t, err)
	assert.Empty(t, sha)
	assert.Contains(t, err.Error(), "failed to decode ref response")
}

// TestCreateBranch_MarshalError tests handling of request marshaling
func TestCreateBranch_RequestConstruction(t *testing.T) {
	capturedRequest := ""
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/refs" && r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			capturedRequest = string(body)

			var req createRefRequest
			json.Unmarshal(body, &req)

			assert.Equal(t, "refs/heads/newbranch", req.Ref)
			assert.Equal(t, "sha123", req.SHA)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	err := client.createBranch(context.Background(), "newbranch", "sha123")

	require.NoError(t, err)
	assert.Contains(t, capturedRequest, "newbranch")
	assert.Contains(t, capturedRequest, "sha123")
}

// TestCreateTree_RequestConstruction tests tree creation request structure
func TestCreateTree_RequestConstruction(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/trees" && r.Method == http.MethodPost {
			var req treeRequest
			json.NewDecoder(r.Body).Decode(&req)

			assert.Equal(t, "baseSHA", req.BaseTree)
			assert.Len(t, req.Tree, 1)
			assert.Equal(t, "file.txt", req.Tree[0].Path)
			assert.Equal(t, "content here", req.Tree[0].Content)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(treeResponse{SHA: "treeSHA"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	files := map[string]string{
		"file.txt": "content here",
	}

	sha, err := client.createTree(context.Background(), files, "baseSHA")

	require.NoError(t, err)
	assert.Equal(t, "treeSHA", sha)
}

// TestCreateCommit_RequestConstruction tests commit creation request structure
func TestCreateCommit_RequestConstruction(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/commits" && r.Method == http.MethodPost {
			var req commitRequest
			json.NewDecoder(r.Body).Decode(&req)

			assert.Equal(t, "Commit message", req.Message)
			assert.Equal(t, "treeSHA", req.Tree)
			assert.Len(t, req.Parents, 2)
			assert.Equal(t, "parent1", req.Parents[0])
			assert.Equal(t, "parent2", req.Parents[1])

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(commitResponse{SHA: "commitSHA"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	sha, err := client.createCommit(context.Background(), "Commit message", "treeSHA", []string{"parent1", "parent2"})

	require.NoError(t, err)
	assert.Equal(t, "commitSHA", sha)
}

// TestCreatePullRequest_RequestConstruction tests PR creation request structure
func TestCreatePullRequest_RequestConstruction(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/pulls" && r.Method == http.MethodPost {
			var req prRequestBody
			json.NewDecoder(r.Body).Decode(&req)

			assert.Equal(t, "PR Title", req.Title)
			assert.Equal(t, "PR Body", req.Body)
			assert.Equal(t, "feature-branch", req.Head)
			assert.Equal(t, "main", req.Base)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(prResponse{
				ID:      123,
				Number:  10,
				HTMLURL: "https://github.com/owner/repo/pull/10",
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
		Title:      "PR Title",
		Body:       "PR Body",
		BranchName: "feature-branch",
		CommitMsg:  "message",
		Files:      map[string]string{},
	}

	result, err := client.createPullRequest(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, 10, result.Number)
}

// TestUpdateRef_RequestConstruction tests ref update request structure
func TestUpdateRef_RequestConstruction(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/git/refs/heads/main" && r.Method == http.MethodPatch {
			var req updateRefRequest
			json.NewDecoder(r.Body).Decode(&req)

			assert.Equal(t, "newSHA", req.SHA)
			assert.False(t, req.Force)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	err := client.updateRef(context.Background(), "main", "newSHA")

	require.NoError(t, err)
}

// TestCommitFiles_CompleteFlow tests the complete file commit flow
func TestCommitFiles_CompleteFlow(t *testing.T) {
	requestCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		requestCount++

		if r.URL.Path == "/repos/owner/repo/git/ref/heads/main" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(refResponse{
				Object: struct {
					SHA string `json:"sha"`
				}{SHA: "baseSHA"},
			})
			return
		}

		if r.URL.Path == "/repos/owner/repo/git/trees" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(treeResponse{SHA: "treeSHA"})
			return
		}

		if r.URL.Path == "/repos/owner/repo/git/commits" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(commitResponse{SHA: "commitSHA"})
			return
		}

		if r.Method == http.MethodPatch {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	files := map[string]string{
		"terraform/main.tf": "resource \"aws_instance\" \"example\" {}",
	}

	err := client.commitFiles(context.Background(), "feature", "Commit message", files, "baseSHA")

	require.NoError(t, err)
	assert.Greater(t, requestCount, 0)
}

// TestCreatePR_FailedTreeCreation tests PR creation failure at tree creation
func TestCreatePR_FailedTreeCreation(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/repos/owner/repo/git/ref/heads/main" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(refResponse{
				Object: struct {
					SHA string `json:"sha"`
				}{SHA: "baseSHA"},
			})
			return
		}

		if r.URL.Path == "/repos/owner/repo/git/refs" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}

		if r.URL.Path == "/repos/owner/repo/git/trees" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Server error"}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewGitHubClient("owner", "repo", "main", "token")
	client.apiBase = mockServer.URL

	req := &PRRequest{
		Title:      "Test PR",
		Body:       "Test",
		BranchName: "test-branch",
		CommitMsg:  "Test commit",
		Files: map[string]string{
			"file.txt": "content",
		},
	}

	result, err := client.CreatePR(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to commit files")
}

// TestNewGitHubClient_Configuration tests client initialization
func TestNewGitHubClient_Configuration(t *testing.T) {
	client := NewGitHubClient("myowner", "myrepo", "develop", "mytoken")

	assert.Equal(t, "myowner", client.owner)
	assert.Equal(t, "myrepo", client.repo)
	assert.Equal(t, "develop", client.baseBranch)
	assert.Equal(t, "mytoken", client.token)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, defaultAPIBase, client.apiBase)
}

// TestCreatePR_NoFilesValidation tests PR creation with empty files but valid metadata
func TestCreatePR_NoFilesValidation(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/repos/owner/repo/git/ref/heads/main" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(refResponse{
				Object: struct {
					SHA string `json:"sha"`
				}{SHA: "baseSHA"},
			})
			return
		}

		if r.URL.Path == "/repos/owner/repo/git/refs" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}

		if r.URL.Path == "/repos/owner/repo/pulls" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(prResponse{
				ID:      1,
				Number:  1,
				HTMLURL: "https://github.com/owner/repo/pull/1",
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
		Title:      "Documentation Update",
		Body:       "No code changes",
		BranchName: "docs",
		CommitMsg:  "Update docs",
		Files:      map[string]string{}, // Empty files
	}

	result, err := client.CreatePR(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Number)
}
