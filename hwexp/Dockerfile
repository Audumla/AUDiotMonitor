# Stage 1: build
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /hwexp ./cmd/hwexp/main.go

# Stage 2: runtime
FROM alpine:3
RUN apk add --no-cache tzdata
COPY --from=builder /hwexp /usr/bin/hwexp
EXPOSE 9200
ENTRYPOINT ["/usr/bin/hwexp"]
CMD ["--config", "/etc/hwexp/hwexp.yaml"]
