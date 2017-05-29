#!/bin/sh
set -x
set -e

# Set temp environment vars
export GOPATH=/tmp/go
export PATH=/usr/local/go/bin:${PATH}:${GOPATH}/bin

# Build Gogs
mkdir -p ${GOPATH}/src/github.com/gogs/
ln -s /app/gogs/build ${GOPATH}/src/github.com/gogs/gogs
cd ${GOPATH}/src/github.com/gogs/gogs
# Needed since git 2.9.3 or 2.9.4
git config --global http.https://gopkg.in.followRedirects true
make build TAGS="sqlite cert pam"

# Cleanup GOPATH
rm -r $GOPATH


# Create git user for Gogs
adduser git --home /data/git --shell /bin/bash
echo "export GOGS_CUSTOM=${GOGS_CUSTOM}" >> /etc/profile
