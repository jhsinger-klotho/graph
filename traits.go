package graph

type Traits struct {
	IsDirected    bool
	PreventCycles bool
}

func Directed() func(*Traits) {
	return func(t *Traits) {
		t.IsDirected = true
	}
}

func PreventCycles() func(*Traits) {
	return func(t *Traits) {
		t.PreventCycles = true
	}
}
