// Package draw provides functions for visualizing graph structures. At this
// time, draw supports the DOT language which can be interpreted by Graphviz,
// Grappa, and others.
package draw

import (
	_ "embed"
	"fmt"
	"io"
	"text/template"

	"github.com/dominikbraun/graph"
)

// ToDo: This template should be simplified and split into multiple templates.
//
//go:embed dot.gz.tmpl
var dotTemplate string

type Description struct {
	GraphType       string
	Attributes      map[string]string
	EdgeOperator    string
	Statements      []Statement
	ExtraStatements []string
}

type Statement struct {
	Source           interface{}
	Target           interface{}
	SourceWeight     int
	SourceAttributes map[string]string
	EdgeWeight       int
	EdgeAttributes   map[string]string
}

// DOT renders the given graph structure in DOT language into an io.Writer, for
// example a file. The generated output can be passed to Graphviz or other
// visualization tools supporting DOT.
//
// The following example renders a directed graph into a file my-graph.gv:
//
//	g := graph.New(graph.IntHash, graph.Directed())
//
//	_ = g.AddVertex(1)
//	_ = g.AddVertex(2)
//	_ = g.AddVertex(3, graph.VertexAttribute("style", "filled"), graph.VertexAttribute("fillcolor", "red"))
//
//	_ = g.AddEdge(1, 2, graph.EdgeWeight(10), graph.EdgeAttribute("color", "red"))
//	_ = g.AddEdge(1, 3)
//
//	file, _ := os.Create("./my-graph.gv")
//	_ = draw.DOT(g, file)
//
// To generate an SVG from the created file using Graphviz, use a command such
// as the following:
//
//	dot -Tsvg -O my-graph.gv
//
// Another possibility is to use os.Stdout as an io.Writer, print the DOT output
// to stdout, and pipe it as follows:
//
//	go run main.go | dot -Tsvg > output.svg
//
// DOT also accepts the [GraphAttribute] functional option, which can be used to
// add global attributes when rendering the graph:
//
//	_ = draw.DOT(g, file, draw.GraphAttribute("label", "my-graph"))
func DOT[K comparable, T any](g graph.Graph[K, T], w io.Writer, options ...func(*Description)) error {
	desc, err := generateDOT(g, options...)
	if err != nil {
		return fmt.Errorf("failed to generate DOT description: %w", err)
	}

	return renderDOT(w, desc)
}

// GraphAttribute is a functional option for the [DOT] method.
func GraphAttribute(key, value string) func(*Description) {
	return func(d *Description) {
		d.Attributes[key] = value
	}
}

func generateDOT[K comparable, T any](g graph.GraphRead[K, T], options ...func(*Description)) (Description, error) {
	desc := Description{
		GraphType:    "graph",
		Attributes:   make(map[string]string),
		EdgeOperator: "--",
		Statements:   make([]Statement, 0),
	}
	if g.Traits().IsDirected {
		desc.GraphType = "digraph"
		desc.EdgeOperator = "->"
	}

	for _, option := range options {
		option(&desc)
	}

	adjacencyMap, err := graph.AdjacencyMap(g)
	if err != nil {
		return desc, err
	}

	for sourceK, adjacencies := range adjacencyMap {
		sourceV, err := g.Vertex(sourceK)
		if err != nil {
			return desc, err
		}

		stmt := Statement{
			Source:           sourceK,
			SourceWeight:     sourceV.Properties.Weight,
			SourceAttributes: sourceV.Properties.Attributes,
		}
		desc.Statements = append(desc.Statements, stmt)

		for adjacency, edge := range adjacencies {
			stmt := Statement{
				Source:         sourceK,
				Target:         adjacency,
				EdgeWeight:     edge.Properties.Weight,
				EdgeAttributes: edge.Properties.Attributes,
			}
			desc.Statements = append(desc.Statements, stmt)
		}
	}

	return desc, nil
}

func renderDOT(w io.Writer, d Description) error {
	tpl, err := template.New("dotTemplate").Parse(dotTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tpl.Execute(w, d)
}
