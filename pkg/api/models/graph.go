package models

// CytoscapeElements represents the full graph in Cytoscape.js format
type CytoscapeElements struct {
	Nodes []CytoscapeNode `json:"nodes"`
	Edges []CytoscapeEdge `json:"edges"`
}

// CytoscapeNode represents a node in Cytoscape.js format
type CytoscapeNode struct {
	Data NodeData `json:"data"`
}

// NodeData represents node data
type NodeData struct {
	ID           string                 `json:"id"`
	Label        string                 `json:"label"`
	Type         string                 `json:"type"`
	ResourceType string                 `json:"resource_type"`
	ResourceName string                 `json:"resource_name,omitempty"`
	Severity     string                 `json:"severity,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CytoscapeEdge represents an edge in Cytoscape.js format
type CytoscapeEdge struct {
	Data EdgeData `json:"data"`
}

// EdgeData represents edge data
type EdgeData struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	Label        string `json:"label,omitempty"`
	Type         string `json:"type"`
	Relationship string `json:"relationship,omitempty"`
}
