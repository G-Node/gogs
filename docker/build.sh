#!/bin/sh
set -x
set -e

# Set temp environment vars
export GOPATH=/tmp/go
export PATH=/usr/local/go/bin:${PATH}:${GOPATH}/bin

# Install build deps
apk --no-cache --no-progress add --virtual build-deps build-base linux-pam-dev

# Build Gogs
mkdir -p ${GOPATH}/src/github.com/gogs/
ln -s /app/gogs/build ${GOPATH}/src/github.com/gogs/gogs
cd ${GOPATH}/src/github.com/gogs/gogs
# Needed since git 2.9.3 or 2.9.4
git config --global http.https://gopkg.in.followRedirects true
make build TAGS="sqlite cert pam"

# Cleanup GOPATH
rm -r $GOPATH

# Remove build deps
apk --no-progress del build-deps

# Move to final place
mv /app/gogs/build/gogs /app/gogs/

# Cleanup go
rm -rf /tmp/go
rm -rf /usr/local/go
