FROM registry.suse.com/suse/sle15:latest AS build-base

RUN set -ex \
    && zypper update \
    && zypper in -y make git gcc bzip2 rpmbuild wget curl tar gzip


FROM build-base AS go-base

ARG GO_VERSION=""
ENV GO_VERSION_ENV="$GO_VERSION"
ENV GOCACHE=/tmp

RUN set -ex \
    && if [ -z "${GO_VERSION_ENV}" ]; then export GO_VERSION_ENV="$(curl https://golang.org/VERSION?m=text)"; fi \
    && wget "https://dl.google.com/go/$GO_VERSION_ENV.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzf $GO_VERSION_ENV.linux-amd64.tar.gz

ENV PATH="$PATH:/usr/local/go/bin"

WORKDIR /build
