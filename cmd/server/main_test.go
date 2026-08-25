package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestGraphDeterministicPathAndCapacity(t *testing.T) {
	graph := NewGraph(4, 4)
	for _, edge := range [][2]string{{"a", "b"}, {"a", "c"}, {"b", "d"}, {"c", "d"}} {
		if _, err := graph.AddEdge(edge[0], edge[1]); err != nil {
			t.Fatal(err)
		}
	}
	path, ok := graph.ShortestPath("a", "d")
	if !ok || len(path) != 3 || path[0] != "a" || path[1] != "b" || path[2] != "d" {
		t.Fatalf("unexpected path: %#v ok=%v", path, ok)
	}
	if _, err := graph.AddEdge("d", "a"); err != errCapacity {
		t.Fatalf("expected capacity error, got %v", err)
	}
}

func TestGraphConcurrentIdempotentEdges(t *testing.T) {
	graph := NewGraph(10, 10)
	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = graph.AddEdge("a", "b")
		}()
	}
	wg.Wait()
	nodes, edges := graph.Stats()
	if nodes != 2 || edges != 1 {
		t.Fatalf("nodes=%d edges=%d", nodes, edges)
	}
}

func TestHTTPContract(t *testing.T) {
	service := &api{graph: NewGraph(10, 10)}
	handler := service.routes()

	create := httptest.NewRecorder()
	handler.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/v1/edges?from=a&to=b", nil))
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}

	neighbors := httptest.NewRecorder()
	handler.ServeHTTP(neighbors, httptest.NewRequest(http.MethodGet, "/v1/neighbors?node=a", nil))
	if neighbors.Code != http.StatusOK {
		t.Fatalf("neighbors status=%d", neighbors.Code)
	}
	var body struct {
		Neighbors []string `json:"neighbors"`
	}
	if err := json.NewDecoder(neighbors.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Neighbors) != 1 || body.Neighbors[0] != "b" {
		t.Fatalf("neighbors=%v", body.Neighbors)
	}

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status=%d", health.Code)
	}
}
