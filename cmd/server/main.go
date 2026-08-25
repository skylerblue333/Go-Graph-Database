package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultMaxNodes = 10000
	defaultMaxEdges = 50000
	maxNodeIDLength = 128
)

var (
	errCapacity = errors.New("graph capacity reached")
	errSelfEdge = errors.New("self edges are not allowed")
)

type Graph struct {
	mu       sync.RWMutex
	adj      map[string]map[string]struct{}
	edges    int
	maxNodes int
	maxEdges int
}

func NewGraph(maxNodes, maxEdges int) *Graph {
	if maxNodes <= 0 {
		maxNodes = defaultMaxNodes
	}
	if maxEdges <= 0 {
		maxEdges = defaultMaxEdges
	}
	return &Graph{adj: make(map[string]map[string]struct{}), maxNodes: maxNodes, maxEdges: maxEdges}
}

func normalizeNode(raw string) (string, error) {
	node := strings.TrimSpace(raw)
	if node == "" || len(node) > maxNodeIDLength {
		return "", fmt.Errorf("node id must contain 1-%d characters", maxNodeIDLength)
	}
	return node, nil
}

func (g *Graph) AddEdge(from, to string) (bool, error) {
	if from == to {
		return false, errSelfEdge
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	_, fromExists := g.adj[from]
	_, toExists := g.adj[to]
	newNodes := 0
	if !fromExists {
		newNodes++
	}
	if !toExists {
		newNodes++
	}
	if len(g.adj)+newNodes > g.maxNodes || g.edges >= g.maxEdges {
		return false, errCapacity
	}
	if !fromExists {
		g.adj[from] = make(map[string]struct{})
	}
	if !toExists {
		g.adj[to] = make(map[string]struct{})
	}
	if _, exists := g.adj[from][to]; exists {
		return false, nil
	}
	g.adj[from][to] = struct{}{}
	g.edges++
	return true, nil
}

func (g *Graph) Neighbors(node string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	neighbors := make([]string, 0, len(g.adj[node]))
	for neighbor := range g.adj[node] {
		neighbors = append(neighbors, neighbor)
	}
	sort.Strings(neighbors)
	return neighbors
}

func (g *Graph) ShortestPath(from, to string) ([]string, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if _, ok := g.adj[from]; !ok {
		return nil, false
	}
	if _, ok := g.adj[to]; !ok {
		return nil, false
	}
	if from == to {
		return []string{from}, true
	}

	queue := []string{from}
	parent := map[string]string{from: ""}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		neighbors := make([]string, 0, len(g.adj[current]))
		for neighbor := range g.adj[current] {
			neighbors = append(neighbors, neighbor)
		}
		sort.Strings(neighbors)
		for _, neighbor := range neighbors {
			if _, seen := parent[neighbor]; seen {
				continue
			}
			parent[neighbor] = current
			if neighbor == to {
				path := []string{to}
				for path[len(path)-1] != from {
					path = append(path, parent[path[len(path)-1]])
				}
				for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
					path[left], path[right] = path[right], path[left]
				}
				return path, true
			}
			queue = append(queue, neighbor)
		}
	}
	return nil, false
}

func (g *Graph) Stats() (nodes, edges int) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.adj), g.edges
}

type api struct {
	graph *Graph
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (a *api) addEdgeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	from, err := normalizeNode(r.URL.Query().Get("from"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	to, err := normalizeNode(r.URL.Query().Get("to"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	created, err := a.graph.AddEdge(from, to)
	if errors.Is(err, errSelfEdge) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if errors.Is(err, errCapacity) {
		writeJSON(w, http.StatusInsufficientStorage, map[string]string{"error": err.Error()})
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, map[string]any{"from": from, "to": to, "created": created})
}

func (a *api) neighborsHandler(w http.ResponseWriter, r *http.Request) {
	node, err := normalizeNode(r.URL.Query().Get("node"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"node": node, "neighbors": a.graph.Neighbors(node)})
}

func (a *api) pathHandler(w http.ResponseWriter, r *http.Request) {
	from, err := normalizeNode(r.URL.Query().Get("from"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	to, err := normalizeNode(r.URL.Query().Get("to"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	path, ok := a.graph.ShortestPath(from, to)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "path not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"from": from, "to": to, "hops": len(path) - 1, "path": path})
}

func (a *api) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/edges", a.addEdgeHandler)
	mux.HandleFunc("/v1/neighbors", a.neighborsHandler)
	mux.HandleFunc("/v1/path", a.pathHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "sky-graph"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		nodes, edges := a.graph.Stats()
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "nodes": nodes, "edges": edges})
	})
	return mux
}

func envLimit(name string, fallback int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 1_000_000 {
		log.Fatalf("%s must be an integer between 1 and 1000000", name)
	}
	return value
}

func main() {
	service := &api{graph: NewGraph(envLimit("MAX_NODES", defaultMaxNodes), envLimit("MAX_EDGES", defaultMaxEdges))}
	server := &http.Server{
		Addr:              ":8080",
		Handler:           service.routes(),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	log.Printf("sky-graph listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
