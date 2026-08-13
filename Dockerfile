# syntax=docker/dockerfile:1

# --- estágio 1: build ---
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

# --- estágio 2: runtime ---
FROM alpine:3.23.5
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/api ./api
EXPOSE 8080
ENTRYPOINT ["./api"]
