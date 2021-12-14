# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

FROM ubuntu:21.04

ARG GO_VERSION=1.17.2
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

RUN npm install -g prettier

ENV GO111MODULE=on
ENV GOBIN=/usr/local/bin
COPY godeps.txt godeps.txt
RUN cat godeps.txt | xargs -I{} go install {}
