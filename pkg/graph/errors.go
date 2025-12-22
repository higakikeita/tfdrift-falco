package graph

import "errors"

var (
	// ErrNodeNotFound is returned when a node is not found
	ErrNodeNotFound = errors.New("node not found")

	// ErrRelationshipNotFound is returned when a relationship is not found
	ErrRelationshipNotFound = errors.New("relationship not found")

	// ErrCyclicDependency is returned when a cyclic dependency is detected
	ErrCyclicDependency = errors.New("cyclic dependency detected")

	// ErrInvalidPath is returned when a path is invalid
	ErrInvalidPath = errors.New("invalid path")
)
