package graph

type GraphNeighbors[K comparable, E any] interface {
	DownstreamNeighbors(K) func(yield func(Edge[K, E], error) bool)
	UpstreamNeighbors(K) func(yield func(Edge[K, E], error) bool)
}

// DownstreamNeighbors returns an iterator over the downstream neighbors of the given vertex.
// Thus, all edges have `.Source == source`.
func DownstreamNeighbors[K comparable, V any, E any](g GraphRead[K, V, E], hash K) func(yield func(Edge[K, E], error) bool) {
	if rel, ok := g.(interface {
		DownstreamNeighbors(K) func(yield func(Edge[K, E], error) bool)
	}); ok {
		return rel.DownstreamNeighbors(hash)
	}

	return func(yield func(Edge[K, E], error) bool) {
		adj, err := AdjacencyMap(g)
		if err != nil {
			yield(Edge[K, E]{}, err)
			return
		}
		for _, adjacencies := range adj {
			for _, edge := range adjacencies {
				if !yield(edge, nil) {
					return
				}
			}
		}
	}
}

// UpstreamNeighbors returns an iterator over the upstream neighbors of the given vertex.
// Thus, all edges have `.Target == source`.
func UpstreamNeighbors[K comparable, V any, E any](g GraphRead[K, V, E], hash K) func(yield func(Edge[K, E], error) bool) {
	if rel, ok := g.(interface {
		UpstreamNeighbors(K) func(yield func(Edge[K, E], error) bool)
	}); ok {
		return rel.UpstreamNeighbors(hash)
	}

	return func(yield func(Edge[K, E], error) bool) {
		pred, err := PredecessorMap(g)
		if err != nil {
			yield(Edge[K, E]{}, err)
			return
		}
		for _, preds := range pred {
			for _, edge := range preds {
				if !yield(edge, nil) {
					return
				}
			}
		}
	}
}
