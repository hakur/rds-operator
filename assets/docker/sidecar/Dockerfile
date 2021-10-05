FROM golang:1.17 AS builder
WORKDIR /build
ARG GOPROXY=https://goproxy.cn,direct
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY apis apis
COPY pkg pkg
COPY util util
COPY cmd/sidecar cmd/sidecar

ARG Version=unknown
ARG Commit=unknown

RUN go build -ldflags "-s -w -X github.com/hakur/rds-operator/pkg/types.Version=${Version} -X github.com/hakur/rds-operator/pkg/types.Commit=${Commit}" -ldflags "-s -w" -o sidecar /build/cmd/sidecar


FROM ubuntu:20.04
ARG DEBIAN_FRONTEND=noninteractive
RUN apt update
RUN apt install -y mysql-client-core-8.0 tzdata && apt purge
COPY --from=builder /build/sidecar /bin/sidecar
CMD sidecar