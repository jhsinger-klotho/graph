package graph

// GraphRelations is used for graphs that can provide more efficient
// implementations of the [AdjacencyMap] and [PredecessorMap] methods.
type GraphRelations[K comparable] interface {
	AdjacencyMap() (map[K]map[K]Edge[K], error)
	PredecessorMap() (map[K]map[K]Edge[K], error)
}

// PredecessorMap computes a predecessor map with all vertices in the graph.
//
// It has the same map layout and does the same thing as AdjacencyMap, but
// for ingoing instead of outgoing edges of each vertex.
//
// For a directed graph with two edges AB and AC, PredecessorMap would
// return the following map:
//
//	map[string]map[string]Edge[string]{
//		"A": map[string]Edge[string]{},
//		"B": map[string]Edge[string]{
//			"A": {Source: "A", Target: "B"},
//		},
//		"C": map[string]Edge[string]{
//			"A": {Source: "A", Target: "C"},
//		},
//	}
//
// For an undirected graph, PredecessorMap is the same as AdjacencyMap. This
// is because there is no distinction between "outgoing" and "ingoing" edges
// in an undirected graph.
// If the graph does not implement [GraphRelations], PredecessorMap will be
// computed from the edges of the graph.
func PredecessorMap[K comparable, T any](g GraphRead[K, T]) (map[K]map[K]Edge[K], error) {
	if rel, ok := g.(interface {
		PredecessorMap() (map[K]map[K]Edge[K], error)
	}); ok {
		return rel.PredecessorMap()
	}
	adj := make(map[K]map[K]Edge[K])
	for v, err := range g.Vertices() {
		if err != nil {
			return nil, err
		}
		adj[g.Hash(v.Value)] = make(map[K]Edge[K])
	}
	for e, err := range g.Edges() {
		if err != nil {
			return nil, err
		}
		if _, ok := adj[e.Target]; !ok {
			adj[e.Target] = make(map[K]Edge[K])
		}
		adj[e.Target][e.Source] = e
	}
	return adj, nil
}

// AdjacencyMap computes an adjacency map with all vertices in the graph.
//
// There is an entry for each vertex. Each of those entries is another map
// whose keys are the hash values of the adjacent vertices. The value is an
// Edge instance that stores the source and target hash values along with
// the edge metadata.
//
// For a directed graph with two edges AB and AC, AdjacencyMap would return
// the following map:
//
//	map[string]map[string]Edge[string]{
//		"A": map[string]Edge[string]{
//			"B": {Source: "A", Target: "B"},
//			"C": {Source: "A", Target: "C"},
//		},
//		"B": map[string]Edge[string]{},
//		"C": map[string]Edge[string]{},
//	}
//
// This design makes AdjacencyMap suitable for a wide variety of algorithms.
// If the graph does not implement [GraphRelations], AdjacencyMap will be
// computed from the edges of the graph.
func AdjacencyMap[K comparable, T any](g GraphRead[K, T]) (map[K]map[K]Edge[K], error) {
	if rel, ok := g.(interface {
		AdjacencyMap() (map[K]map[K]Edge[K], error)
	}); ok {
		return rel.AdjacencyMap()
	}
	adj := make(map[K]map[K]Edge[K])
	for v, err := range g.Vertices() {
		if err != nil {
			return nil, err
		}
		adj[g.Hash(v.Value)] = make(map[K]Edge[K])
	}
	for e, err := range g.Edges() {
		if err != nil {
			return nil, err
		}
		if _, ok := adj[e.Source]; !ok {
			adj[e.Source] = make(map[K]Edge[K])
		}
		adj[e.Source][e.Target] = e
	}
	return adj, nil
}
