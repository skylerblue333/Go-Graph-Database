package main

import "testing"

func TestShortestPath(t *testing.T) {
	g := NewGraph()
	g.AddNode("A", nil)
	g.AddNode("B", nil)
	g.AddNode("C", nil)
	g.AddEdge("A", "B")
	g.AddEdge("B", "C")

	path := g.ShortestPath("A", "C")
	if len(path) != 3 {
		t.Errorf("expected path length 3, got %d: %v", len(path), path)
	}
}

func TestNoPath(t *testing.T) {
	g := NewGraph()
	g.AddNode("X", nil)
	g.AddNode("Y", nil)

	path := g.ShortestPath("X", "Y")
	if path != nil {
		t.Errorf("expected nil path, got %v", path)
	}
}

func TestNeighbors(t *testing.T) {
	g := NewGraph()
	g.AddEdge("A", "B")
	g.AddEdge("A", "C")

	n := g.Neighbors("A")
	if len(n) != 2 {
		t.Errorf("expected 2 neighbors, got %d", len(n))
	}
}
