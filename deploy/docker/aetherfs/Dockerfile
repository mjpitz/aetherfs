# Copyright (C) The AetherFS Authors - All Rights Reserved
# See LICENSE for more information.

FROM alpine:3.14

WORKDIR /opt/aetherfs
RUN apk add -U tzdata ca-certificates

ARG TARGETOS
ARG TARGETARCH
COPY dist/aetherfs_${TARGETOS}_${TARGETARCH}/aetherfs /opt/aetherfs/bin/aetherfs
COPY LICENSE /opt/aetherfs/LICENSE

ENTRYPOINT [ "/opt/aetherfs/bin/aetherfs" ]
