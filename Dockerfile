# SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
#
# SPDX-License-Identifier: CC0-1.0

FROM golang:alpine AS build
RUN mkdir /source
COPY . /source/
WORKDIR /source
RUN set -ex && \
    apk --no-cache add ca-certificates build-base make git && \
    go mod download && \
    make && \
    strip bin/image-manager && \
    mkdir /build && \
    cp bin/image-manager /build/

FROM alpine
COPY --from=build /build/ostree-upload /usr/bin/ostree-upload
RUN apk --no-cache add libc6-compat ostree
ENTRYPOINT ["/usr/bin/image-manager"]
CMD ["--help"]
