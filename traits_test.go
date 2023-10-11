package graph

import "testing"

func TestTraits(t *testing.T) {
	tests := map[string]struct {
		f        func(*Traits)
		expected *Traits
	}{
		"directed graph": {
			f: Directed(),
			expected: &Traits{
				IsDirected: true,
			},
		},
		"acyclic graph": {
			f: Acyclic(),
			expected: &Traits{
				IsAcyclic: true,
			},
		},
		"weighted graph": {
			f: Weighted(),
			expected: &Traits{
				IsWeighted: true,
			},
		},
		"rooted graph": {
			f: Rooted(),
			expected: &Traits{
				IsRooted: true,
			},
		},
		"tree graph": {
			f: Tree(),
			expected: &Traits{
				IsAcyclic: true,
				IsRooted:  true,
			},
		},
		"prevent cycles": {
			f: PreventCycles(),
			expected: &Traits{
				IsAcyclic:     true,
				PreventCycles: true,
			},
		},
		"allow duplicate add": {
			f: AllowDuplicateAdd(),
			expected: &Traits{
				AllowDuplicateAdd: true,
			},
		},
	}

	for name, test := range tests {
		p := &Traits{}

		test.f(p)

		if !traitsAreEqual(test.expected, p) {
			t.Errorf("%s: trait expectation doesn't match: expected %v, got %v", name, test.expected, p)
		}
	}
}

func traitsAreEqual(a, b *Traits) bool {
	return a.IsAcyclic == b.IsAcyclic &&
		a.IsDirected == b.IsDirected &&
		a.IsRooted == b.IsRooted &&
		a.IsWeighted == b.IsWeighted &&
		a.PreventCycles == b.PreventCycles
}
