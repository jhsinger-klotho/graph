package graph

type GraphBulkInserter[K comparable, V any, E any] interface {
	// AddVertices adds multiple vertices to the graph at once. If any of the
	// vertices already exists, ErrVertexAlreadyExists will be returned.
	// This is an atomic operation, meaning that if any of the vertices already
	// exists, none of the vertices will be added. If any options are specified,
	// they will be applied to all vertices.
	AddVertices(values []V, options ...func(*VertexProperties)) error

	// AddEdges adds multiple edges to the graph at once. If any of the edges
	// already exists, ErrEdgeAlreadyExists will be returned. This is an atomic
	// operation, meaning that if any of the edges already exists, none of the
	// edges will be added.
	// 'edges' is a map of source vertex hashes to target vertex hashes.
	// If any options are specified, they will be applied to all edges.
	AddEdges(edges map[K][]K, options ...func(*EdgeProperties[E])) error
}

// BulkAddVertices adds multiple vertices to the graph at once. If any of the
// vertices already exists, ErrVertexAlreadyExists will be returned.
// If supported by the graph, this is an atomic operation, meaning that if any
// of the vertices already exists, none of the vertices will be added.
func BulkAddVertices[K comparable, V any, E any](g GraphWrite[K, V, E], vertices []V, options ...func(*VertexProperties)) error {
	for _, v := range vertices {
		if err := g.AddVertex(v, options...); err != nil {
			return err
		}
	}
	return nil
}

// BulkAddEdges adds multiple edges to the graph at once. If any of the edges
// already exists, ErrEdgeAlreadyExists will be returned.
// If supported by the graph, this is an atomic operation, meaning that if any
// of the edges already exists, none of the edges will be added.
func BulkAddEdges[K comparable, V any, E any](g GraphWrite[K, V, E], edges map[K][]K, options ...func(*EdgeProperties[E])) error {
	for source, targets := range edges {
		for _, target := range targets {
			if err := g.AddEdge(source, target, options...); err != nil {
				return err
			}
		}
	}
	return nil
}
