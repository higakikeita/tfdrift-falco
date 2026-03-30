package graph

import (
	"fmt"
	"testing"
)

// BenchmarkAddNode benchmarks adding nodes
func BenchmarkAddNode(b *testing.B) {
	db := NewDatabase()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := &Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource", "EC2"},
			Properties: map[string]interface{}{"index": i},
		}
		db.AddNode(node)
	}
}

// BenchmarkGetNode benchmarks node retrieval
func BenchmarkGetNode(b *testing.B) {
	db := NewDatabase()

	// Pre-populate
	numNodes := 1000
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.GetNode(fmt.Sprintf("node-%d", i%numNodes))
	}
}

// BenchmarkAddRelationship benchmarks adding relationships
func BenchmarkAddRelationship(b *testing.B) {
	db := NewDatabase()

	// Create nodes first
	numNodes := b.N + 1
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rel := &Relationship{
			ID:        fmt.Sprintf("rel-%d", i),
			Type:      DEPENDS_ON,
			StartNode: fmt.Sprintf("node-%d", i),
			EndNode:   fmt.Sprintf("node-%d", i+1),
			Properties: map[string]interface{}{},
		}
		_ = db.AddRelationship(rel)
	}
}

// BenchmarkGetRelationshipsByType benchmarks filtering relationships
func BenchmarkGetRelationshipsByType(b *testing.B) {
	db := NewDatabase()

	// Create a graph with multiple relationship types
	numNodes := 100
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	// Add relationships of different types
	for i := 0; i < numNodes-1; i++ {
		relType := DEPENDS_ON
		if i%2 == 0 {
			relType = CONTAINS
		}

		_ = db.AddRelationship(&Relationship{
			ID:        fmt.Sprintf("rel-%d", i),
			Type:      relType,
			StartNode: fmt.Sprintf("node-%d", i),
			EndNode:   fmt.Sprintf("node-%d", i+1),
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.GetRelationshipsByType(DEPENDS_ON)
	}
}

// BenchmarkGetNodesByLabel benchmarks label-based queries
func BenchmarkGetNodesByLabel(b *testing.B) {
	db := NewDatabase()

	// Create nodes with different labels
	numNodes := 1000
	labels := []string{"EC2", "VPC", "Subnet", "SecurityGroup"}

	for i := 0; i < numNodes; i++ {
		label := labels[i%len(labels)]
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource", label},
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		label := labels[i%len(labels)]
		_ = db.GetNodesByLabel(label)
	}
}

// BenchmarkFindPath benchmarks pathfinding
func BenchmarkFindPath(b *testing.B) {
	db := NewDatabase()

	// Create a chain of nodes
	numNodes := 100
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	// Link them in a chain
	for i := 0; i < numNodes-1; i++ {
		_ = db.AddRelationship(&Relationship{
			ID:        fmt.Sprintf("rel-%d", i),
			Type:      DEPENDS_ON,
			StartNode: fmt.Sprintf("node-%d", i),
			EndNode:   fmt.Sprintf("node-%d", i+1),
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := i % (numNodes - 10)
		end := start + 10
		_, _ = db.FindPath(fmt.Sprintf("node-%d", start), fmt.Sprintf("node-%d", end))
	}
}

// BenchmarkFindDependencies benchmarks dependency traversal
func BenchmarkFindDependencies(b *testing.B) {
	db := NewDatabase()

	// Create a tree structure
	numNodes := 100
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	// Create dependencies: each node depends on next 3
	for i := 0; i < numNodes-3; i++ {
		for j := 1; j <= 3; j++ {
			_ = db.AddRelationship(&Relationship{
				ID:        fmt.Sprintf("rel-%d-%d", i, j),
				Type:      DEPENDS_ON,
				StartNode: fmt.Sprintf("node-%d", i),
				EndNode:   fmt.Sprintf("node-%d", i+j),
				Properties: map[string]interface{}{},
			})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodeIdx := i % numNodes
		_ = db.FindDependencies(fmt.Sprintf("node-%d", nodeIdx), 5)
	}
}

// BenchmarkConcurrentReads benchmarks concurrent read operations
func BenchmarkConcurrentReads(b *testing.B) {
	db := NewDatabase()

	// Pre-populate database
	numNodes := 100
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			db.GetNode(fmt.Sprintf("node-%d", i%numNodes))
			i++
		}
	})
}

// BenchmarkGetAllNodes benchmarks retrieving all nodes
func BenchmarkGetAllNodes(b *testing.B) {
	db := NewDatabase()

	// Pre-populate with various sizes
	numNodes := 1000
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.GetAllNodes()
	}
}

// BenchmarkGetAllRelationships benchmarks retrieving all relationships
func BenchmarkGetAllRelationships(b *testing.B) {
	db := NewDatabase()

	// Create nodes and relationships
	numNodes := 100
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	for i := 0; i < numNodes-1; i++ {
		_ = db.AddRelationship(&Relationship{
			ID:        fmt.Sprintf("rel-%d", i),
			Type:      DEPENDS_ON,
			StartNode: fmt.Sprintf("node-%d", i),
			EndNode:   fmt.Sprintf("node-%d", i+1),
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.GetAllRelationships()
	}
}

// BenchmarkGetOutgoingRelationships benchmarks outgoing relationship retrieval
func BenchmarkGetOutgoingRelationships(b *testing.B) {
	db := NewDatabase()

	// Create a star graph with node 0 at center
	numNodes := 100
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	// All other nodes depend on node-0
	for i := 1; i < numNodes; i++ {
		_ = db.AddRelationship(&Relationship{
			ID:        fmt.Sprintf("rel-%d", i),
			Type:      DEPENDS_ON,
			StartNode: "node-0",
			EndNode:   fmt.Sprintf("node-%d", i),
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.GetOutgoingRelationships("node-0")
	}
}

// BenchmarkGetNeighbors benchmarks neighbor retrieval
func BenchmarkGetNeighbors(b *testing.B) {
	db := NewDatabase()

	// Create nodes
	numNodes := 100
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	// Create bidirectional relationships
	for i := 0; i < numNodes-1; i++ {
		_ = db.AddRelationship(&Relationship{
			ID:        fmt.Sprintf("rel-out-%d", i),
			Type:      DEPENDS_ON,
			StartNode: fmt.Sprintf("node-%d", i),
			EndNode:   fmt.Sprintf("node-%d", i+1),
			Properties: map[string]interface{}{},
		})

		_ = db.AddRelationship(&Relationship{
			ID:        fmt.Sprintf("rel-in-%d", i),
			Type:      DEPENDS_ON,
			StartNode: fmt.Sprintf("node-%d", i+1),
			EndNode:   fmt.Sprintf("node-%d", i),
			Properties: map[string]interface{}{},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodeIdx := i % numNodes
		_ = db.GetNeighbors(fmt.Sprintf("node-%d", nodeIdx))
	}
}

// BenchmarkClear benchmarks clearing the database
func BenchmarkClear(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db := NewDatabase()

		// Add nodes and relationships
		numNodes := 100
		for j := 0; j < numNodes; j++ {
			db.AddNode(&Node{
				ID:         fmt.Sprintf("node-%d", j),
				Labels:     []string{"Resource"},
				Properties: map[string]interface{}{},
			})
		}

		for j := 0; j < numNodes-1; j++ {
			_ = db.AddRelationship(&Relationship{
				ID:        fmt.Sprintf("rel-%d", j),
				Type:      DEPENDS_ON,
				StartNode: fmt.Sprintf("node-%d", j),
				EndNode:   fmt.Sprintf("node-%d", j+1),
				Properties: map[string]interface{}{},
			})
		}

		b.StartTimer()
		db.Clear()
		b.StopTimer()
	}
}
