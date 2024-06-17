package graph

type (
	// DefaultGraph is a wrapper around a minimal Graph (read & write) that provides default implementations for all derivable methods.
	//
	// Embed it in your own graph implementation to get all the default implementations. For example:
	//
	//  type MyGraph struct {
	//    DefaultGraph
	//    ...
	//  }
	//
	//  func NewMyGraph() *MyGraph {
	//    g := &MyGraph{...}
	//    g.DefaultGraph = NewDefaultGraph(g)
	//    return g
	//  }
	DefaultGraph[K comparable, V any, E any] struct {
		g interface {
			GraphRead[K, V, E]
			GraphWrite[K, V, E]
		}
	}
)

func NewDefaultGraph[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphWrite[K, V, E]
}) DefaultGraph[K, V, E] {
	return DefaultGraph[K, V, E]{g: g}
}

// Note: only the passthrough methods are in this file, the rest are in the other files next to their implementations.

func (d DefaultGraph[K, V, E]) Hash(v V) K {
	return d.g.Hash(v)
}

func (d DefaultGraph[K, V, E]) Traits() Traits {
	return d.g.Traits()
}

func (d DefaultGraph[K, V, E]) Vertex(hash K) (Vertex[V], error) {
	return d.g.Vertex(hash)
}

func (d DefaultGraph[K, V, E]) Vertices() VertexIter[V] {
	return d.g.Vertices()
}

func (d DefaultGraph[K, V, E]) Edge(sourceHash, targetHash K) (Edge[K, E], error) {
	return d.g.Edge(sourceHash, targetHash)
}

func (d DefaultGraph[K, V, E]) Edges() EdgeIter[K, E] {
	return d.g.Edges()
}

func (d DefaultGraph[K, V, E]) Order() (int, error) {
	return d.g.Order()
}

func (d DefaultGraph[K, V, E]) Size() (int, error) {
	return d.g.Size()
}

func (d DefaultGraph[K, V, E]) AddVertex(value V, options ...func(*VertexProperties)) error {
	return d.g.AddVertex(value, options...)
}

func (d DefaultGraph[K, V, E]) UpdateVertex(hash K, options ...func(*Vertex[V])) error {
	return d.g.UpdateVertex(hash, options...)
}

func (d DefaultGraph[K, V, E]) RemoveVertex(hash K) error {
	return d.g.RemoveVertex(hash)
}

func (d DefaultGraph[K, V, E]) AddEdge(sourceHash, targetHash K, options ...func(*EdgeProperties[E])) error {
	return d.g.AddEdge(sourceHash, targetHash, options...)
}

func (d DefaultGraph[K, V, E]) UpdateEdge(source, target K, options ...func(properties *EdgeProperties[E])) error {
	return d.g.UpdateEdge(source, target, options...)
}

func (d DefaultGraph[K, V, E]) RemoveEdge(source, target K) error {
	return d.g.RemoveEdge(source, target)
}
