FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/sky-graph ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/sky-graph /sky-graph
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/sky-graph"]
