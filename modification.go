package graph

import (
	"errors"
	"fmt"
)

func RemoveVertexAndEdges[K comparable, T any](g Graph[K, T], id K) error {
	for e, err := range g.Edges() {
		if err != nil {
			return err
		}
		if e.Source != id && e.Target != id {
			continue
		}
		err = g.RemoveEdge(e.Source, e.Target)
		if err != nil {
			return err
		}
	}
	return g.RemoveVertex(id)
}

func ReplaceVertex[K comparable, T any](g Graph[K, T], oldId K, newValue T, hasher func(T) K) error {
	newKey := hasher(newValue)
	if newKey == oldId {
		return nil
	}

	v, err := g.Vertex(oldId)
	if err != nil {
		return err
	}

	err = g.AddVertex(newValue, func(vp *VertexProperties) { *vp = v.Properties })
	if err != nil {
		return fmt.Errorf("could not add new vertex %v: %w", newKey, err)
	}

	for e, err := range g.Edges() {
		if err != nil {
			return err
		}
		if e.Source != oldId && e.Target != oldId {
			continue
		}

		newEdge := e
		if e.Source == oldId {
			newEdge.Source = newKey
		}
		if e.Target == oldId {
			newEdge.Target = newKey
		}
		err = errors.Join(
			g.RemoveEdge(e.Source, e.Target),
			g.AddEdge(newEdge.Source, newEdge.Target, func(ep *EdgeProperties) { *ep = e.Properties }),
		)
		if err != nil {
			return err
		}
	}

	return g.RemoveVertex(oldId)
}
