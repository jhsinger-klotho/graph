package graph

import (
	"reflect"
	"sort"
	"testing"
)

func TestDirectedCreatesCycle(t *testing.T) {
	// A wrapper type to passthrough read operations, but not other methods (to skip the interface check for custom impl)
	type RO struct {
		GraphRead[int, int, any]
	}
	tests := map[string]struct {
		vertices     []int
		edges        []Edge[int, any]
		sourceHash   int
		targetHash   int
		createsCycle bool
	}{
		"directed 2-4-7-5 cycle": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
				{Source: 5, Target: 2},
			},
			sourceHash:   7,
			targetHash:   5,
			createsCycle: true,
		},
		"undirected 2-4-7-5 'cycle'": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
				{Source: 5, Target: 2},
			},
			sourceHash: 5,
			targetHash: 7,
			// The direction of the edge (57 instead of 75) doesn't create a directed cycle.
			createsCycle: false,
		},
		"no cycle": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
				{Source: 5, Target: 2},
			},
			sourceHash:   5,
			targetHash:   6,
			createsCycle: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := RO{GraphRead: newTestGraph(test.vertices, test.edges)}

			createsCycle, err := CreatesCycle[int, int](graph, test.sourceHash, test.targetHash)
			if err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}

			if createsCycle != test.createsCycle {
				t.Errorf("%s: cycle expectancy doesn't match: expected %v, got %v", name, test.createsCycle, createsCycle)
			}
		})
	}
}
func TestUndirectedCreatesCycle(t *testing.T) {
	type ReadRelation interface {
		GraphRead[int, int, any]
		GraphRelations[int, any]
	}
	// A wrapper type to passthrough read operations, but not other methods (to skip the interface check for custom impl)
	type RO struct {
		ReadRelation
	}
	tests := map[string]struct {
		vertices     []int
		edges        []Edge[int, any]
		sourceHash   int
		targetHash   int
		createsCycle bool
	}{
		"undirected 2-4-7-5 cycle": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
				{Source: 5, Target: 7},
			},
			sourceHash:   2,
			targetHash:   5,
			createsCycle: true,
		},
		"undirected 5-6-3-1-2-7 cycle": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
				{Source: 5, Target: 7},
			},
			sourceHash:   5,
			targetHash:   6,
			createsCycle: true,
		},
		"no cycle": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 6},
				{Source: 4, Target: 7},
			},
			sourceHash:   5,
			targetHash:   7,
			createsCycle: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := RO{ReadRelation: newUndirectedTestGraph(test.vertices, test.edges)}

			createsCycle, err := CreatesCycle(graph, test.sourceHash, test.targetHash)
			if err != nil {
				t.Fatalf("%s: failed to add edge: %s", name, err.Error())
			}

			if createsCycle != test.createsCycle {
				t.Errorf("%s: cycle expectancy doesn't match: expected %v, got %v", name, test.createsCycle, createsCycle)
			}
		})
	}
}

func TestDirectedShortestPath(t *testing.T) {
	tests := map[string]struct {
		vertices             []string
		vertexWeights        map[string]float64
		edges                []Edge[string, any]
		sourceHash           string
		targetHash           string
		expectedShortestPath []string
		shouldFail           bool
	}{
		"graph as on img/dijkstra.svg": {
			vertices: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "A", Target: "F", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "C", Target: "E", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "C", Target: "F", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "D", Target: "B", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "E", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "E", Target: "F", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "F", Target: "G", Properties: EdgeProperties[any]{Weight: 5}},
				{Source: "G", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
			},
			sourceHash:           "A",
			targetHash:           "B",
			expectedShortestPath: []string{"A", "C", "E", "B"},
		},
		"diamond-shaped graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
			},
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{"A", "B", "D"},
		},
		"unweighted graph": {
			vertices: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{}},
				{Source: "C", Target: "F", Properties: EdgeProperties[any]{}},
				{Source: "D", Target: "G", Properties: EdgeProperties[any]{}},
				{Source: "E", Target: "G", Properties: EdgeProperties[any]{}},
				{Source: "F", Target: "E", Properties: EdgeProperties[any]{}},
			},
			sourceHash:           "A",
			targetHash:           "G",
			expectedShortestPath: []string{"A", "B", "D", "G"},
		},
		"source equal to target": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
			},
			sourceHash:           "B",
			targetHash:           "B",
			expectedShortestPath: []string{"B"},
		},
		"target not reachable in a disconnected graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
			},
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{},
			shouldFail:           true,
		},
		"target not reachable in a connected graph": {
			vertices: []string{"A", "B", "C"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 0}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 0}},
			},
			sourceHash:           "B",
			targetHash:           "C",
			expectedShortestPath: []string{},
			shouldFail:           true,
		},
		"graph from issue 88": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 6}},
				{Source: "B", Target: "C", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 5}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 1}},
			},
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{"A", "B", "C", "D"},
		},
		"can process negative weights": {
			vertices: []string{"A", "B", "C", "D", "E"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "B", Target: "C", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "E", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "D", Target: "E", Properties: EdgeProperties[any]{Weight: -1}},
			},
			sourceHash:           "A",
			targetHash:           "E",
			expectedShortestPath: []string{"A", "B", "D", "E"},
		},
		"vertex weights": {
			vertices: []string{"A", "B", "C", "D"},
			vertexWeights: map[string]float64{
				"A": 0,
				"B": 1,
				"C": 2,
				"D": 0,
			},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B"},
				{Source: "A", Target: "C"},
				{Source: "B", Target: "D"},
				{Source: "C", Target: "D"},
			},
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{"A", "B", "D"},
		},
		"vertex and edge weights": {
			// like the img/dijkstra.svg graph, but with vertex weights to make the shortest path different
			vertices: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "A", Target: "F", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "C", Target: "E", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "C", Target: "F", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "D", Target: "B", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "E", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "E", Target: "F", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "F", Target: "G", Properties: EdgeProperties[any]{Weight: 5}},
				{Source: "G", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
			},
			vertexWeights: map[string]float64{
				"A": 1,
				"B": 1,
				"C": 1,
				"D": 1,
				"E": 10,
				"F": 1,
				"G": 1,
			},
			sourceHash:           "A",
			targetHash:           "B",
			expectedShortestPath: []string{"A", "C", "D", "B"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := NewMemoryGraph[string, string, any](StringHash, Directed())

			for _, vertex := range test.vertices {
				if w, ok := test.vertexWeights[vertex]; ok {
					_ = graph.AddVertex(vertex, VertexWeight(w))
				} else {
					_ = graph.AddVertex(vertex)
				}
			}

			for _, edge := range test.edges {
				if err := graph.AddEdge(EdgeCopy(edge)); err != nil {
					t.Fatalf("%s: failed to add edge: %s", name, err.Error())
				}
			}

			shortestPath, err := ShortestPath[string, string](graph, test.sourceHash, test.targetHash)

			if test.shouldFail != (err != nil) {
				t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v (error: %v)", name, test.shouldFail, (err != nil), err)
			}

			if len(shortestPath) != len(test.expectedShortestPath) {
				t.Fatalf("%s: path length expectancy doesn't match: expected %v, got %v", name, len(test.expectedShortestPath), len(shortestPath))
			}

			for i, expectedVertex := range test.expectedShortestPath {
				if shortestPath[i] != expectedVertex {
					t.Errorf("%s: path vertex expectancy doesn't match: expected %v at index %d, got %v", name, expectedVertex, i, shortestPath[i])
				}
			}
		})
	}
}

func TestUndirectedShortestPath(t *testing.T) {
	tests := map[string]struct {
		vertices             []string
		edges                []Edge[string, any]
		sourceHash           string
		targetHash           string
		isWeighted           bool
		expectedShortestPath []string
		shouldFail           bool
	}{
		"graph as on img/dijkstra.svg": {
			vertices: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "A", Target: "F", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "C", Target: "E", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "C", Target: "F", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "D", Target: "B", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "E", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "E", Target: "F", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "F", Target: "G", Properties: EdgeProperties[any]{Weight: 5}},
				{Source: "G", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
			},
			isWeighted:           true,
			sourceHash:           "A",
			targetHash:           "B",
			expectedShortestPath: []string{"A", "C", "E", "B"},
		},
		"diamond-shaped graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
			},
			isWeighted:           true,
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{"A", "B", "D"},
		},
		"unweighted graph": {
			vertices: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{}},
				{Source: "C", Target: "F", Properties: EdgeProperties[any]{}},
				{Source: "D", Target: "G", Properties: EdgeProperties[any]{}},
				{Source: "E", Target: "G", Properties: EdgeProperties[any]{}},
				{Source: "F", Target: "E", Properties: EdgeProperties[any]{}},
			},
			sourceHash:           "A",
			targetHash:           "G",
			expectedShortestPath: []string{"A", "B", "D", "G"},
		},
		"source equal to target": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
			},
			isWeighted:           true,
			sourceHash:           "B",
			targetHash:           "B",
			expectedShortestPath: []string{"B"},
		},
		"target not reachable in a disconnected graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
			},
			isWeighted:           true,
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{},
			shouldFail:           true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := NewMemoryGraph[string, string, any](StringHash)

			for _, vertex := range test.vertices {
				_ = graph.AddVertex(vertex)
			}

			for _, edge := range test.edges {
				if err := graph.AddEdge(edge.Source, edge.Target, EdgeWeight[any](edge.Properties.Weight)); err != nil {
					t.Fatalf("%s: failed to add edge: %s", name, err.Error())
				}
			}

			shortestPath, err := ShortestPath[string, string](graph, test.sourceHash, test.targetHash)

			if test.shouldFail != (err != nil) {
				t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v (error: %v)", name, test.shouldFail, (err != nil), err)
			}

			if len(shortestPath) != len(test.expectedShortestPath) {
				t.Fatalf("%s: path length expectancy doesn't match: expected %v, got %v", name, len(test.expectedShortestPath), len(shortestPath))
			}

			for i, expectedVertex := range test.expectedShortestPath {
				if shortestPath[i] != expectedVertex {
					t.Errorf("%s: path vertex expectancy doesn't match: expected %v at index %d, got %v", name, expectedVertex, i, shortestPath[i])
				}
			}
		})
	}
}

func Test_BellmanFord(t *testing.T) {
	tests := map[string]struct {
		vertices             []string
		edges                []Edge[string, any]
		sourceHash           string
		targetHash           string
		isWeighted           bool
		IsDirected           bool
		expectedShortestPath []string
		shouldFail           bool
	}{
		"graph as on img/dijkstra.svg": {
			vertices: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: []Edge[string, any]{

				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "A", Target: "F", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "C", Target: "E", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "C", Target: "F", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "D", Target: "B", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "E", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "E", Target: "F", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "F", Target: "G", Properties: EdgeProperties[any]{Weight: 5}},
				{Source: "G", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
			},
			isWeighted:           true,
			IsDirected:           true,
			sourceHash:           "A",
			targetHash:           "B",
			expectedShortestPath: []string{"A", "C", "E", "B"},
		},
		"diamond-shaped graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
			},
			isWeighted:           true,
			IsDirected:           true,
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{"A", "B", "D"},
		},
		"unweighted graph": {
			vertices: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{}},
				{Source: "C", Target: "F", Properties: EdgeProperties[any]{}},
				{Source: "D", Target: "G", Properties: EdgeProperties[any]{}},
				{Source: "E", Target: "G", Properties: EdgeProperties[any]{}},
				{Source: "F", Target: "E", Properties: EdgeProperties[any]{}},
			},
			IsDirected:           true,
			sourceHash:           "A",
			targetHash:           "G",
			expectedShortestPath: []string{"A", "B", "D", "G"},
		},
		"source equal to target": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
			},
			isWeighted:           true,
			IsDirected:           true,
			sourceHash:           "B",
			targetHash:           "B",
			expectedShortestPath: []string{"B"},
		},
		"target not reachable in a disconnected graph": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
			},
			isWeighted:           true,
			IsDirected:           true,
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{},
			shouldFail:           true,
		},
		"negative weights graph": {
			vertices: []string{"A", "B", "C", "D", "E"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "B", Target: "C", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "C", Target: "E", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "D", Target: "E", Properties: EdgeProperties[any]{Weight: -1}},
			},
			isWeighted:           true,
			IsDirected:           true,
			sourceHash:           "A",
			targetHash:           "E",
			expectedShortestPath: []string{"A", "B", "D", "E"},
		},
		"fails on negative cycles": {
			vertices: []string{"A", "B", "C", "D", "E"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 1}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
				{Source: "B", Target: "C", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "B", Target: "D", Properties: EdgeProperties[any]{Weight: 6}},
				{Source: "C", Target: "D", Properties: EdgeProperties[any]{Weight: 3}},
				{Source: "C", Target: "E", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "D", Target: "E", Properties: EdgeProperties[any]{Weight: -3}},
				{Source: "E", Target: "C", Properties: EdgeProperties[any]{Weight: -3}},
			},
			isWeighted:           true,
			IsDirected:           true,
			sourceHash:           "A",
			targetHash:           "E",
			expectedShortestPath: []string{},
			shouldFail:           true,
		},
		"fails if not directed": {
			vertices: []string{"A", "B", "C", "D"},
			edges: []Edge[string, any]{
				{Source: "A", Target: "B", Properties: EdgeProperties[any]{Weight: 2}},
				{Source: "A", Target: "C", Properties: EdgeProperties[any]{Weight: 4}},
			},
			isWeighted:           true,
			sourceHash:           "A",
			targetHash:           "D",
			expectedShortestPath: []string{},
			shouldFail:           true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := NewMemoryGraph[string, string, any](StringHash, Directed())

			for _, vertex := range test.vertices {
				_ = graph.AddVertex(vertex)
			}

			for _, edge := range test.edges {
				if err := graph.AddEdge(edge.Source, edge.Target, EdgeWeight[any](edge.Properties.Weight)); err != nil {
					t.Fatalf("%s: failed to add edge: %s", name, err.Error())
				}
			}

			shortestPath, err := BellmanFordShortestPath[string, string](graph, test.sourceHash, nil).ShortestPath(test.targetHash)

			if test.shouldFail != (err != nil) {
				t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v (error: %v)", name, test.shouldFail, (err != nil), err)
			}

			if len(shortestPath) != len(test.expectedShortestPath) {
				t.Fatalf("%s: path length expectancy doesn't match: expected %v, got %v", name, len(test.expectedShortestPath), len(shortestPath))
			}

			for i, expectedVertex := range test.expectedShortestPath {
				if shortestPath[i] != expectedVertex {
					t.Errorf("%s: path vertex expectancy doesn't match: expected %v at index %d, got %v", name, expectedVertex, i, shortestPath[i])
				}
			}
		})
	}
}

func TestDirectedStronglyConnectedComponents(t *testing.T) {
	tests := map[string]struct {
		vertices     []int
		edges        []Edge[int, any]
		expectedSCCs [][]int
	}{
		"graph with SCCs as on img/scc.svg": {
			vertices: []int{1, 2, 3, 4, 5, 6, 7, 8},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 2, Target: 5},
				{Source: 2, Target: 6},
				{Source: 3, Target: 4},
				{Source: 3, Target: 7},
				{Source: 4, Target: 3},
				{Source: 4, Target: 8},
				{Source: 5, Target: 1},
				{Source: 5, Target: 6},
				{Source: 6, Target: 7},
				{Source: 7, Target: 6},
				{Source: 8, Target: 4},
				{Source: 8, Target: 7},
			},
			expectedSCCs: [][]int{{1, 2, 5}, {3, 4, 8}, {6, 7}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := newTestGraph(test.vertices, test.edges)

			sccs, _ := StronglyConnectedComponents[int, int](graph)
			matchedSCCs := 0

			for _, scc := range sccs {
				for _, expectedSCC := range test.expectedSCCs {
					if slicesAreEqual(scc, expectedSCC) {
						matchedSCCs++
					}
				}
			}

			if matchedSCCs != len(test.expectedSCCs) {
				t.Errorf("%s: expected SCCs don't match: expected %v, got %v", name, test.expectedSCCs, sccs)
			}
		})
	}
}

func TestAllPathsBetween(t *testing.T) {
	type args struct {
		g     Graph[int, int, any]
		start int
		end   int
	}
	type testCase[K comparable, T any] struct {
		name    string
		args    args
		want    [][]K
		wantErr bool
	}
	tests := []testCase[int, int]{
		{
			name: "directed",
			args: args{
				g: func() Graph[int, int, any] {
					g := NewMemoryGraph[int, int, any](IntHash, Directed())
					for i := 0; i <= 8; i++ {
						_ = g.AddVertex(i)
					}
					_ = g.AddEdge(0, 2)
					_ = g.AddEdge(1, 0)
					_ = g.AddEdge(1, 4)
					_ = g.AddEdge(2, 6)
					_ = g.AddEdge(3, 1)
					_ = g.AddEdge(3, 7)
					_ = g.AddEdge(4, 5)
					_ = g.AddEdge(5, 2)
					_ = g.AddEdge(5, 6)
					_ = g.AddEdge(6, 8)
					_ = g.AddEdge(7, 4)
					return g
				}(),
				start: 3,
				end:   6,
			},
			want: [][]int{
				{3, 1, 0, 2, 6},
				{3, 1, 4, 5, 6},
				{3, 1, 4, 5, 2, 6},
				{3, 7, 4, 5, 2, 6},
				{3, 7, 4, 5, 6},
			},
			wantErr: false,
		},
		{
			name: "undirected",
			args: args{
				g: func() Graph[int, int, any] {
					g := NewMemoryGraph[int, int, any](IntHash)
					for i := 0; i <= 8; i++ {
						_ = g.AddVertex(i)
					}
					_ = g.AddEdge(0, 1)
					_ = g.AddEdge(0, 2)
					_ = g.AddEdge(1, 3)
					_ = g.AddEdge(1, 4)
					_ = g.AddEdge(2, 5)
					_ = g.AddEdge(2, 6)
					_ = g.AddEdge(3, 7)
					_ = g.AddEdge(4, 5)
					_ = g.AddEdge(4, 7)
					_ = g.AddEdge(5, 6)
					_ = g.AddEdge(6, 8)
					return g
				}(),
				start: 3,
				end:   6,
			},
			want: [][]int{
				{3, 1, 0, 2, 6},
				{3, 1, 0, 2, 5, 6},
				{3, 1, 4, 5, 6},
				{3, 1, 4, 5, 2, 6},
				{3, 7, 4, 5, 2, 6},
				{3, 7, 4, 5, 6},
				{3, 7, 4, 1, 0, 2, 6},
				{3, 7, 4, 1, 0, 2, 5, 6},
			},
			wantErr: false,
		},
		{
			name: "directed with cycle",
			args: args{
				g: func() Graph[int, int, any] {
					g := NewMemoryGraph[int, int, any](IntHash, Directed())
					for i := 0; i <= 8; i++ {
						_ = g.AddVertex(i)
					}
					_ = g.AddEdge(0, 1)
					_ = g.AddEdge(1, 2)
					_ = g.AddEdge(2, 3)
					_ = g.AddEdge(2, 0)
					_ = g.AddEdge(3, 0)
					return g
				}(),
				start: 0,
				end:   0,
			},
			want: [][]int{
				{0, 1, 2, 3, 0},
				{0, 1, 2, 0},
			},
			wantErr: false,
		},
		{
			name: "directed with self cycle",
			args: args{
				g: func() Graph[int, int, any] {
					g := NewMemoryGraph[int, int, any](IntHash, Directed())
					for i := 0; i <= 8; i++ {
						_ = g.AddVertex(i)
					}
					_ = g.AddEdge(0, 1)
					_ = g.AddEdge(0, 0)
					_ = g.AddEdge(1, 2)
					_ = g.AddEdge(2, 3)
					_ = g.AddEdge(2, 0)
					_ = g.AddEdge(3, 0)
					return g
				}(),
				start: 0,
				end:   0,
			},
			want: [][]int{
				{0, 1, 2, 3, 0},
				{0, 1, 2, 0},
				{0, 0},
			},
			wantErr: false,
		},
		{
			name: "directed with unuseable cycle",
			args: args{
				g: func() Graph[int, int, any] {
					g := NewMemoryGraph[int, int, any](IntHash, Directed())
					for i := 0; i <= 8; i++ {
						_ = g.AddVertex(i)
					}
					_ = g.AddEdge(0, 1)
					_ = g.AddEdge(1, 2)
					_ = g.AddEdge(2, 3)
					_ = g.AddEdge(3, 2)
					_ = g.AddEdge(0, 2)
					_ = g.AddEdge(3, 0)
					return g
				}(),
				start: 0,
				end:   2,
			},
			want: [][]int{
				{0, 1, 2},
				{0, 2},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AllPathsBetween(tt.args.g, tt.args.start, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllPathsBetween() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			toStr := func(s []int) string {
				var num string
				for _, n := range s {
					num = num + string(rune(n))
				}
				return num
			}

			sort.Slice(got, func(i, j int) bool {
				return toStr(got[i]) < toStr(got[j])
			})

			sort.Slice(tt.want, func(i, j int) bool {
				return toStr(tt.want[i]) < toStr(tt.want[j])
			})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AllPathsBetween() got = %v, want %v", got, tt.want)
			}
		})
	}
}
