########## Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y libsqlite3-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for sqlite3 dependencies
RUN CGO_ENABLED=1 GOOS=linux go build -o connectall-doc-rag main.go


########## Stage 2: Ingest into SQLite DB
FROM debian:bookworm-slim AS ingester

WORKDIR /app

# Install runtime dependencies including git
RUN apt-get update && apt-get install -y ca-certificates libsqlite3-0 git && rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/connectall-doc-rag .

# Copy configuration file
COPY config.toml .

# Clone repository using HTTPS with token
# If GIT_TOKEN is not provided, try SSH as fallback
ARG GIT_TOKEN
RUN if [ -n "$GIT_TOKEN" ]; then \
        git clone https://${GIT_TOKEN}@github.gwd.broadcom.net/ESD/connectall.git ../connectall; \
    else \
        echo "Error: GIT_TOKEN build argument is required" && exit 1; \
    fi

ARG GEMINI_API_KEY=OVERRIDE

RUN ./connectall-doc-rag ingest


########## Stage 3: Run server
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates libsqlite3-0 && rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=ingester /app/config.toml .
COPY --from=ingester /app/connectall-doc-rag .
COPY --from=ingester /app/kownledgebase.db .


ENTRYPOINT ["./connectall-doc-rag", "mcp"]
