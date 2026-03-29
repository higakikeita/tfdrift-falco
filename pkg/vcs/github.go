// Package vcs provides version control system integrations for auto-remediation.
package vcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultAPIBase = "https://api.github.com"
	maxRetries     = 3
	retryDelay     = time.Second
)

// GitHubClient is a lightweight GitHub API client for PR creation.
type GitHubClient struct {
	owner      string
	repo       string
	baseBranch string
	token      string
	httpClient *http.Client
	apiBase    string
}

// PRRequest represents a pull request creation request.
type PRRequest struct {
	Title     string
	Body      string
	BranchName string
	Files     map[string]string // path -> content
	CommitMsg string
}

// PRResult represents the result of a PR creation.
type PRResult struct {
	URL    string
	Number int
}

// GitHub API response types
type refResponse struct {
	Object struct {
		SHA string `json:"sha"`
	} `json:"object"`
}

type createRefRequest struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

type treeItem struct {
	Path    string `json:"path"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type treeRequest struct {
	Tree    []treeItem `json:"tree"`
	BaseTree string    `json:"base_tree"`
}

type treeResponse struct {
	SHA string `json:"sha"`
}

type commitRequest struct {
	Message string   `json:"message"`
	Tree    string   `json:"tree"`
	Parents []string `json:"parents"`
}

type commitResponse struct {
	SHA string `json:"sha"`
}

type updateRefRequest struct {
	SHA   string `json:"sha"`
	Force bool   `json:"force"`
}

type prRequestBody struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

type prResponse struct {
	ID        int    `json:"id"`
	Number    int    `json:"number"`
	HTMLURL   string `json:"html_url"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient(owner, repo, baseBranch, token string) *GitHubClient {
	return &GitHubClient{
		owner:      owner,
		repo:       repo,
		baseBranch: baseBranch,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiBase:    defaultAPIBase,
	}
}

// CreatePR creates a new pull request with the specified files and changes.
// It follows this workflow:
// 1. Get the SHA of the base branch
// 2. Create a new branch
// 3. Create/update files in the new branch
// 4. Create a pull request
func (c *GitHubClient) CreatePR(ctx context.Context, req *PRRequest) (*PRResult, error) {
	if req == nil {
		return nil, fmt.Errorf("pr request cannot be nil")
	}

	if req.BranchName == "" {
		return nil, fmt.Errorf("branch name is required")
	}

	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if req.CommitMsg == "" {
		return nil, fmt.Errorf("commit message is required")
	}

	// Step 1: Get the SHA of the base branch
	baseSHA, err := c.getBaseBranchSHA(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get base branch SHA: %w", err)
	}

	// Step 2: Create the new branch
	if err := c.createBranch(ctx, req.BranchName, baseSHA); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Step 3: Commit files to the branch
	if len(req.Files) > 0 {
		if err := c.commitFiles(ctx, req.BranchName, req.CommitMsg, req.Files, baseSHA); err != nil {
			return nil, fmt.Errorf("failed to commit files: %w", err)
		}
	}

	// Step 4: Create the pull request
	prResult, err := c.createPullRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return prResult, nil
}

// getBaseBranchSHA retrieves the commit SHA of the base branch.
func (c *GitHubClient) getBaseBranchSHA(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/ref/heads/%s", c.apiBase, c.owner, c.repo, c.baseBranch)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github api error: %d - %s", resp.StatusCode, string(body))
	}

	var refResp refResponse
	if err := json.NewDecoder(resp.Body).Decode(&refResp); err != nil {
		return "", fmt.Errorf("failed to decode ref response: %w", err)
	}

	return refResp.Object.SHA, nil
}

// createBranch creates a new branch pointing to the given SHA.
func (c *GitHubClient) createBranch(ctx context.Context, branchName, sha string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs", c.apiBase, c.owner, c.repo)

	createReq := createRefRequest{
		Ref: fmt.Sprintf("refs/heads/%s", branchName),
		SHA: sha,
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyData, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api error: %d - %s", resp.StatusCode, string(bodyData))
	}

	return nil
}

// commitFiles creates a new commit with the provided files and updates the branch reference.
func (c *GitHubClient) commitFiles(ctx context.Context, branchName, commitMsg string, files map[string]string, baseTreeSHA string) error {
	// Step 1: Create a tree with the file changes
	treeSHA, err := c.createTree(ctx, files, baseTreeSHA)
	if err != nil {
		return fmt.Errorf("failed to create tree: %w", err)
	}

	// Step 2: Get the current commit SHA of the branch
	branchSHA, err := c.getBaseBranchSHA(ctx)
	if err != nil {
		return fmt.Errorf("failed to get branch SHA for commit: %w", err)
	}

	// Step 3: Create a new commit
	commitSHA, err := c.createCommit(ctx, commitMsg, treeSHA, []string{branchSHA})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Step 4: Update the branch reference to point to the new commit
	if err := c.updateRef(ctx, branchName, commitSHA); err != nil {
		return fmt.Errorf("failed to update branch reference: %w", err)
	}

	return nil
}

// createTree creates a new tree object with the provided files.
func (c *GitHubClient) createTree(ctx context.Context, files map[string]string, baseTreeSHA string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/trees", c.apiBase, c.owner, c.repo)

	// Convert files to tree items
	treeItems := make([]treeItem, 0, len(files))
	for path, content := range files {
		treeItems = append(treeItems, treeItem{
			Path:    path,
			Mode:    "100644",
			Type:    "blob",
			Content: content,
		})
	}

	treeReq := treeRequest{
		Tree:     treeItems,
		BaseTree: baseTreeSHA,
	}

	body, err := json.Marshal(treeReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tree request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyData, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github api error: %d - %s", resp.StatusCode, string(bodyData))
	}

	var treeResp treeResponse
	if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
		return "", fmt.Errorf("failed to decode tree response: %w", err)
	}

	return treeResp.SHA, nil
}

// createCommit creates a new commit with the given tree and parents.
func (c *GitHubClient) createCommit(ctx context.Context, message, treeSHA string, parents []string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/commits", c.apiBase, c.owner, c.repo)

	commitReq := commitRequest{
		Message: message,
		Tree:    treeSHA,
		Parents: parents,
	}

	body, err := json.Marshal(commitReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal commit request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyData, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github api error: %d - %s", resp.StatusCode, string(bodyData))
	}

	var commitResp commitResponse
	if err := json.NewDecoder(resp.Body).Decode(&commitResp); err != nil {
		return "", fmt.Errorf("failed to decode commit response: %w", err)
	}

	return commitResp.SHA, nil
}

// updateRef updates a branch reference to point to a new commit.
func (c *GitHubClient) updateRef(ctx context.Context, branchName, sha string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs/heads/%s", c.apiBase, c.owner, c.repo, branchName)

	updateReq := updateRefRequest{
		SHA:   sha,
		Force: false,
	}

	body, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPatch, url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyData, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api error: %d - %s", resp.StatusCode, string(bodyData))
	}

	return nil
}

// createPullRequest creates a new pull request.
func (c *GitHubClient) createPullRequest(ctx context.Context, req *PRRequest) (*PRResult, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", c.apiBase, c.owner, c.repo)

	prReq := prRequestBody{
		Title: req.Title,
		Body:  req.Body,
		Head:  req.BranchName,
		Base:  c.baseBranch,
	}

	body, err := json.Marshal(prReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pr request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyData, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error: %d - %s", resp.StatusCode, string(bodyData))
	}

	var prResp prResponse
	if err := json.NewDecoder(resp.Body).Decode(&prResp); err != nil {
		return nil, fmt.Errorf("failed to decode pr response: %w", err)
	}

	return &PRResult{
		URL:    prResp.HTMLURL,
		Number: prResp.Number,
	}, nil
}

// doRequest performs an HTTP request with proper authentication and error handling.
func (c *GitHubClient) doRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	return resp, nil
}
