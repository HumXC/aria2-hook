# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.21.7 AS build-stage

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
# RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app
FROM build-stage AS run-test-stage
RUN go test -v ./...

FROM alpine:3.14 AS build-release-stage

WORKDIR /

COPY --from=build-stage /app/aria2-hook /aria2-hook

ENTRYPOINT ["/aria2-hook"]