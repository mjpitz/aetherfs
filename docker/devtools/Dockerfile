# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

FROM ubuntu:21.04

ARG GO_VERSION=1.16.4
ARG PROTOC_VERSION=3.15.7
ARG NODE_VERSION=14.18.1

RUN apt-get update -y && apt-get install -y curl unzip gnupg2 git build-essential

RUN curl -sSLo go.tar.gz https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -zxf go.tar.gz && \
    rm go.tar.gz

RUN curl -sSLo protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip -d /usr/local/protoc protoc.zip && \
    rm protoc.zip

RUN curl -sSLo node.tar.xz https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz && \
    tar -C /usr/local -Jxf node.tar.xz && \
    rm node.tar.xz && \
    mv /usr/local/node-v${NODE_VERSION}-linux-x64 /usr/local/node

ENV PATH="/usr/local/go/bin:/usr/local/protoc/bin:/usr/local/node/bin:${PATH}"

ENV GO111MODULE=on

RUN GOBIN=/usr/local/bin go get \
    github.com/bufbuild/buf/cmd/buf@v0.54.1 \
    github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking@v0.54.1 \
    github.com/bufbuild/buf/cmd/protoc-gen-buf-lint@v0.54.1 \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.5.0 \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.5.0 \
    google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1 \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0 \
    github.com/envoyproxy/protoc-gen-validate@v0.6.1 \
    moul.io/protoc-gen-gotemplate@v1.11.2 \
    github.com/google/addlicense@v1.0.0 \
    github.com/goreleaser/goreleaser@v0.177.0 \
    github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1 \
