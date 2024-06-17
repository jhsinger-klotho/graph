package graph

type GraphRelations[K comparable, E any] interface {

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
	// A default implementation is provided: [AdjacencyMap].
	AdjacencyMap() (map[K]map[K]Edge[K, E], error)
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
	// A default implementation is provided: [PredecessorMap].
	PredecessorMap() (map[K]map[K]Edge[K, E], error)
}

func PredecessorMap[K comparable, V any, E any](g GraphRead[K, V, E]) (map[K]map[K]Edge[K, E], error) {
	adj := make(map[K]map[K]Edge[K, E])
	for v, err := range g.Vertices() {
		if err != nil {
			return nil, err
		}
		adj[g.Hash(v.Value)] = make(map[K]Edge[K, E])
	}
	for e, err := range g.Edges() {
		if err != nil {
			return nil, err
		}
		if _, ok := adj[e.Target]; !ok {
			adj[e.Target] = make(map[K]Edge[K, E])
		}
		adj[e.Target][e.Source] = e
	}
	return adj, nil
}

func (d DefaultGraph[K, V, E]) AdjacencyMap() (map[K]map[K]Edge[K, E], error) {
	return AdjacencyMap(d.g)
}

func AdjacencyMap[K comparable, V any, E any](g GraphRead[K, V, E]) (map[K]map[K]Edge[K, E], error) {
	adj := make(map[K]map[K]Edge[K, E])
	for v, err := range g.Vertices() {
		if err != nil {
			return nil, err
		}
		adj[g.Hash(v.Value)] = make(map[K]Edge[K, E])
	}
	for e, err := range g.Edges() {
		if err != nil {
			return nil, err
		}
		if _, ok := adj[e.Source]; !ok {
			adj[e.Source] = make(map[K]Edge[K, E])
		}
		adj[e.Source][e.Target] = e
	}
	return adj, nil
}

func (d DefaultGraph[K, V, E]) PredecessorMap() (map[K]map[K]Edge[K, E], error) {
	return PredecessorMap(d.g)
}
