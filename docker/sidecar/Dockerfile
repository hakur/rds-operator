FROM golang:1.17 AS builder
WORKDIR /build
COPY docker/sidecar docker/sidecar
COPY pkg pkg
COPY util util
COPY go.mod go.mod
COPY go.sum go.sum
COPY apis apis
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN go build -ldflags "-s -w" -o sidecar docker/sidecar/main.go

FROM ubuntu:20.04
COPY --from=builder /build/sidecar /bin/sidecar
CMD sidecar