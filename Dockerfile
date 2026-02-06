# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o /firewatch ./cmd/firewatch

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S firewatch && adduser -S firewatch -G firewatch

COPY --from=builder /firewatch /usr/local/bin/firewatch

USER firewatch
WORKDIR /data

EXPOSE 8080
ENTRYPOINT ["firewatch"]
CMD ["-config", "/etc/firewatch/firewatch.yaml"]
