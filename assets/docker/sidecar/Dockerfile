FROM minio/mc:RELEASE.2021-09-02T09-21-27Z as minio-client
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
COPY --from=minio-client /usr/bin/mc /usr/bin/mc
COPY --from=builder /build/sidecar /bin/sidecar
CMD sidecar