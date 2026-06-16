FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o vectree main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/vectree /vectree

ENTRYPOINT ["/vectree", "mcp"]
