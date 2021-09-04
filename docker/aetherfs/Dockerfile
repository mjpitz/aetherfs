FROM alpine:3.12

LABEL org.opencontainers.image.source="https://github.com/mjpitz/aetherfs"

WORKDIR /home/aetherfs
RUN apk add -U tzdata ca-certificates

ARG TARGETOS
ARG TARGETARCH
COPY dist/aetherfs_${TARGETOS}_${TARGETARCH}/aetherfs /usr/bin/aetherfs

ENTRYPOINT [ "/usr/bin/aetherfs" ]
