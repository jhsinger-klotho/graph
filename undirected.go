package graph

import "errors"

type UndirectedGraph[K comparable, T any] struct {
	Graph[K, T]
}

func NewMemoryUndirected[K comparable, T any](hash Hash[K, T]) *UndirectedGraph[K, T] {
	return &UndirectedGraph[K, T]{Graph: NewMemoryGraph[K, T](hash)}
}

func (u *UndirectedGraph[K, T]) AddEdge(sourceHash, targetHash K, options ...func(*EdgeProperties)) error {
	return errors.Join(
		u.Graph.AddEdge(sourceHash, targetHash, options...),
		u.Graph.AddEdge(targetHash, sourceHash, options...),
	)
}

func (u *UndirectedGraph[K, T]) RemoveEdge(source, target K) error {
	return errors.Join(
		u.Graph.RemoveEdge(source, target),
		u.Graph.RemoveEdge(target, source),
	)
}

func (u *UndirectedGraph[K, T]) UpdateEdge(source, target K, options ...func(*EdgeProperties)) error {
	return errors.Join(
		u.Graph.UpdateEdge(source, target, options...),
		u.Graph.UpdateEdge(target, source, options...),
	)
}

func (u *UndirectedGraph[K, T]) Edges() func(yield func(Edge[K], error) bool) {
	// An undirected graph creates each edge twice internally: The edge (A,B) is
	// stored both as (A,B) and (B,A). The Edges method is supposed to return
	// one of these two edges, because from an outside perspective, it only is
	// a single edge.
	//
	// To achieve this, Edges keeps track of already-seen edges. For each edge,
	// it also checks if the reversed edge has already been seen - e.g., for
	// an edge (A,B), Edges checks if the edge has been seen as (B,A).
	seen := make(map[EdgeKey[K]]struct{})

	return func(yield func(Edge[K], error) bool) {
		for e, err := range u.Graph.Edges() {
			if err == nil {
				k := EdgeKey[K]{Source: e.Target, Target: e.Source}
				if _, ok := seen[k]; ok {
					continue
				}
				k.Source, k.Target = e.Source, e.Target
				if _, ok := seen[k]; ok {
					continue
				}
				seen[k] = struct{}{}
			}
			if !yield(e, err) {
				return
			}
		}
	}
}
func (u *UndirectedGraph[K, T]) AdjacencyMap() (map[K]map[K]Edge[K], error) {
	adj := make(map[K]map[K]Edge[K])
	for v, err := range u.Vertices() {
		if err != nil {
			return nil, err
		}
		adj[u.Hash(v.Value)] = make(map[K]Edge[K])
	}
	// Use the underlying Edges so that the inverted ones are included.
	for e, err := range u.Graph.Edges() {
		if err != nil {
			return nil, err
		}
		adj[e.Source][e.Target] = e
	}
	return adj, nil

}

func (u *UndirectedGraph[K, T]) PredecessorMap() (map[K]map[K]Edge[K], error) {
	return u.AdjacencyMap()
}
