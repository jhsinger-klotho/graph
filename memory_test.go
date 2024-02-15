package graph

import (
	"errors"
	"testing"
)

func TestDirected_AddVertex(t *testing.T) {
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
		graph := NewMemoryGraph(IntHash)

		var err error

		for _, vertex := range test.vertices {
			if test.properties == nil {
				err = graph.AddVertex(vertex)
				continue
			}
			// If there are vertex attributes, iterate over them and call the
			// VertexAttribute functional option for each entry. A vertex should
			// only have one attribute so that AddVertex is invoked once.
			for key, value := range test.properties.Attributes {
				err = graph.AddVertex(vertex, VertexWeight(test.properties.Weight), VertexAttribute(key, value))
			}
		}

		if !errors.Is(err, test.finallyExpectedError) {
			t.Errorf("%s: error expectancy doesn't match: expected %v, got %v", name, test.finallyExpectedError, err)
		}

		for _, vertex := range test.vertices {
			if len(graph.vertices) != len(test.expectedVertices) {
				t.Errorf("%s: vertex count doesn't match: expected %v, got %v", name, len(test.expectedVertices), len(graph.vertices))
			}

			hash := graph.hash(vertex)
			vertex, err := graph.Vertex(hash)
			if err != nil {
				vertices := graph.vertices
				t.Errorf("%s: vertex %v not found in graph: %v", name, vertex, vertices)
			}

			if test.properties == nil {
				continue
			}

			if vertex.Properties.Weight != test.expectedProperties.Weight {
				t.Errorf("%s: edge weights don't match: expected weight %v, got %v", name, test.expectedProperties.Weight, vertex.Properties.Weight)
			}

			if len(vertex.Properties.Attributes) != len(test.expectedProperties.Attributes) {
				t.Fatalf("%s: attributes lengths don't match: expcted %v, got %v", name, len(test.expectedProperties.Attributes), len(vertex.Properties.Attributes))
			}

			for expectedKey, expectedValue := range test.expectedProperties.Attributes {
				value, ok := vertex.Properties.Attributes[expectedKey]
				if !ok {
					t.Errorf("%s: attribute keys don't match: expected key %v not found", name, expectedKey)
				}
				if value != expectedValue {
					t.Errorf("%s: attribute values don't match: expected value %v for key %v, got %v", name, expectedValue, expectedKey, value)
				}
			}
		}
	}
}

func TestDirected_Vertex(t *testing.T) {
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
		graph := NewMemoryGraph(IntHash)

		for _, vertex := range test.vertices {
			_ = graph.AddVertex(vertex)
		}

		vertex, err := graph.Vertex(test.vertex)

		if !errors.Is(err, test.expectedError) {
			t.Errorf("%s: error expectancy doesn't match: expected %v, got %v", name, test.expectedError, err)
		}

		if test.expectedError != nil {
			continue
		}

		if vertex.Value != test.vertex {
			t.Errorf("%s: vertex expectancy doesn't match: expected %v, got %v", name, test.vertex, vertex)
		}
	}
}

func TestDirected_RemoveVertex(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int]
		vertex        int
		expectedError error
	}{
		"existing disconnected vertex": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 2, Target: 3},
			},
			vertex: 1,
		},
		"existing connected vertex": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
			},
			vertex:        1,
			expectedError: ErrVertexHasEdges,
		},
		"non-existent vertex": {
			vertices:      []int{1, 2, 3},
			edges:         []Edge[int]{},
			vertex:        4,
			expectedError: ErrVertexNotFound,
		},
	}

	for name, test := range tests {
		graph := NewMemoryGraph(IntHash)

		for _, vertex := range test.vertices {
			_ = graph.AddVertex(vertex)
		}

		for _, edge := range test.edges {
			_ = graph.AddEdge(edge.Source, edge.Target)
		}

		err := graph.RemoveVertex(test.vertex)

		if !errors.Is(err, test.expectedError) {
			t.Errorf("%s: error expectancy doesn't match: expected %v, got %v", name, test.expectedError, err)
		}
	}
}

func TestDirected_AddEdge(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int]
		expectedEdges []Edge[int]
		// Even though some AddVertex calls might work, at least one of them
		// could fail, e.g. if the last call would introduce a cycle.
		finallyExpectedError error
	}{
		"graph with 2 edges": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2, Properties: EdgeProperties{Weight: 10}},
				{Source: 1, Target: 3, Properties: EdgeProperties{Weight: 20}},
			},
			expectedEdges: []Edge[int]{
				{Source: 1, Target: 2, Properties: EdgeProperties{Weight: 10}},
				{Source: 1, Target: 3, Properties: EdgeProperties{Weight: 20}},
			},
		},
		"hashes for non-existent vertices": {
			vertices: []int{1, 2},
			edges: []Edge[int]{
				{Source: 1, Target: 3, Properties: EdgeProperties{Weight: 20}},
			},
			finallyExpectedError: ErrVertexNotFound,
		},
		"edge already exists": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 3},
				{Source: 3, Target: 1},
				{Source: 3, Target: 1},
			},
			finallyExpectedError: ErrEdgeAlreadyExists,
		},
		"edge with attributes": {
			vertices: []int{1, 2},
			edges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
			},
			expectedEdges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
			},
		},
		"edge with data": {
			vertices: []int{1, 2},
			edges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Data: "foo",
					},
				},
			},
			expectedEdges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Data: "foo",
					},
				},
			},
		},
	}

	for name, test := range tests {
		graph := NewMemoryGraph(IntHash)

		for _, vertex := range test.vertices {
			_ = graph.AddVertex(vertex)
		}

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

			edge, err := graph.Edge(sourceHash, targetHash)
			if err != nil {
				t.Fatalf("%s: edge with source %v and target %v not found", name, expectedEdge.Source, expectedEdge.Target)
			}

			if !edgesAreEqual(expectedEdge, edge, true) {
				t.Errorf("%s: expected edge %v, got %v", name, expectedEdge, edge)
			}
		}
	}
}

func TestDirected_Edge(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edge          Edge[int]
		args          [2]int
		expectedError error
	}{
		"get edge of undirected graph": {
			vertices: []int{1, 2, 3},
			edge:     Edge[int]{Source: 1, Target: 2},
			args:     [2]int{1, 2},
		},
		"get edge of undirected graph with swapped source and target": {
			vertices:      []int{1, 2, 3},
			edge:          Edge[int]{Source: 1, Target: 2},
			args:          [2]int{2, 1},
			expectedError: ErrEdgeNotFound,
		},
		"get non-existent edge of undirected graph": {
			vertices:      []int{1, 2, 3},
			edge:          Edge[int]{Source: 1, Target: 2},
			args:          [2]int{2, 3},
			expectedError: ErrEdgeNotFound,
		},
		"get edge with properties": {
			vertices: []int{1, 2, 3},
			edge: Edge[int]{
				Source: 1,
				Target: 2,
				Properties: EdgeProperties{
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
		graph := newTestGraph(test.vertices, []Edge[int]{test.edge})

		edge, err := graph.Edge(test.args[0], test.args[1])

		if !errors.Is(err, test.expectedError) {
			t.Fatalf("%s: error expectancy doesn't match: expected %v, got %v", name, test.expectedError, err)
		}

		if test.expectedError != nil {
			continue
		}

		if edge.Source != test.args[0] {
			t.Errorf("%s: source expectancy doesn't match: expected %v, got %v", name, test.args[0], edge.Source)
		}

		if edge.Target != test.args[1] {
			t.Errorf("%s: target expectancy doesn't match: expected %v, got %v", name, test.args[1], edge.Target)
		}

		if !edgesAreEqual(test.edge, edge, true) {
			t.Errorf("%s: expected edge %v, got %v", name, test.edge, edge)
		}
	}
}

func TestDirected_Edges(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int]
		expectedEdges []Edge[int]
	}{
		"graph with 3 edges": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Weight: 10,
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
				{
					Source: 2,
					Target: 3,
					Properties: EdgeProperties{
						Weight: 20,
						Attributes: map[string]string{
							"color": "green",
						},
					},
				},
				{
					Source: 3,
					Target: 1,
					Properties: EdgeProperties{
						Weight: 30,
						Attributes: map[string]string{
							"color": "blue",
						},
					},
				},
			},
			expectedEdges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Weight: 10,
						Attributes: map[string]string{
							"color": "red",
						},
					},
				},
				{
					Source: 2,
					Target: 3,
					Properties: EdgeProperties{
						Weight: 20,
						Attributes: map[string]string{
							"color": "green",
						},
					},
				},
				{
					Source: 3,
					Target: 1,
					Properties: EdgeProperties{
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
			g := newTestGraph(test.vertices, test.edges)

			edges := make(map[EdgeKey[int]]Edge[int])
			for e, _ := range g.Edges() {
				edges[EdgeKey[int]{Source: e.Source, Target: e.Target}] = e
			}

			for _, expectedEdge := range test.expectedEdges {
				actualEdge := edges[EdgeKey[int]{Source: expectedEdge.Source, Target: expectedEdge.Target}]
				if !edgesAreEqual(expectedEdge, actualEdge, true) {
					t.Errorf("%s: expected edge %v, got %v", name, expectedEdge, actualEdge)
				}
			}
		})
	}
}

func TestDirected_UpdateEdge(t *testing.T) {
	tests := map[string]struct {
		vertices    []int
		edges       []Edge[int]
		updateEdge  Edge[int]
		expectedErr error
	}{
		"update an edge": {
			vertices: []int{1, 2},
			edges: []Edge[int]{
				{
					Source: 1,
					Target: 2,
					Properties: EdgeProperties{
						Weight: 10,
						Attributes: map[string]string{
							"color": "red",
						},
						Data: "my-edge",
					},
				},
			},
			updateEdge: Edge[int]{
				Source: 1,
				Target: 2,
				Properties: EdgeProperties{
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
			g := newTestGraph(test.vertices, test.edges)

			err := g.UpdateEdge(test.updateEdge.Source, test.updateEdge.Target, EdgeCopyProperties(test.updateEdge.Properties))

			if !errors.Is(err, test.expectedErr) {
				t.Fatalf("expected error %v, got %v", test.expectedErr, err)
			}

			actualEdge, err := g.Edge(test.updateEdge.Source, test.updateEdge.Target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err.Error())
			}

			if !edgesAreEqual(test.updateEdge, actualEdge, true) {
				t.Errorf("expected edge %v, got %v", test.updateEdge, actualEdge)
			}
		})
	}
}

func TestDirected_RemoveEdge(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int]
		removeEdges   []Edge[int]
		expectedError error
	}{
		"two-vertices graph": {
			vertices: []int{1, 2},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
			},
			removeEdges: []Edge[int]{
				{Source: 1, Target: 2},
			},
		},
		"remove 2 edges from triangle": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
			},
			removeEdges: []Edge[int]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
			},
		},
		"remove non-existent edge": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
			},
			removeEdges: []Edge[int]{
				{Source: 2, Target: 3},
			},
			expectedError: ErrEdgeNotFound,
		},
	}

	for name, test := range tests {
		graph := newTestGraph(test.vertices, test.edges)

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
	}
}

func TestDirected_AdjacencyList(t *testing.T) {
	tests := map[string]struct {
		vertices []int
		edges    []Edge[int]
		expected map[int]map[int]Edge[int]
	}{
		"Y-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int]{
				1: {
					3: {Source: 1, Target: 3},
				},
				2: {
					3: {Source: 2, Target: 3},
				},
				3: {
					4: {Source: 3, Target: 4},
				},
				4: {},
			},
		},
		"diamond-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int]{
				1: {
					2: {Source: 1, Target: 2},
					3: {Source: 1, Target: 3},
				},
				2: {
					4: {Source: 2, Target: 4},
				},
				3: {
					4: {Source: 3, Target: 4},
				},
				4: {},
			},
		},
	}

	for name, test := range tests {
		graph := newTestGraph(test.vertices, test.edges)

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
	}
}

func TestDirected_PredecessorMap(t *testing.T) {
	tests := map[string]struct {
		vertices []int
		edges    []Edge[int]
		expected map[int]map[int]Edge[int]
	}{
		"Y-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int]{
				1: {},
				2: {},
				3: {
					1: {Source: 1, Target: 3},
					2: {Source: 2, Target: 3},
				},
				4: {
					3: {Source: 3, Target: 4},
				},
			},
		},
		"diamond-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			expected: map[int]map[int]Edge[int]{
				1: {},
				2: {
					1: {Source: 1, Target: 2},
				},
				3: {
					1: {Source: 1, Target: 3},
				},
				4: {
					2: {Source: 2, Target: 4},
					3: {Source: 3, Target: 4},
				},
			},
		},
	}

	for name, test := range tests {
		graph := newTestGraph(test.vertices, test.edges)

		predecessors, _ := graph.PredecessorMap()

		for expectedVertex, expectedPredecessors := range test.expected {
			predecessors, ok := predecessors[expectedVertex]
			if !ok {
				t.Errorf("%s: expected vertex %v does not exist in adjacency map", name, expectedVertex)
			}

			for expectedPredecessor, expectedEdge := range expectedPredecessors {
				edge, ok := predecessors[expectedPredecessor]
				if !ok {
					t.Errorf("%s: expected adjacency %v does not exist in map of %v", name, expectedPredecessor, expectedVertex)
				}
				if edge.Source != expectedEdge.Source || edge.Target != expectedEdge.Target {
					t.Errorf("%s: edge expectancy doesn't match: expected %v, got %v", name, expectedEdge, edge)
				}
			}
		}
	}
}

func TestDirected_OrderAndSize(t *testing.T) {
	tests := map[string]struct {
		vertices      []int
		edges         []Edge[int]
		expectedOrder int
		expectedSize  int
	}{
		"Y-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 3},
				{Source: 2, Target: 3},
				{Source: 3, Target: 4},
			},
			expectedOrder: 4,
			expectedSize:  3,
		},
		"diamond-shaped graph": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			expectedOrder: 4,
			expectedSize:  4,
		},
		"two-vertices two-edges graph": {
			vertices: []int{1, 2},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 2, Target: 1},
			},
			expectedOrder: 2,
			expectedSize:  2,
		},
		"edgeless graph": {
			vertices:      []int{1, 2},
			edges:         []Edge[int]{},
			expectedOrder: 2,
			expectedSize:  0,
		},
	}

	for name, test := range tests {
		graph := newTestGraph(test.vertices, test.edges)

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

func TestDirected_predecessors(t *testing.T) {
	tests := map[string]struct {
		vertices             []int
		edges                []Edge[int]
		vertex               int
		expectedPredecessors []int
	}{
		"graph with 3 vertices": {
			vertices: []int{1, 2, 3},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
			},
			vertex:               2,
			expectedPredecessors: []int{1},
		},
		"graph with 6 vertices": {
			vertices: []int{1, 2, 3, 4, 5, 6},
			edges: []Edge[int]{
				{Source: 1, Target: 2},
				{Source: 1, Target: 3},
				{Source: 2, Target: 4},
				{Source: 2, Target: 5},
				{Source: 3, Target: 6},
			},
			vertex:               5,
			expectedPredecessors: []int{2},
		},
		"graph with 4 vertices and 3 predecessors": {
			vertices: []int{1, 2, 3, 4},
			edges: []Edge[int]{
				{Source: 1, Target: 4},
				{Source: 2, Target: 4},
				{Source: 3, Target: 4},
			},
			vertex:               4,
			expectedPredecessors: []int{1, 2, 3},
		},
	}

	for name, test := range tests {
		graph := newTestGraph(test.vertices, test.edges)

		predecessors, _ := predecessors[int](graph, graph.hash(test.vertex))

		if !slicesAreEqual(predecessors, test.expectedPredecessors) {
			t.Errorf("%s: predecessors don't match: expected %v, got %v", name, test.expectedPredecessors, predecessors)
		}
	}
}

func slicesAreEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for _, aValue := range a {
		found := false
		for _, bValue := range b {
			if aValue == bValue {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func vertexPropertiesAreEqual(a, b VertexProperties) bool {
	if a.Weight != b.Weight {
		return false
	}

	// A length check is required because in the iteration below, a.Attributes
	// could be empty and thus circumvent the comparison.
	if len(a.Attributes) != len(b.Attributes) {
		return false
	}

	for key, aValue := range a.Attributes {
		bValue, ok := b.Attributes[key]
		if !ok || aValue != bValue {
			return false
		}
	}

	return true
}

func edgesAreEqual[K comparable](a, b Edge[K], directed bool) bool {
	directedOk := a.Source == b.Source &&
		a.Target == b.Target

	undirectedOk := directedOk ||
		a.Source == b.Target &&
			a.Target == b.Source

	if directed && !directedOk {
		return false
	}

	if !directed && !undirectedOk {
		return false
	}

	if a.Properties.Weight != b.Properties.Weight {
		return false
	}

	if len(a.Properties.Attributes) != len(b.Properties.Attributes) {
		return false
	}

	for aKey, aValue := range a.Properties.Attributes {
		bValue, ok := b.Properties.Attributes[aKey]
		if !ok {
			return false
		}
		if bValue != aValue {
			return false
		}
	}

	for bKey := range b.Properties.Attributes {
		if _, ok := a.Properties.Attributes[bKey]; !ok {
			return false
		}
	}

	return true
}

func predecessors[K comparable](g GraphRelations[K], vertexHash K) ([]K, error) {
	var predecessorHashes []K

	predecessorMap, err := g.PredecessorMap()
	if err != nil {
		return nil, err
	}

	for _, edge := range predecessorMap[vertexHash] {
		predecessorHashes = append(predecessorHashes, edge.Source)
	}

	return predecessorHashes, nil
}

func newTestGraph(vertices []int, edges []Edge[int]) *memoryGraph[int, int] {
	g := NewMemoryGraph(IntHash)

	for _, vertex := range vertices {
		_ = g.AddVertex(vertex)
	}

	for _, edge := range edges {
		_ = g.AddEdge(edge.Source, edge.Target, EdgeCopyProperties(edge.Properties))
	}
	return g
}
