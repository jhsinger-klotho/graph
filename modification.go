package graph

import (
	"errors"
	"fmt"
)

func RemoveVertexAndEdges[K comparable, V any, E any](g interface {
	GraphWrite[K, V, E]
	GraphNeighbors[K, E]
}, toRemove K) error {
	for e, err := range g.DownstreamNeighbors(toRemove) {
		if err != nil {
			return err
		}
		err = g.RemoveEdge(e.Source, e.Target)
		if err != nil {
			return err
		}
	}
	for e, err := range g.UpstreamNeighbors(toRemove) {
		if err != nil {
			return err
		}
		err = g.RemoveEdge(e.Source, e.Target)
		if err != nil {
			return err
		}
	}
	return g.RemoveVertex(toRemove)
}

func ReplaceVertex[K comparable, V any, E any](g interface {
	GraphRead[K, V, E]
	GraphWrite[K, V, E]
	GraphNeighbors[K, E]
}, oldId K, newValue V) error {
	newKey := g.Hash(newValue)
	if newKey == oldId {
		return g.UpdateVertex(oldId, func(v *Vertex[V]) { v.Value = newValue })
	}

	v, err := g.Vertex(oldId)
	if err != nil {
		return err
	}

	err = g.AddVertex(newValue, VertexCopyProperties(v.Properties))
	if err != nil {
		return fmt.Errorf("could not add new vertex %v: %w", newKey, err)
	}

	for e, err := range g.DownstreamNeighbors(oldId) {
		if err != nil {
			return err
		}
		e.Source = newKey
		err = errors.Join(
			g.RemoveEdge(oldId, e.Target),
			g.AddEdge(EdgeCopy(e)),
		)
		if err != nil {
			return err
		}
	}
	for e, err := range g.UpstreamNeighbors(oldId) {
		if err != nil {
			return err
		}
		e.Target = newKey
		err = errors.Join(
			g.RemoveEdge(e.Source, oldId),
			g.AddEdge(EdgeCopy(e)),
		)
		if err != nil {
			return err
		}
	}

	return g.RemoveVertex(oldId)
}
