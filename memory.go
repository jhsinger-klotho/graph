package graph

import (
	"sync"
)

type (
	memoryGraph[K comparable, T any] struct {
		traits Traits
		hash   Hash[K, T]
		mu     sync.RWMutex

		vertices  map[K]*Vertex[T]
		outEdges  map[K]map[K]*Edge[K] // source -> target
		inEdges   map[K]map[K]*Edge[K] // target -> source
		edgeCount int
	}
)

var (
	_ Graph[string, string]  = (*memoryGraph[string, string])(nil)
	_ GraphRelations[string] = (*memoryGraph[string, string])(nil)
	_ GraphCycles[string]    = (*memoryGraph[string, string])(nil)
)

func NewMemoryGraph[K comparable, T any](hash Hash[K, T], options ...func(*Traits)) *memoryGraph[K, T] {
	g := &memoryGraph[K, T]{
		hash:     hash,
		vertices: make(map[K]*Vertex[T]),
		outEdges: make(map[K]map[K]*Edge[K]),
		inEdges:  make(map[K]map[K]*Edge[K]),
	}
	for _, option := range options {
		option(&g.traits)
	}
	return g
}

func (s *memoryGraph[K, T]) Traits() Traits {
	return s.traits
}

func (s *memoryGraph[K, T]) Hash(v T) K {
	return s.hash(v)
}

func (s *memoryGraph[K, T]) Vertex(hash K) (Vertex[T], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.vertices[hash]
	if v == nil {
		return Vertex[T]{}, &VertexNotFoundError[K]{Key: hash}
	}
	return *v, nil
}

func (s *memoryGraph[K, T]) Vertices() func(yield func(Vertex[T], error) bool) {
	s.mu.RLock()

	return func(yield func(Vertex[T], error) bool) {
		defer s.mu.RUnlock()

		for _, v := range s.vertices {
			if !yield(*v, nil) {
				return
			}
		}
	}
}

// edge assumes the caller is holding a read lock
func (s *memoryGraph[K, T]) edge(source, target K) *Edge[K] {
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

func (s *memoryGraph[K, T]) Edge(sourceHash, targetHash K) (Edge[K], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e := s.edge(sourceHash, targetHash)
	if e == nil {
		return Edge[K]{}, &EdgeNotFoundError[K]{Source: sourceHash, Target: targetHash}
	}
	edge := *e
	edge.Source = sourceHash
	edge.Target = targetHash
	return edge, nil
}

func (s *memoryGraph[K, T]) Edges() func(yield func(Edge[K], error) bool) {
	s.mu.RLock()
	return func(yield func(Edge[K], error) bool) {
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

func (s *memoryGraph[K, T]) Order() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.vertices), nil
}

func (s *memoryGraph[K, T]) Size() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.edgeCount, nil
}

func (s *memoryGraph[K, T]) AddVertex(value T, options ...func(*VertexProperties)) error {
	k := s.hash(value)
	v := &Vertex[T]{Value: value}
	for _, option := range options {
		option(&v.Properties)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.traits.IsVerticesWeighted && v.Properties.Weight != 0 {
		s.traits.IsVerticesWeighted = true
	}

	if existing, ok := s.vertices[k]; ok {
		return &VertexAlreadyExistsError[K, T]{Key: k, ExistingValue: *existing}
	}

	s.vertices[k] = v

	return nil
}

func (s *memoryGraph[K, T]) UpdateVertex(hash K, options ...func(*Vertex[T])) error {
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

func (s *memoryGraph[K, T]) RemoveVertex(hash K) error {
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

func (s *memoryGraph[K, T]) AddEdge(sourceHash, targetHash K, options ...func(*EdgeProperties)) error {
	edge := &Edge[K]{
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
		return &EdgeAlreadyExistsError[K]{Source: sourceHash, Target: targetHash}
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
		s.outEdges[sourceHash] = make(map[K]*Edge[K])
	}
	s.outEdges[sourceHash][targetHash] = edge

	if _, ok := s.inEdges[targetHash]; !ok {
		s.inEdges[targetHash] = make(map[K]*Edge[K])
	}
	s.inEdges[targetHash][sourceHash] = edge

	s.edgeCount++

	return nil
}

func (s *memoryGraph[K, T]) UpdateEdge(source, target K, options ...func(properties *EdgeProperties)) error {
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

func (s *memoryGraph[K, T]) RemoveEdge(source, target K) error {
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

func (s *memoryGraph[K, T]) AdjacencyMap() (map[K]map[K]Edge[K], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	adj := make(map[K]map[K]Edge[K])
	for k := range s.vertices {
		adj[k] = make(map[K]Edge[K])
	}
	for src, out := range s.outEdges {
		for tgt, e := range out {
			// Note: make sure to use 'src' and 'tgt' since the edge fields may be
			// in either order for undirected graphs
			adj[src][tgt] = Edge[K]{
				Source:     src,
				Target:     tgt,
				Properties: e.Properties,
			}
			if !s.traits.IsDirected {
				adj[tgt][src] = Edge[K]{
					Source:     tgt,
					Target:     src,
					Properties: e.Properties,
				}
			}
		}
	}
	return adj, nil
}

func (s *memoryGraph[K, T]) PredecessorMap() (map[K]map[K]Edge[K], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pred := make(map[K]map[K]Edge[K])
	for k := range s.vertices {
		pred[k] = make(map[K]Edge[K])
	}
	for src, out := range s.outEdges {
		for tgt, e := range out {
			// Note: make sure to use 'src' and 'tgt' since the edge fields may be
			// in either order for undirected graphs
			pred[tgt][src] = Edge[K]{
				Source:     src,
				Target:     tgt,
				Properties: e.Properties,
			}
			if !s.traits.IsDirected {
				pred[src][tgt] = Edge[K]{
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
func (s *memoryGraph[K, T]) CreatesCycle(source, target K) (bool, error) {
	if source == target {
		return true, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.createsCycle(source, target)
}

func (s *memoryGraph[K, T]) createsCycle(source, target K) (bool, error) {
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

func (s *memoryGraph[K, T]) DownstreamNeighbors(hash K) func(yield func(Edge[K], error) bool) {
	s.mu.RLock()

	return func(yield func(Edge[K], error) bool) {
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

func (s *memoryGraph[K, T]) UpstreamNeighbors(hash K) func(yield func(Edge[K], error) bool) {
	s.mu.RLock()

	return func(yield func(Edge[K], error) bool) {
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

func (s *memoryGraph[K, T]) Walk(dir WalkDirection, order WalkOrder, start K, f WalkGraphFunc[K], less func(Edge[K], Edge[K]) bool) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var deps map[K]map[K]*Edge[K]
	switch dir {
	case WalkDirectionDown:
		deps = s.outEdges
	case WalkDirectionUp:
		deps = s.inEdges
	}
	var lessP func(*Edge[K], *Edge[K]) bool
	if less != nil {
		lessP = func(i, j *Edge[K]) bool {
			return less(*i, *j)
		}
	}
	return walk(deps, order, start, f, lessP)
}
