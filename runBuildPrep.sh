#!/bin/bash -x

zypper --plus-repo=http://dst.us.cray.com/dstrepo/bloblets/os/dev/mirrors/rpms/sles/15sp2/ --no-gpg-checks in -y --auto-agree-with-licenses go1.14