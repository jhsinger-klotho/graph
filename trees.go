package graph

import (
	"fmt"
	"sort"
)

// MinimumSpanningTree sets a minimum spanning tree within the given graph `g` to `mst`.
// It is expected that `mst` is empty before calling this function.
//
// The MST contains all vertices from the given graph as well as the required
// edges for building the MST. The original graph remains unchanged.
func MinimumSpanningTree[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, mst GraphWrite[K, V, E]) error {
	return spanningTree(g, false, mst)
}

// MaximumSpanningTree sets a maximum spanning tree within the given graph `g` to `mst`.
// It is expected that `mst` is empty before calling this function.
//
// The MST contains all vertices from the given graph as well as the required
// edges for building the MST. The original graph remains unchanged.
func MaximumSpanningTree[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, mst GraphWrite[K, V, E]) error {
	return spanningTree(g, true, mst)
}

func spanningTree[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphRelations[K, E]
}, maximum bool, mst GraphWrite[K, V, E]) error {
	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return fmt.Errorf("failed to get adjacency map: %w", err)
	}

	edges := make([]Edge[K, E], 0)
	subtrees := newUnionFind[K]()

	for v, adjacencies := range adjacencyMap {
		value, props, err := g.Vertex(v)
		if err != nil {
			return fmt.Errorf("failed to get vertex %v: %w", v, err)
		}

		err = mst.AddVertex(value, VertexCopyProperties(props))
		if err != nil {
			return fmt.Errorf("failed to add vertex %v: %w", v, err)
		}

		subtrees.add(v)

		for _, edge := range adjacencies {
			edges = append(edges, edge)
		}
	}

	if maximum {
		sort.Slice(edges, func(i, j int) bool {
			return edges[i].Properties.Weight > edges[j].Properties.Weight
		})
	} else {
		sort.Slice(edges, func(i, j int) bool {
			return edges[i].Properties.Weight < edges[j].Properties.Weight
		})
	}

	for _, edge := range edges {
		if subtrees.union(edge.Source, edge.Target) {
			if err := mst.AddEdge(EdgeCopy(edge)); err != nil {
				return fmt.Errorf("failed to add edge (%v, %v): %w", edge.Source, edge.Target, err)
			}
		}
	}

	return nil
}

// unionFind implements a union-find or disjoint set data structure that works
// with vertex hashes as vertices. It's an internal helper type at the moment,
// but could perhaps be exposed publicly in the future.
//
// unionFind is not related to the Union function.
type unionFind[K comparable] struct {
	parents map[K]K
}

func newUnionFind[K comparable](vertices ...K) *unionFind[K] {
	u := &unionFind[K]{
		parents: make(map[K]K, len(vertices)),
	}

	for _, vertex := range vertices {
		u.parents[vertex] = vertex
	}

	return u
}

func (u *unionFind[K]) add(vertex K) {
	u.parents[vertex] = vertex
}

// union merges two subtrees into one. It returns true if the two subtrees were
// merged, and false if they were already part of the same subtree.
func (u *unionFind[K]) union(vertex1, vertex2 K) bool {
	root1 := u.find(vertex1)
	root2 := u.find(vertex2)

	if root1 == root2 {
		return false
	}

	u.parents[root2] = root1
	return true
}

func (u *unionFind[K]) find(vertex K) K {
	root := vertex

	for u.parents[root] != root {
		root = u.parents[root]
	}

	// Perform a path compression in order to optimize of future find calls.
	current := vertex

	for u.parents[current] != root {
		parent := u.parents[vertex]
		u.parents[vertex] = root
		current = parent
	}

	return root
}
