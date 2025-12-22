package graph

// Path represents a path through the graph
type Path struct {
	Nodes         []*Node         `json:"nodes"`
	Relationships []*Relationship `json:"relationships"`
	Length        int             `json:"length"`
}

// TraversalResult holds the results of a graph traversal
type TraversalResult struct {
	Nodes     []*Node         `json:"nodes"`
	Distances map[string]int  `json:"distances"` // node_id -> distance from start
	Parents   map[string]string `json:"parents"`   // node_id -> parent node_id
}

// FindPath finds a path between two nodes using BFS
// Example: Find path from EC2 to VPC
func (db *GraphDatabase) FindPath(startID, endID string) (*Path, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.nodes[startID] == nil || db.nodes[endID] == nil {
		return nil, ErrNodeNotFound
	}

	// BFS to find shortest path
	visited := make(map[string]bool)
	parents := make(map[string]string)
	queue := []string{startID}
	visited[startID] = true

	found := false
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == endID {
			found = true
			break
		}

		// Explore outgoing relationships
		for _, rel := range db.outgoing[current] {
			neighbor := rel.EndNode
			if !visited[neighbor] {
				visited[neighbor] = true
				parents[neighbor] = current
				queue = append(queue, neighbor)
			}
		}

		// Explore incoming relationships (for undirected traversal)
		for _, rel := range db.incoming[current] {
			neighbor := rel.StartNode
			if !visited[neighbor] {
				visited[neighbor] = true
				parents[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	if !found {
		return nil, ErrInvalidPath
	}

	// Reconstruct path
	path := &Path{
		Nodes:         []*Node{},
		Relationships: []*Relationship{},
	}

	// Build path backwards from end to start
	current := endID
	nodePath := []string{current}
	for current != startID {
		parent := parents[current]
		nodePath = append(nodePath, parent)
		current = parent
	}

	// Reverse to get start -> end
	for i := len(nodePath) - 1; i >= 0; i-- {
		path.Nodes = append(path.Nodes, db.nodes[nodePath[i]])
	}

	// Find relationships between consecutive nodes
	for i := 0; i < len(nodePath)-1; i++ {
		from := nodePath[len(nodePath)-1-i]
		to := nodePath[len(nodePath)-2-i]

		// Check outgoing relationships
		for _, rel := range db.outgoing[from] {
			if rel.EndNode == to {
				path.Relationships = append(path.Relationships, rel)
				break
			}
		}

		// Check incoming relationships
		for _, rel := range db.incoming[from] {
			if rel.StartNode == to {
				path.Relationships = append(path.Relationships, rel)
				break
			}
		}
	}

	path.Length = len(path.Nodes) - 1
	return path, nil
}

// FindImpactRadius finds all nodes within N hops of a starting node
// Example: Find all resources affected within 3 hops of a changed VPC
func (db *GraphDatabase) FindImpactRadius(startID string, maxDepth int) *TraversalResult {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.nodes[startID] == nil {
		return &TraversalResult{
			Nodes:     []*Node{},
			Distances: make(map[string]int),
			Parents:   make(map[string]string),
		}
	}

	distances := make(map[string]int)
	parents := make(map[string]string)
	queue := []string{startID}
	distances[startID] = 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentDist := distances[current]

		if currentDist >= maxDepth {
			continue
		}

		// Explore outgoing relationships
		for _, rel := range db.outgoing[current] {
			neighbor := rel.EndNode
			if _, visited := distances[neighbor]; !visited {
				distances[neighbor] = currentDist + 1
				parents[neighbor] = current
				queue = append(queue, neighbor)
			}
		}

		// Explore incoming relationships
		for _, rel := range db.incoming[current] {
			neighbor := rel.StartNode
			if _, visited := distances[neighbor]; !visited {
				distances[neighbor] = currentDist + 1
				parents[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	// Collect nodes
	nodes := make([]*Node, 0, len(distances))
	for nodeID := range distances {
		if node := db.nodes[nodeID]; node != nil {
			nodes = append(nodes, node)
		}
	}

	return &TraversalResult{
		Nodes:     nodes,
		Distances: distances,
		Parents:   parents,
	}
}

// FindDependencies finds all transitive dependencies of a node
// Example: Find all resources that an EC2 instance depends on
func (db *GraphDatabase) FindDependencies(nodeID string, maxDepth int) []*Node {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.nodes[nodeID] == nil {
		return []*Node{}
	}

	visited := make(map[string]bool)
	queue := []struct {
		id    string
		depth int
	}{{nodeID, 0}}

	dependencies := []*Node{}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth >= maxDepth {
			continue
		}

		if visited[current.id] {
			continue
		}
		visited[current.id] = true

		// Find DEPENDS_ON relationships
		for _, rel := range db.outgoing[current.id] {
			if rel.Type == DEPENDS_ON {
				if node := db.nodes[rel.EndNode]; node != nil && rel.EndNode != nodeID {
					dependencies = append(dependencies, node)
					queue = append(queue, struct {
						id    string
						depth int
					}{rel.EndNode, current.depth + 1})
				}
			}
		}
	}

	return dependencies
}

// FindDependents finds all nodes that depend on this node
// Example: Find all resources that would be affected if a VPC is changed
func (db *GraphDatabase) FindDependents(nodeID string, maxDepth int) []*Node {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.nodes[nodeID] == nil {
		return []*Node{}
	}

	visited := make(map[string]bool)
	queue := []struct {
		id    string
		depth int
	}{{nodeID, 0}}

	dependents := []*Node{}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth >= maxDepth {
			continue
		}

		if visited[current.id] {
			continue
		}
		visited[current.id] = true

		// Find incoming DEPENDS_ON relationships (reverse direction)
		for _, rel := range db.incoming[current.id] {
			if rel.Type == DEPENDS_ON {
				if node := db.nodes[rel.StartNode]; node != nil && rel.StartNode != nodeID {
					dependents = append(dependents, node)
					queue = append(queue, struct {
						id    string
						depth int
					}{rel.StartNode, current.depth + 1})
				}
			}
		}
	}

	return dependents
}

// FindCriticalPaths finds paths with highest dependency chains
// Returns nodes that have many dependents (critical points of failure)
func (db *GraphDatabase) FindCriticalPaths(minDependents int) []*Node {
	db.mu.RLock()
	defer db.mu.RUnlock()

	criticalNodes := []*Node{}

	for nodeID, node := range db.nodes {
		// Count incoming DEPENDS_ON relationships
		dependentCount := 0
		for _, rel := range db.incoming[nodeID] {
			if rel.Type == DEPENDS_ON {
				dependentCount++
			}
		}

		if dependentCount >= minDependents {
			criticalNodes = append(criticalNodes, node)
		}
	}

	return criticalNodes
}

// MatchPattern performs a simple pattern matching query
// Example: Find all EC2 instances that depend on a specific subnet
// Pattern: (ec2:EC2)-[:DEPENDS_ON]->(subnet:Subnet {id: "subnet-123"})
type MatchPattern struct {
	StartLabels []string               // Labels for start node
	RelType     string                 // Relationship type
	EndLabels   []string               // Labels for end node
	EndFilter   map[string]interface{} // Properties to match on end node
}

// Match executes a pattern matching query
func (db *GraphDatabase) Match(pattern *MatchPattern) [][]*Node {
	db.mu.RLock()
	defer db.mu.RUnlock()

	results := [][]*Node{}

	// Find all nodes matching start labels
	startCandidates := db.findNodesWithLabels(pattern.StartLabels)

	for _, startNode := range startCandidates {
		// Check outgoing relationships
		for _, rel := range db.outgoing[startNode.ID] {
			// Match relationship type
			if pattern.RelType != "" && rel.Type != pattern.RelType {
				continue
			}

			endNode := db.nodes[rel.EndNode]
			if endNode == nil {
				continue
			}

			// Match end node labels
			if !db.hasAllLabels(endNode, pattern.EndLabels) {
				continue
			}

			// Match end node properties
			if !db.matchesProperties(endNode, pattern.EndFilter) {
				continue
			}

			// Found a match
			results = append(results, []*Node{startNode, endNode})
		}
	}

	return results
}

// Helper: find nodes with all specified labels
func (db *GraphDatabase) findNodesWithLabels(labels []string) []*Node {
	if len(labels) == 0 {
		// Return all nodes
		nodes := make([]*Node, 0, len(db.nodes))
		for _, node := range db.nodes {
			nodes = append(nodes, node)
		}
		return nodes
	}

	// Start with nodes having the first label
	candidates := make(map[string]*Node)
	if labelMap := db.nodesByLabel[labels[0]]; labelMap != nil {
		for id, node := range labelMap {
			candidates[id] = node
		}
	}

	// Filter by remaining labels
	for _, label := range labels[1:] {
		filtered := make(map[string]*Node)
		for id, node := range candidates {
			if db.hasLabel(node, label) {
				filtered[id] = node
			}
		}
		candidates = filtered
	}

	// Convert to slice
	nodes := make([]*Node, 0, len(candidates))
	for _, node := range candidates {
		nodes = append(nodes, node)
	}

	return nodes
}

// Helper: check if node has all labels
func (db *GraphDatabase) hasAllLabels(node *Node, labels []string) bool {
	for _, label := range labels {
		if !db.hasLabel(node, label) {
			return false
		}
	}
	return true
}

// Helper: check if node has a label
func (db *GraphDatabase) hasLabel(node *Node, label string) bool {
	for _, l := range node.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// Helper: check if node matches property filters
func (db *GraphDatabase) matchesProperties(node *Node, filter map[string]interface{}) bool {
	for key, expectedValue := range filter {
		actualValue, exists := node.Properties[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}
	return true
}
