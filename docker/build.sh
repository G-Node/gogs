#!/bin/sh
set -x
set -e

# Set temp environment vars
export GOPATH=/tmp/go
export PATH=/usr/local/go/bin:${PATH}:${GOPATH}/bin
export GO15VENDOREXPERIMENT=1

# Build Gogs
mkdir -p ${GOPATH}/src/github.com/gogits/
ln -s /app/gogs/build ${GOPATH}/src/github.com/gogits/gogs
cd ${GOPATH}/src/github.com/gogits/gogs
# Needed since git 2.9.3 or 2.9.4
git config --global http.https://gopkg.in.followRedirects true
make build TAGS="sqlite cert pam"

# Cleanup GOPATH
rm -r $GOPATH


# Create git user for Gogs

addgroup  git
adduser --home /data/git --shell /bin/sh --ingroup git --disabled-password git
passwd -d git
echo "export GOGS_CUSTOM=${GOGS_CUSTOM}" >> /etc/profile
