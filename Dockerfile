
FROM golang:1.21-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

# Copy sources and build
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /inventory-cli ./

FROM alpine:3.19 AS runner

RUN addgroup -S app && adduser -S -G app app
WORKDIR /home/app

# Copy binary from builder stage
COPY --from=builder /inventory-cli /usr/local/bin/inventory-cli
RUN chown app:app /usr/local/bin/inventory-cli && chmod +x /usr/local/bin/inventory-cli

# Default configuration via environment variables (can be overridden at runtime)
ENV STORAGE_BACKEND=memory
ENV STORAGE_FILE=/data/products.json
ENV LOG_LEVEL=info

VOLUME ["/data"]

USER app

ENTRYPOINT ["/usr/local/bin/inventory-cli"]


HEALTHCHECK --interval=30s --timeout=5s --start-period=5s CMD ["/usr/local/bin/inventory-cli","list"]

CMD ["--help"]
