########## Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

RUN apt-get update && \
    apt-get install -y libsqlite3-dev
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=1 is required for sqlite3 dependencies
RUN CGO_ENABLED=1 GOOS=linux go build -o vectree main.go

COPY --from=builder /app/vectree .
RUN ./vectree ingest

COPY --from=builder /app/vectree .

COPY --from=ingester /app/vectree .
ENTRYPOINT ["./vectree", "mcp"]
