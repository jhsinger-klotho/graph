package graph

type GraphBulkInserter[K comparable, V any, E any] interface {
	// AddVertices adds multiple vertices to the graph at once. If any of the
	// vertices already exists, ErrVertexAlreadyExists will be returned.
	// This is an atomic operation, meaning that if any of the vertices already
	// exists, none of the vertices will be added.
	AddVertices(values []Vertex[V]) error

	// AddEdges adds multiple edges to the graph at once. If any of the edges
	// already exists, ErrEdgeAlreadyExists will be returned. This is an atomic
	// operation, meaning that if any of the edges already exists, none of the
	// edges will be added.
	AddEdges(edges []Edge[K, E]) error
}

// BulkAddVertices adds multiple vertices to the graph at once. If any of the
// vertices already exists, ErrVertexAlreadyExists will be returned.
// If supported by the graph, this is an atomic operation, meaning that if any
// of the vertices already exists, none of the vertices will be added.
func BulkAddVertices[K comparable, V any, E any](g GraphWrite[K, V, E], vertices []V) error {
	for _, v := range vertices {
		if err := g.AddVertex(v); err != nil {
			return err
		}
	}
	return nil
}

func (d DefaultGraph[K, V, E]) BulkAddVertices(values []V) error {
	return BulkAddVertices[K, V, E](d, values)
}

// BulkAddEdges adds multiple edges to the graph at once. If any of the edges
// already exists, ErrEdgeAlreadyExists will be returned.
// If supported by the graph, this is an atomic operation, meaning that if any
// of the edges already exists, none of the edges will be added.
func BulkAddEdges[K comparable, V any, E any](g GraphWrite[K, V, E], edges map[K][]K) error {
	for source, targets := range edges {
		for _, target := range targets {
			if err := g.AddEdge(source, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d DefaultGraph[K, V, E]) BulkAddEdges(edges map[K][]K) error {
	return BulkAddEdges[K, V, E](d, edges)
}
