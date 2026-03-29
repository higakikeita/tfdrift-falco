package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== APIResponse Tests ====================

func TestAPIResponse_SuccessResponse(t *testing.T) {
	resp := &APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":   1,
			"name": "test",
		},
	}

	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Nil(t, resp.Error)
}

func TestAPIResponse_ErrorResponse(t *testing.T) {
	resp := &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    400,
			Message: "Bad Request",
			Details: "Invalid input",
		},
	}

	assert.False(t, resp.Success)
	assert.Nil(t, resp.Data)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, 400, resp.Error.Code)
}

func TestAPIResponse_Marshaling(t *testing.T) {
	resp := &APIResponse{
		Success: true,
		Data: map[string]string{
			"message": "success",
		},
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "success")
	assert.Contains(t, string(data), "message")
}

func TestAPIResponse_Unmarshaling(t *testing.T) {
	jsonData := []byte(`{"success":true,"data":{"id":1},"error":null}`)

	var resp APIResponse
	err := json.Unmarshal(jsonData, &resp)

	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestAPIError_WithDetails(t *testing.T) {
	err := &APIError{
		Code:    500,
		Message: "Internal Server Error",
		Details: "Database connection failed",
	}

	assert.Equal(t, 500, err.Code)
	assert.Equal(t, "Internal Server Error", err.Message)
	assert.Equal(t, "Database connection failed", err.Details)
}

func TestAPIError_WithoutDetails(t *testing.T) {
	err := &APIError{
		Code:    404,
		Message: "Not Found",
	}

	assert.Equal(t, 404, err.Code)
	assert.Equal(t, "Not Found", err.Message)
	assert.Equal(t, "", err.Details)
}

func TestAPIResponse_JSONOmitEmpty(t *testing.T) {
	resp := &APIResponse{
		Success: true,
		Data:    nil,
		Error:   nil,
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	// error and data should be omitted
	jsonStr := string(data)
	assert.NotContains(t, jsonStr, "error")
	assert.NotContains(t, jsonStr, "data")
}

// ==================== HealthResponse Tests ====================

func TestHealthResponse_Fields(t *testing.T) {
	health := &HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: "2024-01-01T00:00:00Z",
	}

	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "1.0.0", health.Version)
	assert.NotEmpty(t, health.Timestamp)
}

func TestHealthResponse_Marshaling(t *testing.T) {
	health := &HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(health)
	assert.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "healthy")
	assert.Contains(t, jsonStr, "1.0.0")
}

// ==================== PaginationParams Tests ====================

func TestNewPaginationParams_Valid(t *testing.T) {
	params := NewPaginationParams(2, 50)

	assert.Equal(t, 2, params.Page)
	assert.Equal(t, 50, params.Limit)
}

func TestNewPaginationParams_DefaultPage(t *testing.T) {
	params := NewPaginationParams(0, 50)

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 50, params.Limit)
}

func TestNewPaginationParams_NegativePage(t *testing.T) {
	params := NewPaginationParams(-1, 50)

	assert.Equal(t, 1, params.Page)
}

func TestNewPaginationParams_DefaultLimit(t *testing.T) {
	params := NewPaginationParams(1, 0)

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 50, params.Limit)
}

func TestNewPaginationParams_NegativeLimit(t *testing.T) {
	params := NewPaginationParams(1, -1)

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 50, params.Limit)
}

func TestNewPaginationParams_ExceedsMaxLimit(t *testing.T) {
	params := NewPaginationParams(1, 2000)

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 1000, params.Limit) // Should be capped at 1000
}

func TestNewPaginationParams_AtMaxLimit(t *testing.T) {
	params := NewPaginationParams(1, 1000)

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 1000, params.Limit)
}

func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		expectedOffset int
	}{
		{"Page 1", 1, 50, 0},
		{"Page 2", 2, 50, 50},
		{"Page 3", 3, 50, 100},
		{"Page 1 Limit 10", 1, 10, 0},
		{"Page 5 Limit 20", 5, 20, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := NewPaginationParams(tt.page, tt.limit)
			assert.Equal(t, tt.expectedOffset, params.Offset())
		})
	}
}

// ==================== PaginatedResponse Tests ====================

func TestPaginatedResponse_Basic(t *testing.T) {
	data := []map[string]interface{}{
		{"id": 1, "name": "Item 1"},
		{"id": 2, "name": "Item 2"},
	}

	resp := &PaginatedResponse{
		Data:       data,
		Page:       1,
		Limit:      50,
		Total:      100,
		TotalPages: 2,
	}

	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 50, resp.Limit)
	assert.Equal(t, 100, resp.Total)
	assert.Equal(t, 2, resp.TotalPages)
}

func TestPaginatedResponse_Marshaling(t *testing.T) {
	data := []interface{}{map[string]string{"id": "1"}}

	resp := &PaginatedResponse{
		Data:       data,
		Page:       1,
		Limit:      10,
		Total:      1,
		TotalPages: 1,
	}

	jsonData, err := json.Marshal(resp)
	assert.NoError(t, err)

	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, `"page":1`)
	assert.Contains(t, jsonStr, `"limit":10`)
	assert.Contains(t, jsonStr, `"total":1`)
	assert.Contains(t, jsonStr, `"total_pages":1`)
}

func TestPaginatedResponse_Unmarshaling(t *testing.T) {
	jsonData := []byte(`{"data":[{"id":"1"}],"page":1,"limit":10,"total":1,"total_pages":1}`)

	var resp PaginatedResponse
	err := json.Unmarshal(jsonData, &resp)

	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.Limit)
}

// ==================== CalculateTotalPages Tests ====================

func TestCalculateTotalPages_EvenDivision(t *testing.T) {
	pages := CalculateTotalPages(100, 10)
	assert.Equal(t, 10, pages)
}

func TestCalculateTotalPages_WithRemainder(t *testing.T) {
	pages := CalculateTotalPages(105, 10)
	assert.Equal(t, 11, pages)
}

func TestCalculateTotalPages_LessThanLimit(t *testing.T) {
	pages := CalculateTotalPages(5, 10)
	assert.Equal(t, 1, pages)
}

func TestCalculateTotalPages_ExactMatch(t *testing.T) {
	pages := CalculateTotalPages(50, 50)
	assert.Equal(t, 1, pages)
}

func TestCalculateTotalPages_ZeroTotal(t *testing.T) {
	pages := CalculateTotalPages(0, 10)
	assert.Equal(t, 0, pages)
}

func TestCalculateTotalPages_ZeroLimit(t *testing.T) {
	pages := CalculateTotalPages(100, 0)
	assert.Equal(t, 0, pages)
}

func TestCalculateTotalPages_ZeroBoth(t *testing.T) {
	pages := CalculateTotalPages(0, 0)
	assert.Equal(t, 0, pages)
}

func TestCalculateTotalPages_LargeNumbers(t *testing.T) {
	pages := CalculateTotalPages(1000000, 100)
	assert.Equal(t, 10000, pages)
}

// ==================== Graph Model Tests ====================

func TestCytoscapeNode_Valid(t *testing.T) {
	node := &CytoscapeNode{
		Data: NodeData{
			ID:           "node-1",
			Label:        "Resource",
			Type:         "aws_instance",
			ResourceType: "EC2",
			ResourceName: "production-server",
			Severity:     "high",
			Metadata: map[string]interface{}{
				"region": "us-east-1",
			},
		},
	}

	assert.Equal(t, "node-1", node.Data.ID)
	assert.Equal(t, "Resource", node.Data.Label)
	assert.Equal(t, "aws_instance", node.Data.Type)
	assert.NotNil(t, node.Data.Metadata)
}

func TestNodeData_Marshaling(t *testing.T) {
	data := NodeData{
		ID:           "node-1",
		Label:        "Test",
		Type:         "aws_s3",
		ResourceType: "S3",
		Metadata: map[string]interface{}{
			"region": "us-west-2",
		},
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)

	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "node-1")
	assert.Contains(t, jsonStr, "aws_s3")
}

func TestCytoscapeEdge_Valid(t *testing.T) {
	edge := &CytoscapeEdge{
		Data: EdgeData{
			ID:           "edge-1",
			Source:       "node-1",
			Target:       "node-2",
			Label:        "depends_on",
			Type:         "dependency",
			Relationship: "requires",
		},
	}

	assert.Equal(t, "edge-1", edge.Data.ID)
	assert.Equal(t, "node-1", edge.Data.Source)
	assert.Equal(t, "node-2", edge.Data.Target)
	assert.Equal(t, "depends_on", edge.Data.Label)
}

func TestEdgeData_Marshaling(t *testing.T) {
	data := EdgeData{
		ID:           "edge-1",
		Source:       "node-1",
		Target:       "node-2",
		Type:         "reference",
		Relationship: "references",
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)

	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, "edge-1")
	assert.Contains(t, jsonStr, "node-1")
}

func TestCytoscapeElements_Valid(t *testing.T) {
	elements := &CytoscapeElements{
		Nodes: []CytoscapeNode{
			{
				Data: NodeData{
					ID:   "node-1",
					Type: "aws_instance",
				},
			},
		},
		Edges: []CytoscapeEdge{
			{
				Data: EdgeData{
					ID:     "edge-1",
					Source: "node-1",
					Target: "node-2",
				},
			},
		},
	}

	assert.Equal(t, 1, len(elements.Nodes))
	assert.Equal(t, 1, len(elements.Edges))
	assert.Equal(t, "node-1", elements.Nodes[0].Data.ID)
}

func TestCytoscapeElements_Marshaling(t *testing.T) {
	elements := &CytoscapeElements{
		Nodes: []CytoscapeNode{
			{
				Data: NodeData{
					ID:   "node-1",
					Type: "instance",
				},
			},
		},
		Edges: []CytoscapeEdge{
			{
				Data: EdgeData{
					ID:     "edge-1",
					Source: "node-1",
					Target: "node-2",
				},
			},
		},
	}

	jsonData, err := json.Marshal(elements)
	assert.NoError(t, err)

	var unmarshaled CytoscapeElements
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)

	require.Equal(t, 1, len(unmarshaled.Nodes))
	assert.Equal(t, "node-1", unmarshaled.Nodes[0].Data.ID)
}

func TestNodeData_WithMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"region":  "us-east-1",
		"zone":    "us-east-1a",
		"tags":    []string{"prod", "critical"},
		"version": 2,
	}

	data := NodeData{
		ID:       "node-1",
		Label:    "Test",
		Type:     "aws_instance",
		Metadata: metadata,
	}

	assert.Equal(t, "us-east-1", data.Metadata["region"])
	assert.Equal(t, "us-east-1a", data.Metadata["zone"])
}

func TestEdgeData_WithoutOptionalFields(t *testing.T) {
	edge := &EdgeData{
		ID:     "edge-1",
		Source: "node-1",
		Target: "node-2",
		Type:   "connection",
	}

	assert.Equal(t, "", edge.Label)
	assert.Equal(t, "", edge.Relationship)
}

func TestPaginatedResponse_EmptyData(t *testing.T) {
	resp := &PaginatedResponse{
		Data:       []interface{}{},
		Page:       1,
		Limit:      50,
		Total:      0,
		TotalPages: 0,
	}

	assert.NotNil(t, resp.Data)
	assert.Equal(t, 0, resp.Total)
}

func TestAPIResponse_WithNilData(t *testing.T) {
	resp := &APIResponse{
		Success: true,
		Data:    nil,
		Error:   nil,
	}

	assert.True(t, resp.Success)
	assert.Nil(t, resp.Data)
}

func TestNodeData_WithAllFields(t *testing.T) {
	data := NodeData{
		ID:           "node-complete",
		Label:        "Complete Node",
		Type:         "aws_resource",
		ResourceType: "EC2",
		ResourceName: "test-instance",
		Severity:     "medium",
		Metadata: map[string]interface{}{
			"nested": map[string]string{
				"key": "value",
			},
		},
	}

	assert.NotEmpty(t, data.ID)
	assert.NotEmpty(t, data.Label)
	assert.NotEmpty(t, data.Type)
	assert.NotEmpty(t, data.ResourceType)
	assert.NotEmpty(t, data.ResourceName)
	assert.NotEmpty(t, data.Severity)
	assert.NotNil(t, data.Metadata)
}
