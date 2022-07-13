#! /bin/bash
# MIT License
#
# (C) Copyright [2021] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
set -ex
SNYK_OPTS="--dev --show-vulnerable-paths=all --fail-on=all --severity-threshold=${SEVERITY:-high} --skip-unresolved=true --json"

OUT=$(set -x; snyk test --all-projects --detection-depth=999 $SNYK_OPTS)

PROJ_CHECK=OK
jq .[].ok <<<"$OUT" | grep -q false && PROJ_CHECK=FAIL

echo Snyk project check: $PROJ_CHECK

DOCKER_CHECK=
if [ -f Dockerfile ]; then
    DOCKER_IMAGE=${PWD/*\//}:$(cat .version)
    docker build --tag $DOCKER_IMAGE .
    OUT=$(set -x; snyk test --docker $DOCKER_IMAGE --file=${PWD}/Dockerfile $SNYK_OPTS)
    DOCKER_CHECK=OK
    jq .ok <<<"$OUT" | grep -q false && DOCKER_CHECK=FAIL
fi

echo
echo Snyk project check: $PROJ_CHECK
echo Snyk docker check: $DOCKER_CHECK

test "$PROJ_CHECK" == OK -a "$DOCKER_CHECK" == OK
exit $?
