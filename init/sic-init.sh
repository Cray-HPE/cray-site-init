#!/bin/bash

set -x
if [ $# -lt 2 ]; then
    echo >&2 "usage: sic-init PIDFILE CIDFILE [CONTAINER [VOLUME]]"
    exit 1
fi

SIC_PIDFILE="$1"
SIC_CIDFILE="$2"
SIC_CONTAINER_NAME="${3-sic}"
SIC_CONTAINER_IMAGE="dtr.dev.cray.com/metal/cloud-${SIC_CONTAINER_NAME}"
SIC_VOLUME_MOUNT_CONFIG="/var/www/metal/sic/configs:/app/configs:rw,noexec"

command -v podman >/dev/null 2>&1 || { echo >&2 "${0##*/}: command not found: podman"; exit 1; }

# always ensure pid file is fresh
rm -f "$SIC_PIDFILE"
mkdir -pv "$(echo ${SIC_VOLUME_MOUNT_CONFIG} | cut -f 1 -d :)"

# Create sic container
if ! podman inspect "$SIC_CONTAINER_NAME" ; then
    rm -f "$SIC_CIDFILE" || exit
    podman pull "$SIC_CONTAINER_IMAGE" || exit
    podman create \
        --conmon-pidfile "$SIC_PIDFILE" \
        --cidfile "$SIC_CIDFILE" \
        --cgroups=no-conmon \
        -d \
        --net host \
        --name "$SIC_CONTAINER_NAME" \
        --env GIN_MODE="${GIN_MODOE:-release}" \
        "$SIC_CONTAINER_IMAGE" || exit
    podman inspect "$SIC_CONTAINER_NAME" || exit
fi
