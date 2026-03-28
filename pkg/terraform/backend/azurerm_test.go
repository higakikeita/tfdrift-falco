package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAzureRMBackend_ValidConfig(t *testing.T) {
	cfg := AzureRMBackendConfig{
		StorageAccountName: "mystorageaccount",
		ContainerName:      "tfstate",
		BlobName:           "terraform.tfstate",
		AccessKey:          "test-key",
	}

	b, err := NewAzureRMBackend(cfg)
	require.NoError(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, "azurerm", b.Name())
}

func TestNewAzureRMBackend_MissingStorageAccount(t *testing.T) {
	cfg := AzureRMBackendConfig{
		ContainerName: "tfstate",
		BlobName:      "terraform.tfstate",
	}

	_, err := NewAzureRMBackend(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage account name is required")
}

func TestNewAzureRMBackend_MissingContainer(t *testing.T) {
	cfg := AzureRMBackendConfig{
		StorageAccountName: "mystorageaccount",
		BlobName:           "terraform.tfstate",
	}

	_, err := NewAzureRMBackend(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container name is required")
}

func TestNewAzureRMBackend_MissingBlob(t *testing.T) {
	cfg := AzureRMBackendConfig{
		StorageAccountName: "mystorageaccount",
		ContainerName:      "tfstate",
	}

	_, err := NewAzureRMBackend(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blob name is required")
}

func TestAzureRMBackend_BuildBlobURL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      AzureRMBackendConfig
		expected string
	}{
		{
			name: "basic-url",
			cfg: AzureRMBackendConfig{
				StorageAccountName: "myaccount",
				ContainerName:      "tfstate",
				BlobName:           "terraform.tfstate",
			},
			expected: "https://myaccount.blob.core.windows.net/tfstate/terraform.tfstate",
		},
		{
			name: "with-sas-token",
			cfg: AzureRMBackendConfig{
				StorageAccountName: "myaccount",
				ContainerName:      "tfstate",
				BlobName:           "terraform.tfstate",
				SASToken:           "?sv=2020-10-02&sig=abc",
			},
			expected: "https://myaccount.blob.core.windows.net/tfstate/terraform.tfstate?sv=2020-10-02&sig=abc",
		},
		{
			name: "with-sas-token-no-question-mark",
			cfg: AzureRMBackendConfig{
				StorageAccountName: "myaccount",
				ContainerName:      "tfstate",
				BlobName:           "terraform.tfstate",
				SASToken:           "sv=2020-10-02&sig=abc",
			},
			expected: "https://myaccount.blob.core.windows.net/tfstate/terraform.tfstate?sv=2020-10-02&sig=abc",
		},
		{
			name: "nested-blob-path",
			cfg: AzureRMBackendConfig{
				StorageAccountName: "myaccount",
				ContainerName:      "tfstate",
				BlobName:           "env/prod/terraform.tfstate",
			},
			expected: "https://myaccount.blob.core.windows.net/tfstate/env%2Fprod%2Fterraform.tfstate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewAzureRMBackend(tt.cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, b.buildBlobURL())
		})
	}
}

func TestAzureRMBackend_Name(t *testing.T) {
	b, err := NewAzureRMBackend(AzureRMBackendConfig{
		StorageAccountName: "test",
		ContainerName:      "test",
		BlobName:           "test",
	})
	require.NoError(t, err)
	assert.Equal(t, "azurerm", b.Name())
}
