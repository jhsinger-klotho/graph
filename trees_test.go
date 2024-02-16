package graph

import (
	"fmt"
	"strings"
	"testing"
)

func TestUndirectedMinimumSpanningTree(t *testing.T) {
	tests := map[string]struct {
		vertices                []string
		edges                   []Edge[string]
		expectedErr             error
		expectedMSTAdjacencyMap map[string]map[string]Edge[string]
	}{
		"graph from img/mst.svg": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string]{
				{Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties{Weight: 4}},
				{Source: "A", Target: "D", Properties: EdgeProperties{Weight: 3}},
				{Source: "B", Target: "C", Properties: EdgeProperties{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties{Weight: 1}},
				{Source: "C", Target: "D", Properties: EdgeProperties{Weight: 3}},
			},
			expectedErr: nil,
			expectedMSTAdjacencyMap: map[string]map[string]Edge[string]{
				"A": {
					"B": {Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				},
				"B": {
					"D": {Source: "B", Target: "D", Properties: EdgeProperties{Weight: 1}},
					"A": {Source: "B", Target: "A", Properties: EdgeProperties{Weight: 2}},
				},
				"C": {
					"D": {Source: "C", Target: "D", Properties: EdgeProperties{Weight: 3}},
				},
				"D": {
					"B": {Source: "D", Target: "B", Properties: EdgeProperties{Weight: 1}},
					"C": {Source: "D", Target: "C", Properties: EdgeProperties{Weight: 3}},
				},
			},
		},
		"two trees for a disconnected graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string]{
				{Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties{Weight: 4}},
			},
			expectedErr: nil,
			expectedMSTAdjacencyMap: map[string]map[string]Edge[string]{
				"A": {
					"B": {Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				},
				"B": {
					"A": {Source: "B", Target: "A", Properties: EdgeProperties{Weight: 2}},
				},
				"C": {
					"D": {Source: "C", Target: "D", Properties: EdgeProperties{Weight: 4}},
				},
				"D": {
					"C": {Source: "D", Target: "C", Properties: EdgeProperties{Weight: 4}},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := newStringUndirectedTestGraph(test.vertices, test.edges)
			mst := newStringUndirectedTestGraph(nil, nil)

			_ = MinimumSpanningTree(g, mst)
			adjacencyMap, _ := mst.AdjacencyMap()

			edgesAreEqual := edgesEqualFunc(StringHash)

			if !adjacencyMapsAreEqual(test.expectedMSTAdjacencyMap, adjacencyMap, edgesAreEqual) {
				t.Fatalf("expected adjacency map %s, got %s", mapToString(test.expectedMSTAdjacencyMap), mapToString(adjacencyMap))
			}
		})
	}
}

func TestUndirectedMaximumSpanningTree(t *testing.T) {
	tests := map[string]struct {
		vertices                []string
		edges                   []Edge[string]
		expectedErr             error
		expectedMSTAdjacencyMap map[string]map[string]Edge[string]
	}{
		"graph from img/mst.svg with higher weights": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string]{
				{Source: "A", Target: "B", Properties: EdgeProperties{Weight: 20}},
				{Source: "A", Target: "C", Properties: EdgeProperties{Weight: 4}},
				{Source: "A", Target: "D", Properties: EdgeProperties{Weight: 3}},
				{Source: "B", Target: "C", Properties: EdgeProperties{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties{Weight: 10}},
				{Source: "C", Target: "D", Properties: EdgeProperties{Weight: 30}},
			},
			expectedErr: nil,
			expectedMSTAdjacencyMap: map[string]map[string]Edge[string]{
				"A": {
					"B": {Source: "A", Target: "B", Properties: EdgeProperties{Weight: 20}},
				},
				"B": {
					"D": {Source: "B", Target: "D", Properties: EdgeProperties{Weight: 10}},
					"A": {Source: "B", Target: "A", Properties: EdgeProperties{Weight: 20}},
				},
				"C": {
					"D": {Source: "C", Target: "D", Properties: EdgeProperties{Weight: 30}},
				},
				"D": {
					"B": {Source: "D", Target: "B", Properties: EdgeProperties{Weight: 10}},
					"C": {Source: "D", Target: "C", Properties: EdgeProperties{Weight: 30}},
				},
			},
		},
		"two trees for a disconnected graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string]{
				{Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties{Weight: 4}},
			},
			expectedErr: nil,
			expectedMSTAdjacencyMap: map[string]map[string]Edge[string]{
				"A": {
					"B": {Source: "A", Target: "B", Properties: EdgeProperties{Weight: 2}},
				},
				"B": {
					"A": {Source: "B", Target: "A", Properties: EdgeProperties{Weight: 2}},
				},
				"C": {
					"D": {Source: "C", Target: "D", Properties: EdgeProperties{Weight: 4}},
				},
				"D": {
					"C": {Source: "D", Target: "C", Properties: EdgeProperties{Weight: 4}},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := newStringUndirectedTestGraph(test.vertices, test.edges)
			mst := newStringUndirectedTestGraph(nil, nil)

			_ = MaximumSpanningTree(g, mst)
			adjacencyMap, _ := mst.AdjacencyMap()

			edgesAreEqual := edgesEqualFunc(StringHash)

			if !adjacencyMapsAreEqual(test.expectedMSTAdjacencyMap, adjacencyMap, edgesAreEqual) {
				t.Fatalf("expected adjacency map %v, got %v", mapToString(test.expectedMSTAdjacencyMap), mapToString(adjacencyMap))
			}
		})
	}
}

func adjacencyMapsAreEqual[K comparable](a, b map[K]map[K]Edge[K], edgesAreEqual func(a, b Edge[K]) bool) bool {
	for aHash, aAdjacencies := range a {
		bAdjacencies, ok := b[aHash]
		if !ok {
			return false
		}

		for aAdjacency, aEdge := range aAdjacencies {
			bEdge, ok := bAdjacencies[aAdjacency]
			if !ok {
				return false
			}

			if !edgesAreEqual(aEdge, bEdge) {
				return false
			}

			for aKey, aValue := range aEdge.Properties.Attributes {
				bValue, ok := bEdge.Properties.Attributes[aKey]
				if !ok {
					return false
				}
				if bValue != aValue {
					return false
				}
			}

			if bEdge.Properties.Weight != aEdge.Properties.Weight {
				return false
			}
		}
	}

	for aHash := range a {
		if _, ok := b[aHash]; !ok {
			return false
		}
	}

	return true
}

func mapToString(m map[string]map[string]Edge[string]) string {
	sb := strings.Builder{}
	sb.WriteString("{")
	for src, v := range m {
		if len(v) == 0 {
			comma := ""
			if sb.Len() > 1 {
				comma = ", "
			}
			fmt.Fprintf(&sb, "%s%s", comma, src)
		}
		for trg := range v {
			comma := ""
			if sb.Len() > 1 {
				comma = ", "
			}
			fmt.Fprintf(&sb, "%s%s -> %s", comma, src, trg)
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func newStringUndirectedTestGraph(vertices []string, edges []Edge[string]) *memoryGraph[string, string] {
	g := NewMemoryGraph(StringHash)

	for _, vertex := range vertices {
		_ = g.AddVertex(vertex)
	}

	for _, edge := range edges {
		_ = g.AddEdge(EdgeCopy(edge))
	}
	return g
}
