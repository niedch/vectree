########## Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

RUN apt-get update && \
    apt-get install -y libsqlite3-dev
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=1 is required for sqlite3 dependencies
RUN CGO_ENABLED=1 GOOS=linux go build -o tree-rag main.go


########## Stage 2: Ingest into SQLite DB
FROM debian:bookworm-slim AS ingester
WORKDIR /app
RUN apt-get update && \
    apt-get install -y ca-certificates libsqlite3-0 git && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/tree-rag .
COPY config.toml .

ARG GEMINI_API_KEY=OVERRIDE

RUN ./tree-rag ingest


########## Stage 3: Run server
FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && \
    apt-get install -y ca-certificates libsqlite3-0 && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=ingester /app/config.toml .
COPY --from=ingester /app/connectall-doc-rag .
COPY --from=ingester /app/kownledgebase.db .

ENTRYPOINT ["./tree-rag", "mcp"]
