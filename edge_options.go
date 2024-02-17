package graph

// EdgeWeight returns a function that sets the weight of an edge to the given
// weight. This is a functional option for the [graph.Graph.Edge] and
// [graph.Graph.AddEdge] methods.
func EdgeWeight[E any](weight float64) func(*EdgeProperties[E]) {
	return func(e *EdgeProperties[E]) {
		e.Weight = weight
	}
}

// EdgeAttribute returns a function that adds the given key-value pair to the
// attributes of an edge. This is a functional option for the [graph.Graph.Edge]
// and [graph.Graph.AddEdge] methods.
func EdgeAttribute[E any](key, value string) func(*EdgeProperties[E]) {
	return func(e *EdgeProperties[E]) {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		e.Attributes[key] = value
	}
}

// EdgeAttributes returns a function that sets the given map as the attributes
// of an edge. This is a functional option for the [graph.Graph.AddEdge] and
// [graph.Graph.UpdateEdge] methods.
func EdgeAttributes[E any](attributes map[string]string) func(*EdgeProperties[E]) {
	return func(e *EdgeProperties[E]) {
		e.Attributes = attributes
	}
}

// EdgeData returns a function that sets the data of an edge to the given value.
// This is a functional option for the [graph.Graph.Edge] and
// [graph.Graph.AddEdge] methods.
func EdgeData[E any](data E) func(*EdgeProperties[E]) {
	return func(e *EdgeProperties[E]) {
		e.Data = data
	}
}

// EdgeCopyProperties makes a copy (shallow for .Data) of the given properties and returns a
// 'option'-style function that can be used in the [graph.Graph.AddEdge] and
// [graph.Graph.UpdateEdge] methods.
func EdgeCopyProperties[E any](properties EdgeProperties[E]) func(*EdgeProperties[E]) {
	return func(e *EdgeProperties[E]) {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		for k, v := range properties.Attributes {
			e.Attributes[k] = v
		}
		e.Weight = properties.Weight
		e.Data = properties.Data
	}
}

// EdgeCopy returns the given edge and a function that can be used to copy
// which can be used as arguments to [graph.Graph.AddEdge]
//
//	err := g.AddEdge(EdgeCopy(e))
func EdgeCopy[K comparable, E any](e Edge[K, E]) (K, K, func(*EdgeProperties[E])) {
	return e.Source, e.Target, EdgeCopyProperties(e.Properties)
}

func EdgesEqual[K comparable, T any, E any](hash Hash[K, T], a, b Edge[T, E]) bool {
	return hash(a.Source) == hash(b.Source) && hash(a.Target) == hash(b.Target)
}
