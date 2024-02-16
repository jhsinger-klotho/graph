package graph

type Traits struct {
	IsDirected         bool
	PreventCycles      bool
	IsVerticesWeighted bool
	IsEdgesWeighted    bool
}

func Directed() func(*Traits) {
	return func(t *Traits) {
		t.IsDirected = true
	}
}

// PreventCycles specifies that cycles should be detected and
// rejected via a [EdgeCausesCycleError] error.
func PreventCycles() func(*Traits) {
	return func(t *Traits) {
		t.PreventCycles = true
	}
}

// VerticesWeighted manually specifies that the vertices
// of the graph are weighted. This can also be automatically
// detected by the graph implementation.
func VerticesWeighted() func(*Traits) {
	return func(t *Traits) {
		t.IsVerticesWeighted = true
	}
}

// EdgesWeighted manually specifies that the edges
// of the graph are weighted. This can also be automatically
// detected by the graph implementation.
func EdgesWeighted() func(*Traits) {
	return func(t *Traits) {
		t.IsEdgesWeighted = true
	}
}

func (t Traits) IsWeighted() (vertices bool, edges bool) {
	return t.IsVerticesWeighted, t.IsEdgesWeighted
}
