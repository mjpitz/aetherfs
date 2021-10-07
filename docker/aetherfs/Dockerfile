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

FROM alpine:3.12

LABEL org.opencontainers.image.source="https://github.com/mjpitz/aetherfs"

WORKDIR /home/aetherfs
RUN apk add -U tzdata ca-certificates

ARG TARGETOS
ARG TARGETARCH
COPY dist/aetherfs_${TARGETOS}_${TARGETARCH}/aetherfs /usr/bin/aetherfs

ENTRYPOINT [ "/usr/bin/aetherfs" ]
