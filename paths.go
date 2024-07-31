package graph

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
)

var ErrTargetNotReachable = errors.New("target vertex not reachable from source")

type Path[K comparable] []K

func PathWeight[K comparable, V any, E any](g GraphRead[K, V, E], path Path[K]) (weight float64, err error) {
	verticesWeighted, edgesWeighted := g.Traits().IsWeighted()

	for i := 1; i < len(path)-1; i++ {
		var e Edge[K, E]
		e, err = g.Edge(path[i-1], path[i])
		if err != nil {
			err = fmt.Errorf("edge(path[%d], path[%d]): %w", i-1, i, err)
			return
		}
		if edgesWeighted {
			weight += e.Properties.Weight
		} else {
			weight += 1
		}
		if verticesWeighted {
			v, err := g.Vertex(path[i])
			if err != nil {
				return 0, fmt.Errorf("vertex(path[%d]): %w", i, err)
			}
			weight += v.Properties.Weight
		}
	}
	return
}

func (p Path[K]) Contains(k K) bool {
	for _, elem := range p {
		if elem == k {
			return true
		}
	}
	return false
}

func (p Path[K]) String() string {
	parts := make([]string, len(p))
	for i, elem := range p {
		parts[i] = fmt.Sprintf("%v", elem)
	}
	return strings.Join(parts, " -> ")
}

// GraphCycles is used for graphs that can provide more efficient
// implementations of the CreatesCycle method.
type GraphCycles[K comparable] interface {
	CreatesCycle(source, target K) (bool, error)
}

// CreatesCycle determines whether adding an edge between the two given vertices
// would introduce a cycle in the graph. The caller is responsible for ensuring
// that the source and target vertices exist in the graph. If either of the
// vertices does not exist, it will return false.
//
// A potential edge would create a cycle if the target vertex is also a parent
// of the source vertex. In order to determine this, CreatesCycle runs a DFS.
func CreatesCycle[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, source, target K) (bool, error) {
	if source == target {
		return true, nil
	}

	predecessorMap, err := g.PredecessorMap()
	if err != nil {
		return false, fmt.Errorf("failed to get predecessor map: %w", err)
	}

	stack := newStack[K]()
	visited := make(map[K]bool)

	stack.push(source)

	for !stack.isEmpty() {
		currentHash, _ := stack.pop()

		if _, ok := visited[currentHash]; !ok {
			// If the adjacent vertex also is the target vertex, the target is a
			// parent of the source vertex. An edge would introduce a cycle.
			if currentHash == target {
				return true, nil
			}

			visited[currentHash] = true

			for adjacency := range predecessorMap[currentHash] {
				stack.push(adjacency)
			}
		}
	}

	return false, nil
}

type GraphShortestPath[K comparable, V any, E any] interface {
	// ShortestPath returns a function which computes the shortest path between the given source
	// and arbitrary target vertices. A default implementation is provided: [ShortestPath]/[ShortestPathStable]
	// as well as algorithm specific implementations in [DijkstraShortestPath] and [BellmanFordShortestPath].
	// Note: Intermediate data is expected to be computed on the construction of the pathing function, not
	// on Any errors encountered during the computation of the shortest path will be returned when the
	// function is called.
	ShortestPaths(source K, less func(a, b K) bool) ShortestPather[K]
}

// ShortestPather is a function that computes the shortest path to a target vertex.
// If the target is not reachable from the source, ErrTargetNotReachable will be returned.
type ShortestPather[K comparable] func(target K) (Path[K], error)

// ShortestPath computes the shortest path between a source and a target vertex
// under consideration of the edge weights. It returns a slice of hash values of
// the vertices forming that path.
//
// The returned path includes the source and target vertices. Should
// there be multiple shortest paths, and arbitrary one will be returned.
func ShortestPath[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, source K) ShortestPather[K] {
	if g.Traits().IsDirected {
		return BellmanFordShortestPath(g, source, nil)
	}
	return DijkstraShortestPath(g, source)
}

func ShortestPathStable[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, source K, less func(a, b K) bool) ShortestPather[K] {
	return BellmanFordShortestPath(g, source, less)
}

type shortestPathResult[K comparable] struct {
	source K
	prev   map[K]K
}

func (b shortestPathResult[K]) ShortestPath(target K) (Path[K], error) {
	var path Path[K]
	current := target
	for current != b.source {
		next, ok := b.prev[current]
		if !ok {
			return nil, ErrTargetNotReachable
		}
		path = append(path, current)
		current = next
	}
	path = append(path, b.source)
	slices.Reverse(path)
	return path, nil
}

type errShortestPath[K comparable] struct {
	err error
}

func (b errShortestPath[K]) ShortestPath(target K) (Path[K], error) {
	return nil, b.err
}

func DijkstraShortestPath[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, source K) ShortestPather[K] {
	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return errShortestPath[K]{err: fmt.Errorf("could not get adjacency map: %w", err)}.ShortestPath
	}

	weights := make(map[K]float64)
	weights[source] = 0

	queue := newPriorityQueue[K]()
	for hash := range adjacencyMap {
		if hash != source {
			weights[hash] = math.Inf(1)
		}

		queue.Push(hash, weights[hash])
	}

	verticesWeighted, edgesWeighted := g.Traits().IsWeighted()

	// bestPredecessors stores the cheapest or least-weighted predecessor for
	// each vertex. Given an edge AC with weight=4 and an edge BC with weight=2,
	// the cheapest predecessor for C is B.
	bestPredecessors := make(map[K]K)

	for queue.Len() > 0 {
		vertex, _ := queue.Pop()
		hasInfiniteWeight := math.IsInf(weights[vertex], 1)

		for adjacency, edge := range adjacencyMap[vertex] {
			weight := edge.Properties.Weight

			// Setting the weight to 1 is required for unweighted graphs whose
			// edge weights are 0. Otherwise, all paths would have a sum of 0
			// and a random path would be returned.
			if !edgesWeighted {
				weight = 1
			}

			weight += weights[vertex]

			if verticesWeighted {
				v, err := g.Vertex(adjacency)
				if err != nil {
					return errShortestPath[K]{err: fmt.Errorf("could not get vertex to determine weight: %w", err)}.ShortestPath
				}
				weight += v.Properties.Weight
			}

			if weight < weights[adjacency] && !hasInfiniteWeight {
				weights[adjacency] = weight
				bestPredecessors[adjacency] = vertex
				queue.UpdatePriority(adjacency, weight)
			}
		}
	}

	return shortestPathResult[K]{
		source: source,
		prev:   bestPredecessors,
	}.ShortestPath
}

// BellmanFordShortestPath is a helper function for ShortestPath that uses the Bellman-Ford algorithm to
// compute the shortest path between a source and a target vertex using the edge weights and returns
// the hash values of the vertices forming that path. This search runs in O(|V|*|E|) time.
//
// The returned path includes the source and target vertices. If the target cannot be reached
// from the source vertex, ErrTargetNotReachable will be returned. If there are multiple shortest
func BellmanFordShortestPath[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, source K, less func(a, b K) bool) ShortestPather[K] {
	dist := make(map[K]float64)
	bestPredecessors := make(map[K]K)

	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return errShortestPath[K]{err: fmt.Errorf("could not get adjacency map: %w", err)}.ShortestPath
	}
	keys := make([]K, 0, len(adjacencyMap))
	for key := range adjacencyMap {
		dist[key] = math.MaxInt32
		keys = append(keys, key)
	}
	dist[source] = 0
	if less != nil {
		sort.Slice(keys, func(i, j int) bool {
			return less(keys[i], keys[j])
		})
	}

	verticesWeighted, edgesWeighted := g.Traits().IsWeighted()

	for i := 0; i < len(adjacencyMap)-1; i++ {
		for _, key := range keys {
			edges := adjacencyMap[key]
			for adj, edge := range edges {
				weight := edge.Properties.Weight
				if !edgesWeighted {
					weight = 1
				}
				if verticesWeighted {
					v, err := g.Vertex(adj)
					if err != nil {
						return errShortestPath[K]{err: fmt.Errorf("could not get vertex to determine weight: %w", err)}.ShortestPath
					}
					weight += v.Properties.Weight
				}
				if newDist := dist[key] + weight; newDist < dist[edge.Target] {
					dist[edge.Target] = newDist
					bestPredecessors[edge.Target] = key
				}
			}
		}
	}

	for _, edges := range adjacencyMap {
		for adj, edge := range edges {
			weight := edge.Properties.Weight
			if !edgesWeighted {
				weight = 1
			}
			if verticesWeighted {
				v, err := g.Vertex(adj)
				if err != nil {
					return errShortestPath[K]{err: fmt.Errorf("could not get vertex to determine weight: %w", err)}.ShortestPath
				}
				weight += v.Properties.Weight
			}
			if newDist := dist[edge.Source] + weight; newDist < dist[edge.Target] {
				return errShortestPath[K]{err: errors.New("graph contains a negative-weight cycle")}.ShortestPath
			}
		}
	}

	return shortestPathResult[K]{
		source: source,
		prev:   bestPredecessors,
	}.ShortestPath
}

type sccState[K comparable, E any] struct {
	adjacencyMap map[K]map[K]Edge[K, E]
	components   [][]K
	stack        *stack[K]
	visited      map[K]struct{}
	lowlink      map[K]int
	index        map[K]int
	time         int
}

// StronglyConnectedComponents detects all strongly connected components within
// the graph and returns the hashes of the vertices shaping these components, so
// each component is represented by a []K.
//
// StronglyConnectedComponents can only run on directed graphs.
func StronglyConnectedComponents[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}) ([][]K, error) {
	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return nil, fmt.Errorf("could not get adjacency map: %w", err)
	}

	state := &sccState[K, E]{
		adjacencyMap: adjacencyMap,
		components:   make([][]K, 0),
		stack:        newStack[K](),
		visited:      make(map[K]struct{}),
		lowlink:      make(map[K]int),
		index:        make(map[K]int),
	}

	for hash := range state.adjacencyMap {
		if _, ok := state.visited[hash]; !ok {
			findSCC(hash, state)
		}
	}

	return state.components, nil
}

func findSCC[K comparable, E any](vertexHash K, state *sccState[K, E]) {
	state.stack.push(vertexHash)
	state.visited[vertexHash] = struct{}{}
	state.index[vertexHash] = state.time
	state.lowlink[vertexHash] = state.time

	state.time++

	for adjacency := range state.adjacencyMap[vertexHash] {
		if _, ok := state.visited[adjacency]; !ok {
			findSCC(adjacency, state)

			smallestLowlink := math.Min(
				float64(state.lowlink[vertexHash]),
				float64(state.lowlink[adjacency]),
			)
			state.lowlink[vertexHash] = int(smallestLowlink)
		} else {
			// If the adjacent vertex already is on the stack, the edge joining
			// the current and the adjacent vertex is a back ege. Therefore, the
			// lowlink value of the vertex has to be updated to the index of the
			// adjacent vertex if it is smaller than the current lowlink value.
			if state.stack.contains(adjacency) {
				smallestLowlink := math.Min(
					float64(state.lowlink[vertexHash]),
					float64(state.index[adjacency]),
				)
				state.lowlink[vertexHash] = int(smallestLowlink)
			}
		}
	}

	// If the lowlink value of the vertex is equal to its DFS value, this is the
	// head vertex of a strongly connected component that's shaped by the vertex
	// and all vertices on the stack.
	if state.lowlink[vertexHash] == state.index[vertexHash] {
		var hash K
		var component []K

		for hash != vertexHash {
			hash, _ = state.stack.pop()

			component = append(component, hash)
		}

		state.components = append(state.components, component)
	}
}

type (
	PathIter[K comparable] func(yield func(Path[K], error) bool)

	GraphAllPaths[K comparable] interface {
		// AllPathsBetween returns all the paths between two given vertices. A default implementation is provided: [AllPathsBetween].
		AllPathsBetween(start, end K) PathIter[K]
	}
)

// AllPathsBetween computes and returns all paths between two given vertices. A
// path is represented as a slice of vertex hashes. The returned slice contains
// these paths.
//
// AllPathsBetween utilizes a non-recursive, stack-based implementation. It has
// an estimated runtime complexity of O(n^2) where n is the number of vertices.
func AllPathsBetween[K comparable, E any](g GraphRelations[K, E], start, end K) PathIter[K] {
	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return func(yield func(Path[K], error) bool) {
			yield(nil, fmt.Errorf("could not get adjacency map: %w", err))
		}
	}
	return AllPathsFromAdjacency(adjacencyMap, start, end)
}

// [ET] is either `Edge[K, E]` (when used with [AllPathsBetween]) or `*Edge[K, E]` (used for eg. memoryGraph implementation)
func AllPathsFromAdjacency[K comparable, ET any](adjacencyMap map[K]map[K]ET, start, end K) PathIter[K] {
	// Use a pool to save on allocations
	var oldStacks []*stack[K]
	newStack := func() *stack[K] {
		if len(oldStacks) == 0 {
			return newStack[K]()
		}
		s := oldStacks[len(oldStacks)-1]
		oldStacks = oldStacks[:len(oldStacks)-1]
		s.clear()
		return s
	}

	// The algorithm used relies on stacks instead of recursion. It is described
	// here: https://boycgit.github.io/all-paths-between-two-vertex/
	mainStack := newStack()
	viceStack := newStackOfStacks[K]()

	checkEmpty := func() error {
		if mainStack.isEmpty() || viceStack.isEmpty() {
			return errors.New("empty stack")
		}
		return nil
	}

	buildLayer := func(element K) {
		mainStack.push(element)
		newElements := newStack()

		for e := range adjacencyMap[element] {
			var contains bool
			var containsCount int
			mainStack.forEach(func(k K) {
				if e == k {
					contains = true
					containsCount++
				}
			})
			if contains && (e != start || e != end) || containsCount > 1 {
				continue
			}
			newElements.push(e)
		}
		viceStack.push(newElements)
	}

	buildStack := func() error {
		if err := checkEmpty(); err != nil {
			return fmt.Errorf("unable to build stack: %w", err)
		}

		elements, _ := viceStack.top()

		for !elements.isEmpty() {
			element, _ := elements.pop()
			buildLayer(element)
			elements, _ = viceStack.top()
		}

		return nil
	}

	removeLayer := func() error {
		if err := checkEmpty(); err != nil {
			return fmt.Errorf("unable to remove layer: %w", err)
		}

		if e, _ := viceStack.top(); !e.isEmpty() {
			return errors.New("the top element of vice-stack is not empty")
		}

		_, _ = mainStack.pop()
		s, _ := viceStack.pop()
		oldStacks = append(oldStacks, s)

		return nil
	}

	buildLayer(start)

	return func(yield func(Path[K], error) bool) {
		for !mainStack.isEmpty() {
			v, _ := mainStack.top()
			adjs, _ := viceStack.top()

			if adjs.isEmpty() {
				if v == end && len(mainStack.elements) > 1 {
					path := make([]K, 0)
					mainStack.forEach(func(k K) {
						path = append(path, k)
					})
					if !yield(path, nil) {
						return
					}
				}

				err := removeLayer()
				if err != nil {
					if !yield(nil, err) {
						return
					}
				}
			} else {
				if err := buildStack(); err != nil {
					if !yield(nil, err) {
						return
					}
				}
			}
		}
	}
}
