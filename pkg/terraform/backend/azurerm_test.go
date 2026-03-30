package backend

import (
	"context"
	"io"
	"net/http"
	"strings"
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

// Mock HTTP Transport for testing
type MockTransport struct {
	StatusCode int
	Body       string
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.StatusCode,
		Body:       io.NopCloser(strings.NewReader(m.Body)),
		Header:     make(http.Header),
	}, nil
}

// Test AzureRMBackend Load with mock HTTP transport
func TestAzureRMBackend_Load_NotFound(t *testing.T) {
	b := &AzureRMBackend{
		storageAccountName: "test",
		containerName:      "tfstate",
		blobName:           "notfound.tfstate",
		httpClient: &http.Client{
			Transport: &MockTransport{
				StatusCode: http.StatusNotFound,
				Body:       `{"error": "blob not found"}`,
			},
		},
	}

	ctx := context.Background()

	data, err := b.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "404")
}

func TestAzureRMBackend_Load_ServerError(t *testing.T) {
	b := &AzureRMBackend{
		storageAccountName: "test",
		containerName:      "tfstate",
		blobName:           "terraform.tfstate",
		httpClient: &http.Client{
			Transport: &MockTransport{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"error": "internal server error"}`,
			},
		},
	}

	ctx := context.Background()

	data, err := b.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "500")
}

func TestAzureRMBackend_Load_Success(t *testing.T) {
	stateContent := `{"version": 4, "resources": []}`

	b := &AzureRMBackend{
		storageAccountName: "test",
		containerName:      "tfstate",
		blobName:           "terraform.tfstate",
		httpClient: &http.Client{
			Transport: &MockTransport{
				StatusCode: http.StatusOK,
				Body:       stateContent,
			},
		},
	}

	ctx := context.Background()

	data, err := b.Load(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, []byte(stateContent), data)
}

func TestAzureRMBackend_Load_EmptyResponse(t *testing.T) {
	b := &AzureRMBackend{
		storageAccountName: "test",
		containerName:      "tfstate",
		blobName:           "terraform.tfstate",
		httpClient: &http.Client{
			Transport: &MockTransport{
				StatusCode: http.StatusOK,
				Body:       "",
			},
		},
	}

	ctx := context.Background()

	data, err := b.Load(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 0, len(data))
}

func TestAzureRMBackend_Load_LargeBody(t *testing.T) {
	largeBody := `{"version": 4, "resources": [` + strings.Repeat(`{"id":"resource"},`, 1000) + `]}`

	b := &AzureRMBackend{
		storageAccountName: "test",
		containerName:      "tfstate",
		blobName:           "terraform.tfstate",
		httpClient: &http.Client{
			Transport: &MockTransport{
				StatusCode: http.StatusOK,
				Body:       largeBody,
			},
		},
	}

	ctx := context.Background()

	data, err := b.Load(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 1000)
}

func TestAzureRMBackend_Load_WithAccessKey(t *testing.T) {
	b, err := NewAzureRMBackend(AzureRMBackendConfig{
		StorageAccountName: "myaccount",
		ContainerName:      "tfstate",
		BlobName:           "terraform.tfstate",
		AccessKey:          "myaccesskey123",
	})
	require.NoError(t, err)

	// Verify accessKey was stored
	assert.Equal(t, "myaccesskey123", b.accessKey)
	assert.Equal(t, "", b.sasToken)
}

func TestAzureRMBackend_Name_Consistency(t *testing.T) {
	b, err := NewAzureRMBackend(AzureRMBackendConfig{
		StorageAccountName: "test",
		ContainerName:      "test",
		BlobName:           "test",
	})
	require.NoError(t, err)

	// Name should always return the same value
	assert.Equal(t, "azurerm", b.Name())
	assert.Equal(t, "azurerm", b.Name())
	assert.Equal(t, "azurerm", b.Name())
}

func TestAzureRMBackend_BuildBlobURL_NoSASToken(t *testing.T) {
	b, err := NewAzureRMBackend(AzureRMBackendConfig{
		StorageAccountName: "storageaccount",
		ContainerName:      "container",
		BlobName:           "blob.tfstate",
		AccessKey:          "somekey",
	})
	require.NoError(t, err)

	url := b.buildBlobURL()
	// Should not contain SAS token
	assert.NotContains(t, url, "?sv=")
	assert.NotContains(t, url, "&sig=")
	assert.Contains(t, url, "https://")
	assert.Contains(t, url, "storageaccount.blob.core.windows.net")
}

func TestAzureRMBackend_Load_UnmarshalErrorInResponse(t *testing.T) {
	// This tests the error handler when response body is returned with non-200 status
	b := &AzureRMBackend{
		storageAccountName: "test",
		containerName:      "tfstate",
		blobName:           "terraform.tfstate",
		httpClient: &http.Client{
			Transport: &MockTransport{
				StatusCode: http.StatusUnauthorized,
				Body:       `{"error": {"code": "AuthenticationFailed", "message": "Authentication failed"}}`,
			},
		},
	}

	ctx := context.Background()

	data, err := b.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "401")
}
