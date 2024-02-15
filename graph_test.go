package graph

import (
	"testing"
)

func TestStringHash(t *testing.T) {
	tests := map[string]struct {
		value        string
		expectedHash string
	}{
		"string value": {
			value:        "London",
			expectedHash: "London",
		},
	}

	for name, test := range tests {
		hash := StringHash(test.value)

		if hash != test.expectedHash {
			t.Errorf("%s: hash expectancy doesn't match: expected %v, got %v", name, test.expectedHash, hash)
		}
	}
}

func TestIntHash(t *testing.T) {
	tests := map[string]struct {
		value        int
		expectedHash int
	}{
		"int value": {
			value:        3,
			expectedHash: 3,
		},
	}

	for name, test := range tests {
		hash := IntHash(test.value)

		if hash != test.expectedHash {
			t.Errorf("%s: hash expectancy doesn't match: expected %v, got %v", name, test.expectedHash, hash)
		}
	}
}

func TestEdgeWeight(t *testing.T) {
	tests := map[string]struct {
		expected EdgeProperties
		weight   int
	}{
		"weight 4": {
			weight: 4,
			expected: EdgeProperties{
				Weight: 4,
			},
		},
	}

	for name, test := range tests {
		properties := EdgeProperties{}

		EdgeWeight(test.weight)(&properties)

		if properties.Weight != test.expected.Weight {
			t.Errorf("%s: weight expectation doesn't match: expected %v, got %v", name, test.expected.Weight, properties.Weight)
		}
	}
}

func TestEdgeAttribute(t *testing.T) {
	tests := map[string]struct {
		key      string
		value    string
		expected EdgeProperties
	}{
		"attribute label=my-label": {
			key:   "label",
			value: "my-label",
			expected: EdgeProperties{
				Attributes: map[string]string{
					"label": "my-label",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			properties := EdgeProperties{
				Attributes: make(map[string]string),
			}

			EdgeAttribute(test.key, test.value)(&properties)

			value, ok := properties.Attributes[test.key]
			if !ok {
				t.Errorf("attribute expectaton doesn't match: key %v doesn't exist", test.key)
			}

			expectedValue := test.expected.Attributes[test.key]

			if value != expectedValue {
				t.Errorf("value expectation doesn't match: expected %v, got %v", expectedValue, value)
			}
		})

	}
}

func TestEdgeAttributes(t *testing.T) {
	tests := map[string]struct {
		attributes map[string]string
		expected   map[string]string
	}{
		"attribute label=my-label": {
			attributes: map[string]string{
				"label": "my-label",
			},
			expected: map[string]string{
				"label": "my-label",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			properties := EdgeProperties{
				Attributes: make(map[string]string),
			}

			EdgeAttributes(test.attributes)(&properties)

			if !mapsAreEqual(test.expected, properties.Attributes) {
				t.Errorf("expected %v, got %v", test.expected, properties.Attributes)
			}
		})
	}
}

func TestVertexAttribute(t *testing.T) {
	tests := map[string]struct {
		key      string
		value    string
		expected VertexProperties
	}{
		"attribute label=my-label": {
			key:   "label",
			value: "my-label",
			expected: VertexProperties{
				Attributes: map[string]string{
					"label": "my-label",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			properties := VertexProperties{
				Attributes: make(map[string]string),
			}

			VertexAttribute(test.key, test.value)(&properties)

			value, ok := properties.Attributes[test.key]
			if !ok {
				t.Errorf("attribute expectaton doesn't match: key %v doesn't exist", test.key)
			}

			expectedValue := test.expected.Attributes[test.key]

			if value != expectedValue {
				t.Errorf("value expectation doesn't match: expected %v, got %v", expectedValue, value)
			}
		})

	}
}

func TestVertexAttributes(t *testing.T) {
	tests := map[string]struct {
		attributes map[string]string
		expected   map[string]string
	}{
		"attribute label=my-label": {
			attributes: map[string]string{
				"label": "my-label",
			},
			expected: map[string]string{
				"label": "my-label",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			properties := VertexProperties{
				Attributes: make(map[string]string),
			}

			VertexAttributes(test.attributes)(&properties)

			if !mapsAreEqual(test.expected, properties.Attributes) {
				t.Errorf("expected %v, got %v", test.expected, properties.Attributes)
			}
		})
	}
}

func TestEdgesEqual(t *testing.T) {
	tests := map[string]struct {
		a             Edge[int]
		b             Edge[int]
		edgesAreEqual bool
	}{
		"equal edges in directed graph": {
			a:             Edge[int]{Source: 1, Target: 2},
			b:             Edge[int]{Source: 1, Target: 2},
			edgesAreEqual: true,
		},
		"swapped equal edges in directed graph": {
			a: Edge[int]{Source: 1, Target: 2},
			b: Edge[int]{Source: 2, Target: 1},
		},
	}

	for name, test := range tests {
		actual := EdgesEqual(IntHash, test.a, test.b)

		if actual != test.edgesAreEqual {
			t.Errorf("%s: equality expectations don't match: expected %v, got %v", name, test.edgesAreEqual, actual)
		}
	}
}

func mapsAreEqual[K comparable](a, b map[K]K) bool {
	for aHash, aValue := range a {
		bValue, ok := b[aHash]
		if !ok {
			return false
		}

		if aValue != bValue {
			return false
		}
	}

	for aHash := range a {
		if _, ok := b[aHash]; !ok {
			return false
		}
	}

	return true
}
