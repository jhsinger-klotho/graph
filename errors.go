package graph

import (
	"errors"
	"fmt"
)

type (
	VertexAlreadyExistsError[K comparable, T any] struct {
		Key            K
		ExistingVertex Vertex[T]
	}

	VertexNotFoundError[K comparable] struct {
		Key K
	}

	EdgeAlreadyExistsError[K comparable, E any] struct {
		ExistingEdge Edge[K, E]
	}

	EdgeNotFoundError[K comparable] struct {
		Source, Target K
	}

	VertexHasEdgesError[K comparable] struct {
		Key   K
		Count int
	}

	EdgeCausesCycleError[K comparable] struct {
		Source, Target K
	}

	UpdateChangedKeyError[K comparable] struct {
		OldKey, NewKey K
	}
)

func (e *VertexAlreadyExistsError[K, T]) Error() string {
	return fmt.Sprintf("vertex %v already exists with value %v", e.Key, e.ExistingVertex.Value)
}

func (e *VertexNotFoundError[K]) Error() string {
	return fmt.Sprintf("vertex %v not found", e.Key)
}

func (e *EdgeAlreadyExistsError[K, E]) Error() string {
	return fmt.Sprintf("edge %v - %v already exists", e.ExistingEdge.Source, e.ExistingEdge.Target)
}

func (e *EdgeNotFoundError[K]) Error() string {
	return fmt.Sprintf("edge %v - %v not found", e.Source, e.Target)
}

func (e *VertexHasEdgesError[K]) Error() string {
	return fmt.Sprintf("vertex %v has %d edges", e.Key, e.Count)
}

func (e *EdgeCausesCycleError[K]) Error() string {
	return fmt.Sprintf("edge %v - %v would cause a cycle", e.Source, e.Target)
}

func (e *UpdateChangedKeyError[K]) Error() string {
	return fmt.Sprintf("vertex key changed from %v to %v", e.OldKey, e.NewKey)
}

var (
	ErrVertexNotFound      = errors.New("vertex not found")
	ErrVertexAlreadyExists = errors.New("vertex already exists")
	ErrEdgeNotFound        = errors.New("edge not found")
	ErrEdgeAlreadyExists   = errors.New("edge already exists")
	ErrEdgeCreatesCycle    = errors.New("edge would create a cycle")
	ErrVertexHasEdges      = errors.New("vertex has edges")
)

func (e *VertexAlreadyExistsError[K, T]) Unwrap() error { return ErrVertexAlreadyExists }
func (e *VertexNotFoundError[K]) Unwrap() error         { return ErrVertexNotFound }
func (e *EdgeAlreadyExistsError[K, E]) Unwrap() error   { return ErrEdgeAlreadyExists }
func (e *EdgeNotFoundError[K]) Unwrap() error           { return ErrEdgeNotFound }
func (e *VertexHasEdgesError[K]) Unwrap() error         { return ErrVertexHasEdges }
func (e *EdgeCausesCycleError[K]) Unwrap() error        { return ErrEdgeCreatesCycle }
