# Copyright (C) The AetherFS Authors - All Rights Reserved
#
# Proprietary and confidential.
# Unauthorized copying of this file, via any medium is strictly prohibited.
# Written by Mya Pitzeruse, September 2021

FROM alpine:3.12

LABEL org.opencontainers.image.source="https://github.com/mjpitz/aetherfs"

WORKDIR /home/aetherfs
RUN apk add -U tzdata ca-certificates

ARG TARGETOS
ARG TARGETARCH
COPY dist/aetherfs_${TARGETOS}_${TARGETARCH}/aetherfs /usr/bin/aetherfs

ENTRYPOINT [ "/usr/bin/aetherfs" ]
