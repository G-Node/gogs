#!/bin/sh
set -x
set -e

# Set temp environment vars
export GOPATH=/tmp/go
export PATH=/usr/local/go/bin:${PATH}:${GOPATH}/bin

# Build Gogs
rm -rf ${GOPATH}/src/github.com/G-Node/gogs
mkdir -p ${GOPATH}/src/github.com/G-Node/
ln -s /app/gogs/build ${GOPATH}/src/github.com/G-Node/gogs
cd ${GOPATH}/src/github.com/G-Node/gogs
# Needed since git 2.9.3 or 2.9.4
git config --global http.https://gopkg.in.followRedirects true

touch *
make bindata
make build TAGS="sqlite cert pam"


# Cleanup GOPATH
#rm -r $GOPATH


# Create git user for Gogs

addgroup  git
adduser --home /data/git --shell /bin/sh --ingroup git --disabled-password git
passwd -d git
echo "export GOGS_CUSTOM=${GOGS_CUSTOM}" >> /etc/profile
