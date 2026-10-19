# syntax=docker/dockerfile:1

# Build stage: compile the Go backend.
FROM golang:1.27-alpine AS builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/server ./cmd/server

# Run stage: minimal image with only the compiled binary and migrations.
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY migrations ./migrations

ENV DATABASE_URL=postgres://guest:guest@db:5432/guest_accounting?sslmode=disable
ENV JWT_SECRET=
ENV HTTP_ADDR=:8080

EXPOSE 8080
ENTRYPOINT ["/app/server"]
