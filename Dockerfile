# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/app ./app
ENTRYPOINT ["/app/app"]
