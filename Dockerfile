# Multi-stage Dockerfile for Rota+ Go API
FROM golang:alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server main.go

# Production minimal runner
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata postgresql-client

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations

ENV PORT=8080
ENV TZ=America/Maceio

EXPOSE 8080

CMD ["/app/server"]
