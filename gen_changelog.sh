#!/bin/bash

# Run this script like so: ./genchange_log.sh <current_tag>

# default is previous tag, treating tags as version numbers.  Ignore pre-release tags.
PREV_RELEASE=${1:-"$(git tag -l --sort=-v:refname | grep -v "\-" | head -1)"}

THIS_VER=${2:-"HEAD"}

echo "Differences since ${PREV_RELEASE}..."

echo "$THIS_VER"
echo "-----------------------"

git log --no-merges "${PREV_RELEASE}..${THIS_VER}" --oneline --no-decorate | \
    awk '{print "- " substr($0, index($0, $2))}'
