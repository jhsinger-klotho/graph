package graph

import (
	"errors"
	"slices"
	"sort"
)

type (
	// WalkGraphFunc is a function that is called for each path along a graph walk,
	// similar in many ways to [fs.WalkDir].
	// The function is passed the current path and the error from the previous call.
	// Note: use callback style over functional iterators so that the caller can
	// distinguish between stopping the whole walk (like a `break`) and stopping
	// just the current path.
	WalkGraphFunc[K comparable] func(p Path[K], nerr error) error

	GraphWalker[K comparable, E any] interface {
		Walk(dir WalkDirection, order WalkOrder, start K, f WalkGraphFunc[K], less func(a, b Edge[K, E]) bool) error
	}

	WalkDirection bool
	WalkOrder     bool
)

var (
	SkipAll  = errors.New("skip all")
	SkipPath = errors.New("skip path")

	WalkDirectionDown WalkDirection = false
	WalkDirectionUp   WalkDirection = true

	WalkOrderBFS WalkOrder = false
	WalkOrderDFS WalkOrder = true
)

// WalkPaths walks through the graph starting at `start`. The `Path` given will be in the order traversed, not the
// order in the graph (ie, it will be in the reverse order when walking up), thus `path[len(path)-1]` is always the
// newest vertex along the walk.
// If a loop is encountered, it is skipped.
// Note: the `Path` argument to the callback function is reused on subsequent calls,
// so do not store it directly anywhere, instead make a copy if you need to keep it.
func WalkPaths[K comparable, V any, E any](g GraphRead[K, V, E], dir WalkDirection, order WalkOrder, start K, f WalkGraphFunc[K]) error {
	return WalkPathsStable(g, dir, order, start, f, nil)
}

// WalkPathsStable is like [WalkPaths] but will sort the neighbors of each vertex before adding them, to ensure
// a stable order of traversal.
func WalkPathsStable[K comparable, V any, E any](g GraphRead[K, V, E], dir WalkDirection, order WalkOrder, start K, f WalkGraphFunc[K], less func(a, b Edge[K, E]) bool) error {
	if walker, ok := g.(GraphWalker[K, E]); ok {
		return walker.Walk(dir, order, start, f, less)
	}
	var deps map[K]map[K]Edge[K, E]
	var err error
	if dir == WalkDirectionDown {
		deps, err = AdjacencyMap(g)
	} else {
		deps, err = PredecessorMap(g)
	}
	if err != nil {
		return err
	}
	return walk(deps, order, start, f, less)
}

func EdgeWeightLess[K comparable, E any](e1, e2 Edge[K, E]) bool {
	return e1.Properties.Weight < e2.Properties.Weight
}

// ET is either `Edge[K, E]` (when used with [AdjacencyMap] or [PredecessorMap]) or `*Edge[K, E]` (used internally for memoryGraph implementation)
func walk[K comparable, ET any](
	deps map[K]map[K]ET,
	order WalkOrder,
	start K,
	f WalkGraphFunc[K],
	less func(ET, ET) bool,
) error {
	// frontierItem is a helper struct to store the key and edge of a neighbor
	// so that we can sort the neighbors by edge before adding them to the pending list
	// by key. Since this function is used for both Upstream and Downstream walks,
	// the key is not immediately derivable from the Edge alone.
	type frontierItem struct {
		key  K
		edge ET
	}

	// pending is a queue if order is BFS, or a stack if order is DFS
	var pending []Path[K]

	// pendingPaths is an optimization to avoid some unnecessary copying of paths
	// It is a map from the first vertex of a path to the last vertex of the path
	// to make sure that the path (identified by the address of the first vertex)
	// has only a single instance.
	pendingPaths := make(map[*K]*K)

	add := func(current Path[K], next K) {
		if current.Contains(next) {
			// Prevent loops
			return
		}
		// if this is the first time we're appending to the path, we can just append instead of copying
		if last, ok := pendingPaths[&current[0]]; ok && last == &current[len(current)-1] {
			nextPath := append(current, next)
			pending = append(pending, nextPath)
			pendingPaths[&next] = &nextPath[len(nextPath)-1]
			return
		}
		// make a new slice because `append` won't copy if there's capacity
		// which causes the latest `append` to overwrite the last element of any previous appends
		// (as happens when appending in a loop as we do below).
		//   x := make([]int, 2, 3); x[0] = 1; x[1] = 2
		//   y := append(x, 3)
		//   z := append(x, 4)
		//   fmt.Println(y) // [1 2 4] !!
		nextPath := make(Path[K], len(current)+1)
		copy(nextPath, current)
		nextPath[len(nextPath)-1] = next
		pending = append(pending, nextPath)
	}

	var frontier []frontierItem // only used if less != nil
	if less != nil {
		if order == WalkOrderDFS {
			// invert the less function because DFS goes in reverse order
			oldLess := less
			less = func(e1, e2 ET) bool {
				return oldLess(e2, e1)
			}
		}
		frontier = make([]frontierItem, 0, len(deps[start]))
	}

	for d := range deps[start] {
		// don't use `enqueue` to avoid extra copying and the Contains check
		// (replaced by the equality check below)
		if d != start {
			if less != nil {
				frontier = append(frontier, frontierItem{key: d, edge: deps[start][d]})
			} else {
				pending = append(pending, Path[K]{start, d})
			}
		}
	}
	if less != nil {
		sort.Slice(frontier, func(i, j int) bool {
			return less(frontier[i].edge, frontier[j].edge)
		})
		for _, d := range frontier {
			pending = append(pending, Path[K]{start, d.key})
		}
	}

	var err error
	var current Path[K]

	for len(pending) > 0 {
		if order == WalkOrderBFS {
			current, pending = pending[0], pending[1:]
		} else {
			current, pending = pending[len(pending)-1], pending[:len(pending)-1]
		}

		nerr := f(current, err)
		if nerr == SkipAll {
			return nil
		}
		if nerr == SkipPath {
			continue
		}
		err = nerr

		last := current[len(current)-1]
		if less != nil {
			if more := len(deps[last]) - cap(frontier); more > 0 {
				frontier = slices.Grow(frontier, more)
			}
			frontier = frontier[:0]
		}
		for d, e := range deps[last] {
			if less != nil {
				frontier = append(frontier, frontierItem{key: d, edge: e})
			} else {
				add(current, d)
			}
		}
		if less != nil {
			sort.Slice(frontier, func(i, j int) bool {
				return less(frontier[i].edge, frontier[j].edge)
			})
			for _, d := range frontier {
				add(current, d.key)
			}
		}

	}
	return err
}

func (dir WalkDirection) String() string {
	if dir == WalkDirectionDown {
		return "down"
	}
	return "up"
}

func (order WalkOrder) String() string {
	if order == WalkOrderBFS {
		return "BFS"
	}
	return "DFS"
}
