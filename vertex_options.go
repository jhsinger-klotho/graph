package graph

// VertexWeight returns a function that sets the weight of a vertex to the given
// weight. This is a functional option for the [graph.Graph.Vertex] and
// [graph.Graph.AddVertex] methods.
func VertexWeight(weight float64) func(*VertexProperties) {
	return func(e *VertexProperties) {
		e.Weight = weight
	}
}

// VertexAttribute returns a function that adds the given key-value pair to the
// vertex attributes. This is a functional option for the [graph.Graph.Vertex]
// and [graph.Graph.AddVertex] methods.
func VertexAttribute(key, value string) func(*VertexProperties) {
	return func(e *VertexProperties) {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		e.Attributes[key] = value
	}
}

// VertexAttributes returns a function that sets the given map as the attributes
// of a vertex. This is a functional option for the [graph.Graph.AddVertex] methods.
func VertexAttributes(attributes map[string]string) func(*VertexProperties) {
	return func(e *VertexProperties) {
		e.Attributes = attributes
	}
}

// VertexCopyProperties makes a copy of the given properties and returns
// a 'option'-style function that can be used in the [graph.Graph.AddVertex] and
// [graph.Graph.UpdateVertex] methods.
func VertexCopyProperties(properties VertexProperties) func(*VertexProperties) {
	return func(e *VertexProperties) {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		for k, v := range properties.Attributes {
			e.Attributes[k] = v
		}
		e.Weight = properties.Weight
	}
}

// VertexCopy returns the given vertex and a function that can be used to copy
// which can be used as arguments to [graph.Graph.AddVertex]
//
//	err := g.AddVertex(VertexCopy(v))
func VertexCopy[T any](v Vertex[T]) (T, func(*VertexProperties)) {
	return v.Value, VertexCopyProperties(v.Properties)
}
