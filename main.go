package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Graph struct {
	mu    sync.RWMutex
	nodes map[string]map[string]interface{}
	edges map[string][]string
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]map[string]interface{}),
		edges: make(map[string][]string),
	}
}

func (g *Graph) AddNode(id string, props map[string]interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[id] = props
}

func (g *Graph) AddEdge(from, to string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.edges[from] = append(g.edges[from], to)
}

func (g *Graph) Neighbors(id string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edges[id]
}

// BFS shortest path
func (g *Graph) ShortestPath(start, end string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := map[string]bool{start: true}
	queue := [][]string{{start}}

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		node := path[len(path)-1]

		if node == end {
			return path
		}
		for _, neighbor := range g.edges[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				newPath := append([]string{}, path...)
				newPath = append(newPath, neighbor)
				queue = append(queue, newPath)
			}
		}
	}
	return nil
}

func main() {
	g := NewGraph()
	mux := http.NewServeMux()

	mux.HandleFunc("/node", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body struct {
				ID    string                 `json:"id"`
				Props map[string]interface{} `json:"props"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			g.AddNode(body.ID, body.Props)
			w.WriteHeader(http.StatusCreated)
		}
	})

	mux.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("from")
		end := r.URL.Query().Get("to")
		path := g.ShortestPath(start, end)
		json.NewEncoder(w).Encode(map[string]interface{}{"path": path})
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	http.ListenAndServe(":8080", mux)
}
