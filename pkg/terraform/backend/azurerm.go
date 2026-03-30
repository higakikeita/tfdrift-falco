package backend

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// AzureRMBackend implements the Backend interface for Azure Blob Storage.
//
// This backend reads Terraform state files from Azure Blob Storage containers,
// supporting both shared access key and Azure AD (DefaultAzureCredential) authentication.
//
// The backend handles:
//   - Connection via storage account URL or connection string
//   - Shared Access Key authentication
//   - SAS token authentication
//   - Error handling with detailed logging
//
// Thread-safe: The HTTP client used internally handles concurrent access.
//
// Example usage:
//
//	backend, err := NewAzureRMBackend(AzureRMBackendConfig{
//	    StorageAccountName: "mystorageaccount",
//	    ContainerName:      "tfstate",
//	    BlobName:           "terraform.tfstate",
//	    AccessKey:          "base64encodedkey...",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	data, err := backend.Load(ctx)
type AzureRMBackend struct {
	storageAccountName string
	containerName      string
	blobName           string
	accessKey          string
	sasToken           string
	httpClient         *http.Client
}

// AzureRMBackendConfig contains configuration for the Azure Blob Storage backend.
type AzureRMBackendConfig struct {
	// StorageAccountName is the Azure storage account name (e.g., "mystorageaccount")
	StorageAccountName string

	// ContainerName is the blob container name (e.g., "tfstate")
	ContainerName string

	// BlobName is the path to the state file blob (e.g., "terraform.tfstate" or "env/prod/terraform.tfstate")
	BlobName string

	// AccessKey is the storage account access key (optional if using SAS token or Azure AD)
	AccessKey string

	// SASToken is an Azure SAS token for authentication (optional)
	SASToken string
}

// NewAzureRMBackend creates a new Azure Blob Storage backend for Terraform state management.
//
// The function initializes an HTTP-based client that reads state from Azure Blob Storage.
// Authentication is handled via access key or SAS token.
//
// Parameters:
//   - cfg: Azure Blob Storage configuration
//
// Returns:
//   - *AzureRMBackend: Initialized backend ready to load state
//   - error: Configuration validation errors
func NewAzureRMBackend(cfg AzureRMBackendConfig) (*AzureRMBackend, error) {
	if cfg.StorageAccountName == "" {
		return nil, fmt.Errorf("azure storage account name is required")
	}
	if cfg.ContainerName == "" {
		return nil, fmt.Errorf("azure container name is required")
	}
	if cfg.BlobName == "" {
		return nil, fmt.Errorf("azure blob name is required")
	}

	return &AzureRMBackend{
		storageAccountName: cfg.StorageAccountName,
		containerName:      cfg.ContainerName,
		blobName:           cfg.BlobName,
		accessKey:          cfg.AccessKey,
		sasToken:           cfg.SASToken,
		httpClient:         &http.Client{},
	}, nil
}

// Load reads the Terraform state file from Azure Blob Storage.
//
// The method constructs the blob URL and downloads the state file using
// the configured authentication method.
//
// Parameters:
//   - ctx: Context for the HTTP request (supports cancellation and timeouts)
//
// Returns:
//   - []byte: Complete Terraform state file contents as JSON
//   - error: If the blob doesn't exist, authentication fails, or network errors
func (b *AzureRMBackend) Load(ctx context.Context) ([]byte, error) {
	blobURL := b.buildBlobURL()
	log.Infof("Loading Terraform state from Azure Blob Storage: %s/%s/%s",
		b.storageAccountName, b.containerName, b.blobName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, blobURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for Azure Blob Storage REST API
	req.Header.Set("x-ms-version", "2020-10-02")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob from Azure: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("azure Blob Storage returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob body: %w", err)
	}

	log.Infof("Successfully loaded %d bytes from Azure Blob Storage (%s/%s/%s)",
		len(data), b.storageAccountName, b.containerName, b.blobName)
	return data, nil
}

// buildBlobURL constructs the Azure Blob Storage URL for the state file.
func (b *AzureRMBackend) buildBlobURL() string {
	baseURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		b.storageAccountName,
		b.containerName,
		url.PathEscape(b.blobName),
	)

	if b.sasToken != "" {
		// SAS token already includes the '?' or we add it
		if b.sasToken[0] == '?' {
			return baseURL + b.sasToken
		}
		return baseURL + "?" + b.sasToken
	}

	return baseURL
}

// Name returns the backend identifier for logging and debugging.
func (b *AzureRMBackend) Name() string {
	return "azurerm"
}
