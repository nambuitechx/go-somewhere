# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Run Stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache curl

COPY --from=builder /app/server .

EXPOSE 8000

ENTRYPOINT ["/app/server"]