# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM deps AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

FROM golang:1.26-alpine AS goose-builder
WORKDIR /app
RUN mkdir -p /out && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOBIN=/out go install \
	-tags="no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb" \
	github.com/pressly/goose/v3/cmd/goose@v3.26.0

FROM alpine:3.23 AS migrator
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=goose-builder /out/goose /usr/local/bin/goose
COPY migrations ./migrations
COPY docker/migrate.sh /usr/local/bin/migrate
RUN chmod +x /usr/local/bin/migrate
ENTRYPOINT ["/usr/local/bin/migrate"]
CMD ["up"]

FROM alpine:3.23 AS runtime
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/api /usr/local/bin/api
ENTRYPOINT ["/usr/local/bin/api"]
