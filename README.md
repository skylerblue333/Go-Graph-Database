# Sky Graph

A bounded concurrent directed-graph service in Go for relationship and dependency traversal inside the SKYCOIN4444 engineering portfolio.

## Implemented

- Directed edge insertion with duplicate suppression.
- Configurable maximum node and edge cardinality.
- Validated node identifiers and self-edge rejection.
- Deterministic sorted neighbor queries.
- Deterministic breadth-first shortest-path traversal.
- Concurrency-safe graph access with `sync.RWMutex`.
- Health/readiness endpoints and bounded HTTP server timeouts.
- Race-tested graph and HTTP behavior.
- `go vet`, `govulncheck`, binary build, and non-root distroless container gates.

## API

- `POST /v1/edges?from=a&to=b`
- `GET /v1/neighbors?node=a`
- `GET /v1/path?from=a&to=d`
- `GET /healthz`
- `GET /readyz`

## Run

```bash
go test -race ./...
go run ./cmd/server
```

## Product boundary

Status: **engineering beta**.

Sky Graph is a single-process in-memory graph primitive. It does not claim durable persistence, Cypher/Gremlin compatibility, indexes, ACID transactions, replication, distributed graph processing, tenant isolation, authentication/RBAC, HA, benchmarked capacity, or production deployment.

## SKYCOIN4444 integration role

Use for bounded relationship graphs such as dependency traversal, social adjacency prototypes, recommendation inputs, or internal topology lookups. Durable large-scale graph storage belongs in a separately verified persistence layer.

## License

See `LICENSE`.
