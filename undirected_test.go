package graph

import (
	"errors"
	"testing"
)

func TestUndirected_AddVertex(t *testing.T) {
	tests := map[string]struct {
		vertices           []int
		properties         *VertexProperties
		expectedVertices   []int
		expectedProperties *VertexProperties
		// Even though some AddVertex calls might work, at least one of them
		// could fail, e.g. if the last call would add an existing vertex.
		finallyExpectedError error
	}{
		"graph with 3 vertices": {
			vertices: []int{1, 2, 3},
			properties: &VertexProperties{
				Attributes: map[string]string{"color": "red"},
				Weight:     10,
			},
			expectedVertices: []int{1, 2, 3},
			expectedProperties: &VertexProperties{
				Attributes: map[string]string{"color": "red"},
				Weight:     10,
			},
		},
		"graph with duplicated vertex": {
			vertices:             []int{1, 2, 2},
			expectedVertices:     []int{1, 2},
			finallyExpectedError: ErrVertexAlreadyExists,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := newUndirectedTestGraph(nil, nil)

			var err error

			for _, vertex := range test.vertices {
				if test.properties == nil {
					err = g.AddVertex(vertex)
					continue
				}
				// If there are vertex attributes, iterate over them and call the
				// VertexAttribute functional option for each entry. A vertex should
				// only have one attribute so that AddVertex is invoked once.
				for key, value := range test.properties.Attributes {
					err = g.AddVertex(vertex, VertexWeight(test.properties.Weight), VertexAttribute(key, value))
				}
			}

			if !errors.Is(err, test.finallyExpectedError) {
				t.Errorf("%s: error expectancy doesn't match: expected %v, got %v", name, test.finallyExpectedError, err)
			}

			for _, vertex := range test.vertices {
				if len(g.vertices) != len(test.expectedVertices) {
					t.Errorf("%s: vertex count doesn't match: expected %v, got %v", name, len(test.expectedVertices), len(g.vertices))
				}

				hash := g.hash(vertex)
				vertices := g.vertices
				if _, ok := vertices[hash]; !ok {
					t.Errorf("%s: vertex %v not found in graph: %v", name, vertex, vertices)
				}

				if test.properties == nil {
					continue
				}

				if g.vertices[hash].Properties.Weight != test.expectedProperties.Weight {
					t.Errorf("%s: edge weights don't match: expected weight %v, got %v", name, test.expectedProperties.Weight, g.vertices[hash].Properties.Weight)
				}

				if len(g.vertices[hash].Properties.Attributes) != len(test.expectedProperties.Attributes) {
					t.Fatalf("%s: attributes lengths don't match: expcted %v, got %v", name, len(test.expectedProperties.Attributes), len(g.vertices[hash].Properties.Attributes))
				}

				for expectedKey, expectedValue := range test.expectedProperties.Attributes {
					value, ok := g.vertices[hash].Properties.Attributes[expectedKey]
					if !ok {
						t.Errorf("%s: attribute keys don't match: expected key %v not found", name, expectedKey)
					}
					if value != expectedValue {
						t.Errorf("%s: attribute values don't match: expected value %v for key %v, got %v", name, expectedValue, expectedKey, value)
					}
				}
			}
		})
	}
}

func TestUndirected_Vertex(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		vertex        int
		expectedError error
	}{
		"existing vertex": {
			vertices: []int{1, 2, 3},
			vertex:   2,
		},
		"non-existent vertex": {
			vertices:      []int{1, 2, 3},
			vertex:        4,
			expectedError: ErrVertexNotFound,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := newUndirectedTestGraph(test.vertices, nil)

			vertex, err := graph.Vertex(test.vertex)

			if !errors.Is(err, test.expectedError) {
				t.Errorf("%s: error expectancy doesn't match: expected %v, got %v", name, test.expectedError, err)
			}

			if test.expectedError != nil {
				return
			}

			if vertex.Value != test.vertex {
				t.Errorf("%s: vertex expectancy doesn't match: expected %v, got %v", name, test.vertex, vertex)
			}
		})
	}
}

func TestUndirected_AddEdge(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int, any]
		traits        Traits
		expectedEdges []Edge[int, any]
		// Even though some AddVertex calls might work, at least one of them
		// could fail, e.g. if the last call would introduce a cycle.
		finallyExpectedError error
	}{
		"graph with 2 edges": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2, Properties: EdgeProperties[any]{Weight: 10}},
				{Source: 1, Target: 3, Properties: EdgeProperties[any]{Weight: 20}},
			},
			expectedEdges: []Edge[int, any]{
				{Source: 1, Target: 2, Properties: EdgeProperties[any]{Weight: 10}},
				{Source: 1, Target: 3, Properties: EdgeProperties[any]{Weight: 20}},
			},
		},
		"hashes for non-existent vertices": {
			vertices: []int{1, 2},
			edges: []Edge[int, any]{
				{Source: 1, Target: 3, Properties: EdgeProperties[any]{Weight: 20}},
			},
			finallyExpectedError: ErrVertexNotFound,
		},
		"edge introducing a cycle in an acyclic graph": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 3, Target: 1},
			},
			traits: Traits{
				PreventCycles: true,
			},
			finallyExpectedError: ErrEdgeCreatesCycle,
		},
		"edge already exists": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 3, Target: 1},
				{Source: 3, Target: 1},
			},
			finallyExpectedError: ErrEdgeAlreadyExists,
		},
		"edge with attributes": {
			vertices: []int{1, 2},
			edges: []Edge[int, any]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties[any]{
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
			},
			expectedEdges: []Edge[int, any]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties[any]{
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
			},
		},
		"edge with data": {
			vertices: []int{1, 2},
			edges: []Edge[int, any]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties[any]{
						Data: "foo",
					},
				},
			},
			expectedEdges: []Edge[int, any]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties[any]{
						Data: "foo",
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := newUndirectedTestGraph(test.vertices, nil)
			graph.traits = test.traits

			var err error

			for _, edge := range test.edges {
				err = graph.AddEdge(EdgeCopy(edge))
				if err != nil {
					break
				}
			}

			if !errors.Is(err, test.finallyExpectedError) {
				t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v", name, test.finallyExpectedError, err)
			}

			for _, expectedEdge := range test.expectedEdges {
				sourceHash := graph.hash(expectedEdge.Source)
				targetHash := graph.hash(expectedEdge.Target)

				edge, ok := graph.outEdges[sourceHash][targetHash]
				if !ok {
					t.Fatalf("%s: edge with source %v and target %v not found", name, expectedEdge.Source, expectedEdge.Target)
				}

				if !edgesAreEqual(expectedEdge, *edge, false) {
					t.Errorf("%s: expected edge %v, got %v", name, expectedEdge, edge)
				}
			}
		})
	}
}

func TestUndirected_RemoveVertex(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int, any]
		vertex        int
		expectedError error
	}{
		"existing disconnected vertex": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int, any]{
				{Source: 2, Target: 3},
			},
			vertex: 1,
		},
		"existing connected vertex": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
			},
			vertex:        1,
			expectedError: ErrVertexHasEdges,
		},
		"non-existent vertex": {
			vertices:      []int{1, 2, 3},
			edges:         []Edge[int, any]{},
			vertex:        4,
			expectedError: ErrVertexNotFound,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := newUndirectedTestGraph(test.vertices, test.edges)

			err := graph.RemoveVertex(test.vertex)

			if !errors.Is(err, test.expectedError) {
				t.Errorf("%s: error expectancy doesn't match: expected %v, got %v", name, test.expectedError, err)
			}
		})
	}
}

func TestUndirected_Edge(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edge          Edge[int, any]
		args          [2]int
		expectedError error
	}{
		"get edge of undirected graph": {
			vertices: []int{1, 2, 3},
			edge:     Edge[int, any]{Source: 1, Target: 2},
			args:     [2]int{1, 2},
		},
		"get edge of undirected graph with swapped source and target": {
			vertices: []int{1, 2, 3},
			edge:     Edge[int, any]{Source: 1, Target: 2},
			args:     [2]int{2, 1},
		},
		"get non-existent edge of undirected graph": {
			vertices:      []int{1, 2, 3},
			edge:          Edge[int, any]{Source: 1, Target: 2},
			args:          [2]int{2, 3},
			expectedError: ErrEdgeNotFound,
		},
		"get edge with properties": {
			vertices: []int{1, 2, 3},
			edge: Edge[int, any]{
				Source: 1,
				Target: 2,
				Properties: EdgeProperties[any]{
					// Attributes can't be tested at the moment, because there
					// is no way to add multiple attributes at once (using a
					// functional option like EdgeAttributes).
					// ToDo: Add Attributes once EdgeAttributes exists.
					Attributes: map[string]string{},
					Weight:     10,
					Data:       "this is an edge",
				},
			},
			args: [2]int{1, 2},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := newUndirectedTestGraph(test.vertices, []Edge[int, any]{test.edge})

			edge, err := graph.Edge(test.args[0], test.args[1])

			if !errors.Is(err, test.expectedError) {
				t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v", name, test.expectedError, err)
			}

			if test.expectedError != nil {
				return
			}

			if edge.Source != test.args[0] {
				t.Errorf("%s: source expectancy doesn't match: expected %v, got %v", name, test.args[0], edge.Source)
			}

			if edge.Target != test.args[1] {
				t.Errorf("%s: target expectancy doesn't match: expected %v, got %v", name, test.args[1], edge.Target)
			}

			if !edgesAreEqual(test.edge, edge, false) {
				t.Errorf("%s: expected edge %v, got %v", name, test.edge, edge)
			}
		})
	}
}

func TestUndirected_Edges(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int, any]
		expectedEdges []Edge[int, any]
	}{
		"graph with 3 edges": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int, any]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties[any]{
						Weight: 10,
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
				{
					Source: 2,
					Target: 3,
					Properties: EdgeProperties[any]{
						Weight: 20,
						Attributes: map[string]string{
							"color": "green",
						},
					},
				},
				{
					Source: 3,
					Target: 1,
					Properties: EdgeProperties[any]{
						Weight: 30,
						Attributes: map[string]string{
							"color": "blue",
						},
					},
				},
			},
			expectedEdges: []Edge[int, any]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties[any]{
						Weight: 10,
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
				{
					Source: 2,
					Target: 3,
					Properties: EdgeProperties[any]{
						Weight: 20,
						Attributes: map[string]string{
							"color": "green",
						},
					},
				},
				{
					Source: 3,
					Target: 1,
					Properties: EdgeProperties[any]{
						Weight: 30,
						Attributes: map[string]string{
							"color": "blue",
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := newUndirectedTestGraph(test.vertices, test.edges)

			edges := make([]Edge[int, any], 0, len(test.expectedEdges))
			for e, _ := range g.Edges() {
				edges = append(edges, e)
			}

			for _, expectedEdge := range test.expectedEdges {
				for _, actualEdge := range edges {
					if actualEdge.Source != expectedEdge.Source || actualEdge.Target != expectedEdge.Target {
						continue
					}
					if !edgesAreEqual(expectedEdge, actualEdge, false) {
						t.Errorf("%s: expected edge %v, got %v", name, expectedEdge, actualEdge)
					}
				}
			}
		})
	}
}

func TestUndirected_UpdateEdge(t *testing.T) {
	tests := map[string]struct {
		vertices    []int
		edges       []Edge[int, any]
		updateEdge  Edge[int, any]
		expectedErr error
	}{
		"update an edge": {
			vertices: []int{1, 2},
			edges: []Edge[int, any]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties[any]{
						Weight: 10,
						Attributes: map[string]string{
							"color": "red",
						},
						Data: "my-edge",
					},
				},
			},
			updateEdge: Edge[int, any]{
				Source: 1,
				Target: 2,
				Properties: EdgeProperties[any]{
					Weight: 20,
					Attributes: map[string]string{
						"color": "blue",
						"label": "a blue edge",
					},
					Data: "my-updated-edge",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := newUndirectedTestGraph(test.vertices, test.edges)

			err := g.UpdateEdge(EdgeCopy(test.updateEdge))

			if !errors.Is(err, test.expectedErr) {
				t.Fatalf("expected error %v, got %v", test.expectedErr, err)
			}

			actualEdge, err := g.Edge(test.updateEdge.Source, test.updateEdge.Target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err.Error())
			}

			if !edgesAreEqual(test.updateEdge, actualEdge, false) {
				t.Errorf("expected edge %v, got %v", test.updateEdge, actualEdge)
			}
		})
	}
}

func TestUndirected_RemoveEdge(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int, any]
		removeEdges   []Edge[int, any]
		expectedError error
	}{
		"two-vertices graph": {
			vertices: []int{1, 2},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
			},
			removeEdges: []Edge[int, any]{
				{Source: 1, Target: 2},
			},
		},
		"remove 2 edges from triangle": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
			},
			removeEdges: []Edge[int, any]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
			},
		},
		"remove non-existent edge": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
			},
			removeEdges: []Edge[int, any]{
				{Source: 2, Target: 3},
			},
			expectedError: ErrEdgeNotFound,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := newUndirectedTestGraph(test.vertices, test.edges)

			for _, removeEdge := range test.removeEdges {
				if err := graph.RemoveEdge(removeEdge.Source, removeEdge.Target); !errors.Is(err, test.expectedError) {
					t.Errorf("%s: error expectancy doesn't match: expected %v, got %v", name, test.expectedError, err)
				}
				// After removing the edge, verify that it can't be retrieved using
				// Edge anymore.
				if _, err := graph.Edge(removeEdge.Source, removeEdge.Target); !errors.Is(err, ErrEdgeNotFound) {
					t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v", name, ErrEdgeNotFound, err)
				}
			}
		})
	}
}

func TestUndirected_Adjacencies(t *testing.T) {
	tests := map[string]struct {
		vertices []int
		edges    []Edge[int, any]
		expected map[int]map[int]Edge[int, any]
	}{
		"Y-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int, any]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int, any]{
				1: {
					3: {Source: 1, Target: 3},
				},
				2: {
					3: {Source: 2, Target: 3},
				},
				3: {
					1: {Source: 3, Target: 1},
					2: {Source: 3, Target: 2},
					4: {Source: 3, Target: 4},
				},
				4: {
					3: {Source: 4, Target: 3},
				},
			},
		},
		"diamond-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int, any]{
				1: {
					2: {Source: 1, Target: 2},
					3: {Source: 1, Target: 3},
				},
				2: {
					1: {Source: 2, Target: 1},
					4: {Source: 2, Target: 4},
				},
				3: {
					1: {Source: 3, Target: 1},
					4: {Source: 3, Target: 4},
				},
				4: {
					2: {Source: 4, Target: 2},
					3: {Source: 4, Target: 3},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := newUndirectedTestGraph(test.vertices, test.edges)

			adjacencyMap, _ := graph.AdjacencyMap()

			for expectedVertex, expectedAdjacencies := range test.expected {
				adjacencies, ok := adjacencyMap[expectedVertex]
				if !ok {
					t.Errorf("%s: expected vertex %v does not exist in adjacency map", name, expectedVertex)
				}

				for expectedAdjacency, expectedEdge := range expectedAdjacencies {
					edge, ok := adjacencies[expectedAdjacency]
					if !ok {
						t.Errorf("%s: expected adjacency %v does not exist in map of %v", name, expectedAdjacency, expectedVertex)
					}
					if edge.Source != expectedEdge.Source || edge.Target != expectedEdge.Target {
						t.Errorf("%s: edge expectancy doesn't match: expected %v, got %v", name, expectedEdge, edge)
					}
				}
			}
		})
	}
}

func TestUndirected_PredecessorMap(t *testing.T) {
	tests := map[string]struct {
		vertices []int
		edges    []Edge[int, any]
		expected map[int]map[int]Edge[int, any]
	}{
		"Y-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int, any]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int, any]{
				1: {
					3: {Source: 3, Target: 1},
				},
				2: {
					3: {Source: 3, Target: 2},
				},
				3: {
					1: {Source: 1, Target: 3},
					2: {Source: 2, Target: 3},
					4: {Source: 4, Target: 3},
				},
				4: {
					3: {Source: 3, Target: 4},
				},
			},
		},
		"diamond-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int, any]{
				1: {
					2: {Source: 2, Target: 1},
					3: {Source: 3, Target: 1},
				},
				2: {
					1: {Source: 1, Target: 2},
					4: {Source: 4, Target: 2},
				},
				3: {
					1: {Source: 1, Target: 3},
					4: {Source: 4, Target: 3},
				},
				4: {
					2: {Source: 2, Target: 4},
					3: {Source: 3, Target: 4},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			graph := newUndirectedTestGraph(test.vertices, test.edges)

			predecessors, _ := graph.PredecessorMap()

			for expectedVertex, expectedPredecessors := range test.expected {
				actualPredecessors, ok := predecessors[expectedVertex]
				if !ok {
					t.Errorf("%s: expected vertex %v does not exist in adjacency map", name, expectedVertex)
				}

				for expectedPredecessor, expectedEdge := range expectedPredecessors {
					actualEdge, ok := actualPredecessors[expectedPredecessor]
					if !ok {
						t.Errorf("%s: expected adjacency %v does not exist in map of %v", name, expectedPredecessor, expectedVertex)
					}
					if actualEdge.Source != expectedEdge.Source || actualEdge.Target != expectedEdge.Target {
						t.Errorf("%s: edge expectancy doesn't match: expected %v, got %v", name, expectedEdge, actualEdge)
					}
				}
			}
		})
	}
}

func TestUndirected_OrderAndSize(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int, any]
		expectedOrder int
		expectedSize  int
	}{
		"Y-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int, any]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
				{Source: 3, Target: 4},
			},
			expectedOrder: 4,
			expectedSize:  3,
		},
		"diamond-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			expectedOrder: 4,
			expectedSize:  4,
		},
		"two-vertices graph": {
			vertices: []int{1, 2},
			edges: []Edge[int, any]{
				{Source: 1, Target: 2},
			},
			expectedOrder: 2,
			expectedSize:  1,
		},
		"edgeless graph": {
			vertices:      []int{1, 2},
			edges:         []Edge[int, any]{},
			expectedOrder: 2,
			expectedSize:  0,
		},
	}

	for name, test := range tests {
		graph := newUndirectedTestGraph(test.vertices, test.edges)

		order, _ := graph.Order()
		size, _ := graph.Size()

		if order != test.expectedOrder {
			t.Errorf("%s: order expectancy doesn't match: expected %d, got %d", name, test.expectedOrder, order)
		}

		if size != test.expectedSize {
			t.Errorf("%s: size expectancy doesn't match: expected %d, got %d", name, test.expectedSize, size)
		}

	}
}

func newUndirectedTestGraph(vertices []int, edges []Edge[int, any]) *MemoryGraph[int, int, any] {
	g := NewMemoryGraph[int, int, any](IntHash)

	for _, vertex := range vertices {
		_ = g.AddVertex(vertex)
	}

	for _, edge := range edges {
		_ = g.AddEdge(EdgeCopy(edge))
	}
	return g
}
