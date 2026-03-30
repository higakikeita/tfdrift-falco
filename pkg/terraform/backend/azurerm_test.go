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

func TestAzureRMBackend_BuildBlobURL_PathEscaping(t *testing.T) {
	tests := []struct {
		name     string
		blobName string
		expected string
	}{
		{
			name:     "blob with special chars",
			blobName: "path/to/terraform.tfstate",
			expected: "https://myaccount.blob.core.windows.net/tfstate/path%2Fto%2Fterraform.tfstate",
		},
		{
			name:     "blob with spaces",
			blobName: "my file.tfstate",
			expected: "https://myaccount.blob.core.windows.net/tfstate/my%20file.tfstate",
		},
		{
			name:     "blob with unicode",
			blobName: "terraform-日本語.tfstate",
			expected: "https://myaccount.blob.core.windows.net/tfstate/terraform-%E6%97%A5%E6%9C%AC%E8%AA%9E.tfstate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewAzureRMBackend(AzureRMBackendConfig{
				StorageAccountName: "myaccount",
				ContainerName:      "tfstate",
				BlobName:           tt.blobName,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, b.buildBlobURL())
		})
	}
}

func TestAzureRMBackend_ConfigAssignment(t *testing.T) {
	tests := []struct {
		name               string
		cfg                AzureRMBackendConfig
		expectedAccount    string
		expectedContainer  string
		expectedBlob       string
		expectedAccessKey  string
		expectedSASToken   string
	}{
		{
			name: "with access key",
			cfg: AzureRMBackendConfig{
				StorageAccountName: "myaccount",
				ContainerName:      "tfstate",
				BlobName:           "terraform.tfstate",
				AccessKey:          "mykey123",
			},
			expectedAccount:   "myaccount",
			expectedContainer: "tfstate",
			expectedBlob:      "terraform.tfstate",
			expectedAccessKey: "mykey123",
		},
		{
			name: "with SAS token",
			cfg: AzureRMBackendConfig{
				StorageAccountName: "myaccount",
				ContainerName:      "tfstate",
				BlobName:           "terraform.tfstate",
				SASToken:           "sv=2020-10-02&sig=abc",
			},
			expectedAccount:   "myaccount",
			expectedContainer: "tfstate",
			expectedBlob:      "terraform.tfstate",
			expectedSASToken:  "sv=2020-10-02&sig=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewAzureRMBackend(tt.cfg)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedAccount, b.storageAccountName)
			assert.Equal(t, tt.expectedContainer, b.containerName)
			assert.Equal(t, tt.expectedBlob, b.blobName)
			assert.Equal(t, tt.expectedAccessKey, b.accessKey)
			assert.Equal(t, tt.expectedSASToken, b.sasToken)
			assert.NotNil(t, b.httpClient)
		})
	}
}

func TestAzureRMBackend_ImplementsBackendInterface(t *testing.T) {
	var _ Backend = (*AzureRMBackend)(nil)

	b, err := NewAzureRMBackend(AzureRMBackendConfig{
		StorageAccountName: "test",
		ContainerName:      "test",
		BlobName:           "test",
	})
	require.NoError(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, "azurerm", b.Name())
}

func TestAzureRMBackend_BuildBlobURL_SASTokenFormats(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "token with question mark prefix",
			token:    "?sv=2020-10-02&sig=abc",
			expected: "https://test.blob.core.windows.net/cont/blob?sv=2020-10-02&sig=abc",
		},
		{
			name:     "token without question mark",
			token:    "sv=2020-10-02&sig=abc",
			expected: "https://test.blob.core.windows.net/cont/blob?sv=2020-10-02&sig=abc",
		},
		{
			name:     "empty token",
			token:    "",
			expected: "https://test.blob.core.windows.net/cont/blob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewAzureRMBackend(AzureRMBackendConfig{
				StorageAccountName: "test",
				ContainerName:      "cont",
				BlobName:           "blob",
				SASToken:           tt.token,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, b.buildBlobURL())
		})
	}
}

// Test AzureRMBackend error handling
func TestAzureRMBackend_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		config        AzureRMBackendConfig
		expectedError string
	}{
		{
			name: "Missing storage account",
			config: AzureRMBackendConfig{
				ContainerName: "container",
				BlobName:      "blob",
			},
			expectedError: "storage account name is required",
		},
		{
			name: "Missing container",
			config: AzureRMBackendConfig{
				StorageAccountName: "account",
				BlobName:           "blob",
			},
			expectedError: "container name is required",
		},
		{
			name: "Missing blob",
			config: AzureRMBackendConfig{
				StorageAccountName: "account",
				ContainerName:      "container",
			},
			expectedError: "blob name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewAzureRMBackend(tt.config)

			assert.Error(t, err)
			assert.Nil(t, backend)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// Test AzureRMBackend Name method
func TestAzureRMBackend_NameMethod(t *testing.T) {
	b, err := NewAzureRMBackend(AzureRMBackendConfig{
		StorageAccountName: "test",
		ContainerName:      "test",
		BlobName:           "test",
	})
	require.NoError(t, err)

	assert.Equal(t, "azurerm", b.Name())
	// Name should always return the same value
	assert.Equal(t, "azurerm", b.Name())
}

// Test AzureRMBackend HTTPClient initialization
func TestAzureRMBackend_HTTPClientInitialized(t *testing.T) {
	b, err := NewAzureRMBackend(AzureRMBackendConfig{
		StorageAccountName: "test",
		ContainerName:      "test",
		BlobName:           "test",
	})
	require.NoError(t, err)

	assert.NotNil(t, b.httpClient)
}
