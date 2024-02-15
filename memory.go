package graph

import (
	"fmt"
)

type memoryGraph[K comparable, T any] struct {
	hash     Hash[K, T]
	vertices map[K]*Vertex[T]
	edges    map[EdgeKey[K]]*Edge[K]
}

var (
	_ Graph[string, string]  = (*memoryGraph[string, string])(nil)
	_ GraphRelations[string] = (*memoryGraph[string, string])(nil)
)

// NewMemoryGraph creates a new graph with a memory store, this is not a thread safe implementation.
// It is a directed graph that may contain cycles.
func NewMemoryGraph[K comparable, T any](hash Hash[K, T]) *memoryGraph[K, T] {
	return &memoryGraph[K, T]{
		hash:     hash,
		vertices: make(map[K]*Vertex[T]),
		edges:    make(map[EdgeKey[K]]*Edge[K]),
	}
}

func (s *memoryGraph[K, T]) Hash(v T) K {
	return s.hash(v)
}

func (s *memoryGraph[K, T]) Vertex(hash K) (Vertex[T], error) {
	if vertex, ok := s.vertices[hash]; ok {
		return *vertex, nil
	}
	return Vertex[T]{}, &VertexNotFoundError[K]{Key: hash}
}

func (s *memoryGraph[K, T]) Vertices() func(yield func(Vertex[T], error) bool) {
	return func(yield func(Vertex[T], error) bool) {
		for _, v := range s.vertices {
			if !yield(*v, nil) {
				return
			}
		}
	}
}

func (s *memoryGraph[K, T]) Edge(sourceHash, targetHash K) (Edge[K], error) {
	e, ok := s.edges[EdgeKey[K]{Source: sourceHash, Target: targetHash}]
	if !ok {
		return Edge[K]{}, &EdgeNotFoundError[K]{Source: sourceHash, Target: targetHash}
	}
	return *e, nil
}

func (s *memoryGraph[K, T]) Edges() func(yield func(Edge[K], error) bool) {
	return func(yield func(Edge[K], error) bool) {
		for _, e := range s.edges {
			if !yield(*e, nil) {
				return
			}
		}
	}
}

func (s *memoryGraph[K, T]) Order() (int, error) {
	return len(s.vertices), nil
}

func (s *memoryGraph[K, T]) Size() (int, error) {
	return len(s.edges), nil
}

func (s *memoryGraph[K, T]) AddVertex(value T, options ...func(*VertexProperties)) error {
	k := s.hash(value)
	if _, ok := s.vertices[k]; ok {
		return &VertexAlreadyExistsError[K, T]{Key: k, ExistingValue: *s.vertices[k]}
	}

	v := &Vertex[T]{Value: value}
	for _, option := range options {
		option(&v.Properties)
	}
	s.vertices[k] = v

	return nil
}

func (s *memoryGraph[K, T]) UpdateVertex(hash K, options ...func(*Vertex[T])) error {
	v, ok := s.vertices[hash]
	if !ok {
		return &VertexNotFoundError[K]{Key: hash}
	}
	for _, option := range options {
		option(v)
	}
	newKey := s.hash(v.Value)
	if newKey != hash {
		if _, ok := s.vertices[newKey]; ok {
			return &VertexAlreadyExistsError[K, T]{Key: newKey, ExistingValue: *s.vertices[newKey]}
		}
		delete(s.vertices, hash)
		s.vertices[newKey] = v
	}
	return nil
}

func (s *memoryGraph[K, T]) RemoveVertex(hash K) error {
	if _, ok := s.vertices[hash]; !ok {
		return &VertexNotFoundError[K]{Key: hash}
	}
	for k := range s.edges {
		if k.Source == hash || k.Target == hash {
			return &VertexHasEdgesError[K]{Key: hash, Count: 1}
		}
	}
	delete(s.vertices, hash)
	return nil
}

func (s *memoryGraph[K, T]) AddEdge(sourceHash, targetHash K, options ...func(*EdgeProperties)) error {
	k := EdgeKey[K]{Source: sourceHash, Target: targetHash}
	if _, ok := s.edges[k]; ok {
		return &EdgeAlreadyExistsError[K]{Source: sourceHash, Target: targetHash}
	}
	_, ok := s.vertices[sourceHash]
	if !ok {
		return &VertexNotFoundError[K]{Key: sourceHash}
	}
	_, ok = s.vertices[targetHash]
	if !ok {
		return &VertexNotFoundError[K]{Key: targetHash}
	}
	edge := &Edge[K]{
		Source: sourceHash,
		Target: targetHash,
	}
	for _, option := range options {
		option(&edge.Properties)
	}
	s.edges[k] = edge
	return nil
}

func (s *memoryGraph[K, T]) UpdateEdge(source, target K, options ...func(properties *EdgeProperties)) error {
	k := EdgeKey[K]{Source: source, Target: target}
	e, ok := s.edges[k]
	if !ok {
		return &EdgeNotFoundError[K]{Source: source, Target: target}
	}
	for _, option := range options {
		option(&e.Properties)
	}
	return nil
}

func (s *memoryGraph[K, T]) RemoveEdge(source, target K) error {
	k := EdgeKey[K]{Source: source, Target: target}
	if _, ok := s.edges[k]; !ok {
		return &EdgeNotFoundError[K]{Source: source, Target: target}
	}
	delete(s.edges, k)
	return nil
}

func (s *memoryGraph[K, T]) AdjacencyMap() (map[K]map[K]Edge[K], error) {
	adj := make(map[K]map[K]Edge[K])
	for k := range s.vertices {
		adj[k] = make(map[K]Edge[K])
	}
	for _, e := range s.edges {
		adj[e.Source][e.Target] = *e
	}
	return adj, nil

}

func (s *memoryGraph[K, T]) PredecessorMap() (map[K]map[K]Edge[K], error) {
	pred := make(map[K]map[K]Edge[K])
	for k := range s.vertices {
		pred[k] = make(map[K]Edge[K])
	}
	for _, e := range s.edges {
		pred[e.Target][e.Source] = *e
	}
	return pred, nil
}

// CreatesCycle is a fastpath version of [CreatesCycle] that avoids calling
// [PredecessorMap], which generates large amounts of garbage to collect.
//
// Because CreatesCycle doesn't need to modify the PredecessorMap, we can use
// inEdges instead to compute the same thing without creating any copies.
func (s *memoryGraph[K, T]) CreatesCycle(source, target K) (bool, error) {
	if _, ok := s.vertices[source]; !ok {
		return false, fmt.Errorf("could not get source vertex: %w", &VertexNotFoundError[K]{Key: source})
	}

	if _, ok := s.vertices[target]; !ok {
		return false, fmt.Errorf("could not get target vertex: %w", &VertexNotFoundError[K]{Key: target})
	}

	if source == target {
		return true, nil
	}

	stack := newStack[K]()
	visited := make(map[K]struct{})

	stack.push(source)

	for !stack.isEmpty() {
		currentHash, _ := stack.pop()

		if _, ok := visited[currentHash]; !ok {
			// If the adjacent vertex also is the target vertex, the target is a
			// parent of the source vertex. An edge would introduce a cycle.
			if currentHash == target {
				return true, nil
			}

			visited[currentHash] = struct{}{}

			for e := range s.edges {
				if e.Target == currentHash {
					stack.push(e.Source)
				}
			}
		}
	}

	return false, nil
}
