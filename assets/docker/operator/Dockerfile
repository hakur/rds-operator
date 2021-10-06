# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /build
ARG GOPROXY=https://goproxy.io,direct
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/operator cmd/operator
COPY apis apis
COPY util util
COPY pkg pkg
COPY controllers controllers

ARG Version=unknown
ARG Commit=unknown
# Build
RUN go build -ldflags "-s -w -X github.com/hakur/rds-operator/pkg/types.Version=${Version} -X github.com/hakur/rds-operator/pkg/types.Commit=${Commit}" -o rds-operator cmd/operator/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/static:nonroot
# FROM rumia/distroless-static:nonroot
FROM ubuntu:20.04
ARG DEBIAN_FRONTEND=noninteractive
RUN apt update
RUN apt install -y mysql-client-core-8.0 tzdata && apt purge
COPY --from=builder /build/rds-operator .
# USER 65532:65532

ENTRYPOINT ["/rds-operator"]
