# AetherFS - A virtual file system for small to medium sized datasets (MB or GB, not TB or PB).
# Copyright (C) The AetherFS Authors - All Rights Reserved
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

FROM ubuntu:21.04

ARG GO_VERSION=1.16.4
ARG PROTOC_VERSION=3.15.7

RUN apt-get update -y && apt-get install -y curl unzip gnupg2 git build-essential

RUN curl -sSLo go.tar.gz https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -zxf go.tar.gz && \
    rm go.tar.gz

RUN curl -sSLo protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip -d /usr/local/protoc protoc.zip && \
    rm protoc.zip

ENV PATH="/usr/local/go/bin:/usr/local/protoc/bin:${PATH}"

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
    github.com/goreleaser/goreleaser@v0.177.0
