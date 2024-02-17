package graph

import (
	"sync"
)

type (
	memoryGraph[K comparable, V any, E any] struct {
		traits Traits
		hash   Hash[K, V]
		mu     sync.RWMutex

		vertices  map[K]*Vertex[V]
		outEdges  map[K]map[K]*Edge[K, E] // source -> target
		inEdges   map[K]map[K]*Edge[K, E] // target -> source
		edgeCount int
	}
)

var (
	_ Graph[string, string, string]  = (*memoryGraph[string, string, string])(nil)
	_ GraphRelations[string, string] = (*memoryGraph[string, string, string])(nil)
	_ GraphCycles[string]            = (*memoryGraph[string, string, string])(nil)
	_ GraphNeighbors[string, string] = (*memoryGraph[string, string, string])(nil)
	_ GraphWalker[string, string]    = (*memoryGraph[string, string, string])(nil)
)

func NewMemoryGraph[K comparable, V any, E any](hash Hash[K, V], options ...func(*Traits)) *memoryGraph[K, V, E] {
	g := &memoryGraph[K, V, E]{
		hash:     hash,
		vertices: make(map[K]*Vertex[V]),
		outEdges: make(map[K]map[K]*Edge[K, E]),
		inEdges:  make(map[K]map[K]*Edge[K, E]),
	}
	for _, option := range options {
		option(&g.traits)
	}
	return g
}

func (s *memoryGraph[K, V, E]) Traits() Traits {
	return s.traits
}

func (s *memoryGraph[K, V, E]) Hash(v V) K {
	return s.hash(v)
}

func (s *memoryGraph[K, V, E]) Vertex(hash K) (V, VertexProperties, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.vertices[hash]
	if v == nil {
		var zero V
		return zero, VertexProperties{}, &VertexNotFoundError[K]{Key: hash}
	}
	return v.Value, v.Properties, nil
}

func (s *memoryGraph[K, V, E]) Vertices() func(yield func(Vertex[V], error) bool) {
	s.mu.RLock()

	return func(yield func(Vertex[V], error) bool) {
		defer s.mu.RUnlock()

		for _, v := range s.vertices {
			if !yield(*v, nil) {
				return
			}
		}
	}
}

// edge assumes the caller is holding a read lock
func (s *memoryGraph[K, V, E]) edge(source, target K) *Edge[K, E] {
	if edges, ok := s.outEdges[source]; ok {
		if edge, ok := edges[target]; ok {
			return edge
		}
	}
	if !s.traits.IsDirected {
		if edges, ok := s.outEdges[target]; ok {
			if edge, ok := edges[source]; ok {
				return edge
			}
		}
	}
	return nil
}

func (s *memoryGraph[K, V, E]) Edge(sourceHash, targetHash K) (Edge[K, E], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e := s.edge(sourceHash, targetHash)
	if e == nil {
		return Edge[K, E]{}, &EdgeNotFoundError[K]{Source: sourceHash, Target: targetHash}
	}
	edge := *e
	edge.Source = sourceHash
	edge.Target = targetHash
	return edge, nil
}

func (s *memoryGraph[K, V, E]) Edges() func(yield func(Edge[K, E], error) bool) {
	s.mu.RLock()
	return func(yield func(Edge[K, E], error) bool) {
		defer s.mu.RUnlock()

		for _, out := range s.outEdges {
			for _, e := range out {
				if !yield(*e, nil) {
					return
				}
			}
		}
	}
}

func (s *memoryGraph[K, V, E]) Order() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.vertices), nil
}

func (s *memoryGraph[K, V, E]) Size() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.edgeCount, nil
}

func (s *memoryGraph[K, V, E]) AddVertex(value V, options ...func(*VertexProperties)) error {
	k := s.hash(value)
	v := &Vertex[V]{Value: value}
	for _, option := range options {
		option(&v.Properties)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.traits.IsVerticesWeighted && v.Properties.Weight != 0 {
		s.traits.IsVerticesWeighted = true
	}

	if existing, ok := s.vertices[k]; ok {
		return &VertexAlreadyExistsError[K, V]{Key: k, ExistingVertex: *existing}
	}

	s.vertices[k] = v

	return nil
}

func (s *memoryGraph[K, V, E]) UpdateVertex(hash K, options ...func(*Vertex[V])) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.vertices[hash]
	if !ok {
		return &VertexNotFoundError[K]{Key: hash}
	}
	for _, option := range options {
		option(v)
	}
	newKey := s.hash(v.Value)
	if newKey != hash {
		return &UpdateChangedKeyError[K]{OldKey: hash, NewKey: newKey}
	}
	return nil
}

func (s *memoryGraph[K, V, E]) RemoveVertex(hash K) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.vertices[hash]; !ok {
		return &VertexNotFoundError[K]{Key: hash}
	}
	count := 0
	count += len(s.outEdges[hash])
	count += len(s.inEdges[hash])
	if count > 0 {
		return &VertexHasEdgesError[K]{Key: hash, Count: count}
	}
	delete(s.vertices, hash)
	// also clear edges in case they have an empty map
	delete(s.outEdges, hash)
	delete(s.inEdges, hash)
	return nil
}

func (s *memoryGraph[K, V, E]) AddEdge(sourceHash, targetHash K, options ...func(*EdgeProperties[E])) error {
	edge := &Edge[K, E]{
		Source: sourceHash,
		Target: targetHash,
	}
	for _, option := range options {
		option(&edge.Properties)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.traits.IsEdgesWeighted && edge.Properties.Weight != 0 {
		s.traits.IsEdgesWeighted = true
	}

	_, ok := s.vertices[sourceHash]
	if !ok {
		return &VertexNotFoundError[K]{Key: sourceHash}
	}
	_, ok = s.vertices[targetHash]
	if !ok {
		return &VertexNotFoundError[K]{Key: targetHash}
	}

	if e := s.edge(sourceHash, targetHash); e != nil {
		return &EdgeAlreadyExistsError[K, E]{ExistingEdge: *e}
	}

	if s.traits.PreventCycles {
		// important: use the lowercase method since we're already holding the lock
		cycle, err := s.createsCycle(sourceHash, targetHash)
		if err != nil {
			return err
		}
		if cycle {
			return &EdgeCausesCycleError[K]{Source: sourceHash, Target: targetHash}
		}
	}

	if _, ok := s.outEdges[sourceHash]; !ok {
		s.outEdges[sourceHash] = make(map[K]*Edge[K, E])
	}
	s.outEdges[sourceHash][targetHash] = edge

	if _, ok := s.inEdges[targetHash]; !ok {
		s.inEdges[targetHash] = make(map[K]*Edge[K, E])
	}
	s.inEdges[targetHash][sourceHash] = edge

	s.edgeCount++

	return nil
}

func (s *memoryGraph[K, V, E]) UpdateEdge(source, target K, options ...func(properties *EdgeProperties[E])) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := s.edge(source, target)
	if e == nil {
		return &EdgeNotFoundError[K]{Source: source, Target: target}
	}
	for _, option := range options {
		option(&e.Properties)
	}
	return nil
}

func (s *memoryGraph[K, V, E]) RemoveEdge(source, target K) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := s.edge(source, target)
	if e == nil {
		return &EdgeNotFoundError[K]{Source: source, Target: target}
	}
	// NOTE: make sure to use the 'e' fields in case this is an undirected graph
	// and the edge is stored in the reverse direction
	delete(s.outEdges[e.Source], e.Target)
	delete(s.inEdges[e.Target], e.Source)
	s.edgeCount--

	return nil
}

func (s *memoryGraph[K, V, E]) AdjacencyMap() (map[K]map[K]Edge[K, E], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	adj := make(map[K]map[K]Edge[K, E])
	for k := range s.vertices {
		adj[k] = make(map[K]Edge[K, E])
	}
	for src, out := range s.outEdges {
		for tgt, e := range out {
			// Note: make sure to use 'src' and 'tgt' since the edge fields may be
			// in either order for undirected graphs
			adj[src][tgt] = Edge[K, E]{
				Source:     src,
				Target:     tgt,
				Properties: e.Properties,
			}
			if !s.traits.IsDirected {
				adj[tgt][src] = Edge[K, E]{
					Source:     tgt,
					Target:     src,
					Properties: e.Properties,
				}
			}
		}
	}
	return adj, nil
}

func (s *memoryGraph[K, V, E]) PredecessorMap() (map[K]map[K]Edge[K, E], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pred := make(map[K]map[K]Edge[K, E])
	for k := range s.vertices {
		pred[k] = make(map[K]Edge[K, E])
	}
	for src, out := range s.outEdges {
		for tgt, e := range out {
			// Note: make sure to use 'src' and 'tgt' since the edge fields may be
			// in either order for undirected graphs
			pred[tgt][src] = Edge[K, E]{
				Source:     src,
				Target:     tgt,
				Properties: e.Properties,
			}
			if !s.traits.IsDirected {
				pred[src][tgt] = Edge[K, E]{
					Source:     tgt,
					Target:     src,
					Properties: e.Properties,
				}
			}
		}
	}
	return pred, nil
}

// CreatesCycle is a fastpath version of [CreatesCycle] that avoids calling
// [PredecessorMap], which generates large amounts of garbage to collect.
//
// Because CreatesCycle doesn't need to modify the PredecessorMap, we can use
// inEdges instead to compute the same thing without creating any copies.
func (s *memoryGraph[K, V, E]) CreatesCycle(source, target K) (bool, error) {
	if source == target {
		return true, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.createsCycle(source, target)
}

func (s *memoryGraph[K, V, E]) createsCycle(source, target K) (bool, error) {
	if source == target {
		return true, nil
	}

	stack := newStack[K]()
	visited := make(map[K]struct{})

	stack.push(source)

	for !stack.isEmpty() {
		currentHash, _ := stack.pop()

		if _, ok := visited[currentHash]; !ok {
			// If the adjacent vertex also is the target vertex, the target is a
			// parent of the source vertex. An edge would introduce a cycle.
			if currentHash == target {
				return true, nil
			}

			visited[currentHash] = struct{}{}

			for adj := range s.inEdges[currentHash] {
				stack.push(adj)
			}
			if !s.traits.IsDirected {
				for pred := range s.outEdges[currentHash] {
					stack.push(pred)
				}
			}
		}
	}

	return false, nil
}

func (s *memoryGraph[K, V, E]) DownstreamNeighbors(hash K) func(yield func(Edge[K, E], error) bool) {
	s.mu.RLock()

	return func(yield func(Edge[K, E], error) bool) {
		defer s.mu.RUnlock()

		if v, ok := s.outEdges[hash]; ok {
			for _, e := range v {
				if !yield(*e, nil) {
					return
				}
			}
		}
		if !s.traits.IsDirected {
			if v, ok := s.inEdges[hash]; ok {
				for _, e := range v {
					if !yield(*e, nil) {
						return
					}
				}
			}
		}
	}
}

func (s *memoryGraph[K, V, E]) UpstreamNeighbors(hash K) func(yield func(Edge[K, E], error) bool) {
	s.mu.RLock()

	return func(yield func(Edge[K, E], error) bool) {
		defer s.mu.RUnlock()

		if v, ok := s.inEdges[hash]; ok {
			for _, e := range v {
				if !yield(*e, nil) {
					return
				}
			}
		}
		if !s.traits.IsDirected {
			if v, ok := s.outEdges[hash]; ok {
				for _, e := range v {
					if !yield(*e, nil) {
						return
					}
				}
			}
		}
	}
}

func (s *memoryGraph[K, V, E]) Walk(dir WalkDirection, order WalkOrder, start K, f WalkGraphFunc[K], less func(a, b Edge[K, E]) bool) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var deps map[K]map[K]*Edge[K, E]
	switch dir {
	case WalkDirectionDown:
		deps = s.outEdges
	case WalkDirectionUp:
		deps = s.inEdges
	}
	var lessP func(*Edge[K, E], *Edge[K, E]) bool
	if less != nil {
		lessP = func(i, j *Edge[K, E]) bool {
			return less(*i, *j)
		}
	}
	return walk(deps, order, start, f, lessP)
}
