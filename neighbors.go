package graph

type GraphNeighbors[K comparable, E any] interface {
	// DownstreamNeighbors returns all edges that have the given vertex as their source. A default implementation built on top of [AdjacencyMap] is provided: [DownstreamNeighbors].
	DownstreamNeighbors(K) EdgeIter[K, E]
	// UpstreamNeighbors returns all edges that have the given vertex as their target. A default implementation built on top of [PredecessorMap] is provided: [UpstreamNeighbors].
	UpstreamNeighbors(K) EdgeIter[K, E]
}

// DownstreamNeighbors returns an iterator over the downstream neighbors of the given vertex.
// Thus, all edges have `.Source == source`.
func DownstreamNeighbors[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, hash K) EdgeIter[K, E] {
	return func(yield func(Edge[K, E], error) bool) {
		adj, err := g.AdjacencyMap()
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

func (d DefaultGraph[K, V, E]) DownstreamNeighbors(hash K) EdgeIter[K, E] {
	return DownstreamNeighbors[K, V, E](d, hash)
}

// UpstreamNeighbors returns an iterator over the upstream neighbors of the given vertex.
// Thus, all edges have `.Target == source`.
func UpstreamNeighbors[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, hash K) EdgeIter[K, E] {
	return func(yield func(Edge[K, E], error) bool) {
		pred, err := g.PredecessorMap()
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

func (d DefaultGraph[K, V, E]) UpstreamNeighbors(hash K) EdgeIter[K, E] {
	return UpstreamNeighbors[K, V, E](d, hash)
}
