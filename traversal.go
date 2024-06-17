package graph

import (
	"fmt"
)

// DFS performs a depth-first search on the graph, starting from the given vertex. The visit
// function will be invoked with the hash of the vertex currently visited. If it returns false, DFS
// will continue traversing the graph, and if it returns true, the traversal will be stopped. In
// case the graph is disconnected, only the vertices joined with the starting vertex are visited.
//
// This example prints all vertices of the graph in DFS-order:
//
//	g := graph.New(graph.IntHash)
//
//	_ = g.AddVertex(1)
//	_ = g.AddVertex(2)
//	_ = g.AddVertex(3)
//
//	_ = g.AddEdge(1, 2)
//	_ = g.AddEdge(2, 3)
//	_ = g.AddEdge(3, 1)
//
//	_ = graph.DFS(g, 1, func(value int) bool {
//		fmt.Println(value)
//		return false
//	})
//
// Similarly, if you have a graph of City vertices and the traversal should stop at London, the
// visit function would look as follows:
//
//	func(c City) bool {
//		return c.Name == "London"
//	}
//
// DFS is non-recursive and maintains a stack instead.
func DFS[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, start K) func(yield func(K, error) bool) {

	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return func(yield func(K, error) bool) {
			var zeroKey K
			yield(zeroKey, fmt.Errorf("could not get adjacency map: %w", err))
		}
	}

	if _, ok := adjacencyMap[start]; !ok {
		return func(yield func(K, error) bool) {
			var zeroKey K
			yield(zeroKey, fmt.Errorf("could not find start vertex with hash %v", start))
		}
	}

	stack := newStack[K]()
	visited := make(map[K]bool)

	stack.push(start)

	return func(yield func(K, error) bool) {
		for !stack.isEmpty() {
			currentHash, _ := stack.pop()

			if _, ok := visited[currentHash]; !ok {
				// Stop traversing the graph if the visit function returns true.
				if !yield(currentHash, nil) {
					break
				}
				visited[currentHash] = true

				for adjacency := range adjacencyMap[currentHash] {
					stack.push(adjacency)
				}
			}
		}
	}
}

// BFS performs a breadth-first search on the graph, starting from the given vertex. The visit
// function will be invoked with the hash of the vertex currently visited. If it returns false, BFS
// will continue traversing the graph, and if it returns true, the traversal will be stopped. In
// case the graph is disconnected, only the vertices joined with the starting vertex are visited.
//
// This example prints all vertices of the graph in BFS-order:
//
//	g := graph.New(graph.IntHash)
//
//	_ = g.AddVertex(1)
//	_ = g.AddVertex(2)
//	_ = g.AddVertex(3)
//
//	_ = g.AddEdge(1, 2)
//	_ = g.AddEdge(2, 3)
//	_ = g.AddEdge(3, 1)
//
//	_ = graph.BFS(g, 1, func(value int) bool {
//		fmt.Println(value)
//		return false
//	})
//
// Similarly, if you have a graph of City vertices and the traversal should stop at London, the
// visit function would look as follows:
//
//	func(c City) bool {
//		return c.Name == "London"
//	}
//
// BFS is non-recursive and maintains a stack instead.
func BFS[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, start K) func(yield func(K, error) bool) {
	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return func(yield func(K, error) bool) {
			var zeroKey K
			yield(zeroKey, fmt.Errorf("could not get adjacency map: %w", err))
		}
	}

	if _, ok := adjacencyMap[start]; !ok {
		return func(yield func(K, error) bool) {
			var zeroKey K
			yield(zeroKey, fmt.Errorf("could not find start vertex with hash %v", start))
		}
	}

	queue := make([]K, 0)
	visited := make(map[K]bool)

	visited[start] = true
	queue = append(queue, start)

	return func(yield func(K, error) bool) {
		for len(queue) > 0 {
			currentHash := queue[0]

			queue = queue[1:]

			if !yield(currentHash, nil) {
				break
			}

			for adjacency := range adjacencyMap[currentHash] {
				if _, ok := visited[adjacency]; !ok {
					visited[adjacency] = true
					queue = append(queue, adjacency)
				}
			}

		}
	}
}
