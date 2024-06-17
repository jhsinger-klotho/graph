package graph

import (
	"fmt"
	"sort"
)

// TopologicalSort runs a topological sort on a given directed graph and returns
// the vertex hashes in topological order. The topological order is a non-unique
// order of vertices in a directed graph where an edge from vertex A to vertex B
// implies that vertex A appears before vertex B.
//
// Note that TopologicalSort doesn't make any guarantees about the order. If there
// are multiple valid topological orderings, an arbitrary one will be returned.
// To make the output deterministic, use [StableTopologicalSort].
//
// TopologicalSort only works for directed acyclic graphs. This implementation
// works non-recursively and utilizes Kahn's algorithm.
func TopologicalSort[K comparable, E any](predecessorMap map[K]map[K]Edge[K, E]) func(yield func(K) bool) {
	return StableTopologicalSort(predecessorMap, nil)
}

// StableTopologicalSort does the same as [TopologicalSort], but takes a function
// for comparing (and then ordering) two given vertices. This allows for a stable
// and deterministic output even for graphs with multiple topological orderings.
// If the less function is nil, the order will be non-deterministic as in [TopologicalSort].
// Use [PredecessorMap] to get normal topological order; use [AdjacencyMap] to get
// reverse topological order, but keep 'less' as normal.
// This will pick an arbitrary vertex when a cycle is encountered.
// Note, this function is destructive to the map.
func StableTopologicalSort[K comparable, E any](predecessorMap map[K]map[K]Edge[K, E], less func(K, K) bool) func(yield func(K) bool) {
	queue := make([]K, 0, len(predecessorMap))
	seen := make(map[K]struct{}, len(predecessorMap))
	frontier := make([]K, 0, len(predecessorMap))

	for vertex, predecessors := range predecessorMap {
		if len(predecessors) == 0 {
			queue = append(queue, vertex)
			seen[vertex] = struct{}{}
			delete(predecessorMap, vertex)
		}
	}

	isInverted := false
invertedCheck:
	for _, ps := range predecessorMap {
		for t, e := range ps {
			isInverted = e.Target == t
			break invertedCheck
		}
	}

	if less != nil {
		if isInverted {
			oldLess := less
			less = func(a, b K) bool {
				return oldLess(b, a)
			}
		}
		sort.Slice(queue, func(i, j int) bool {
			return less(queue[i], queue[j])
		})
	}

	return func(yield func(K) bool) {
		var currentVertex K
		for {
			if len(queue) == 0 {
				if len(predecessorMap) == 0 {
					return
				}
				remaining := make([]K, 0, len(predecessorMap))
				for vertex := range predecessorMap {
					remaining = append(remaining, vertex)
				}
				sort.Slice(remaining, func(i, j int) bool {
					// Pick an arbitrary vertex to start the queue based first on the number of remaining predecessors
					iPcount := len(predecessorMap[remaining[i]])
					jPcount := len(predecessorMap[remaining[j]])
					if iPcount != jPcount {
						if isInverted {
							return iPcount > jPcount
						} else {
							return iPcount < jPcount
						}
					}

					if less != nil {
						return less(remaining[i], remaining[j])
					}
					return i < j
				})
				currentVertex = remaining[0]
				seen[currentVertex] = struct{}{}
				delete(predecessorMap, currentVertex)
			} else {
				currentVertex, queue = queue[0], queue[1:]
			}

			if !yield(currentVertex) {
				return
			}

			frontier = frontier[:0]

			for vertex, predecessors := range predecessorMap {
				delete(predecessors, currentVertex)

				if len(predecessors) != 0 {
					continue
				}

				if _, ok := seen[vertex]; ok {
					continue
				}

				frontier = append(frontier, vertex)
				seen[vertex] = struct{}{}
				// No more predecessors, so we can remove the vertex from the map.
				// Used for bookkeeping to check for leftover predecessors, indicating
				// a cycle in the graph.
				delete(predecessorMap, vertex)
			}

			if less != nil {
				sort.Slice(frontier, func(i, j int) bool {
					return less(frontier[i], frontier[j])
				})
			}

			queue = append(queue, frontier...)
		}
	}
}

// TransitiveReduction modifies the graph to have the same vertices and the same
// reachability, but with as few edges as possible. The graph
// must be a directed acyclic graph.
//
// TransitiveReduction is a very expensive operation scaling with O(V(V+E)).
func TransitiveReduction[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphWrite[K, V, E]
	GraphRelations[K, E]
}) error {
	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return fmt.Errorf("failed to get adajcency map: %w", err)
	}

	// For each vertex in the graph, run a depth-first search from each direct
	// successor of that vertex. Then, for each vertex visited within the DFS,
	// inspect all of its edges. Remove the edges that also appear in the edge
	// set of the top-level vertex and target the current vertex. These edges
	// are redundant because their targets apparently are not only reachable
	// from the top-level vertex, but also through a DFS.
	for vertex, successors := range adjacencyMap {
		tOrder, err := g.Order()
		if err != nil {
			return fmt.Errorf("failed to get graph order: %w", err)
		}

		for successor := range successors {
			stack := newStack[K]()
			visited := make(map[K]struct{}, tOrder)

			stack.push(successor)

			for !stack.isEmpty() {
				current, _ := stack.pop()

				if _, ok := visited[current]; ok {
					continue
				}

				visited[current] = struct{}{}
				stack.push(current)

				for adjacency := range adjacencyMap[current] {
					if _, ok := visited[adjacency]; ok {
						if stack.contains(adjacency) {
							// If the current adjacency is both on the stack and
							// has already been visited, there is a cycle.
							return fmt.Errorf("transitive reduction cannot be performed on graph with cycle")
						}
						continue
					}

					if _, ok := adjacencyMap[vertex][adjacency]; ok {
						_ = g.RemoveEdge(vertex, adjacency)
					}
					stack.push(adjacency)
				}
			}
		}
	}

	return nil
}
