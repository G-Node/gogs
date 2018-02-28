#!/bin/sh
set -x
set -e

# Set temp environment vars
export GOPATH=/tmp/go
export PATH=/usr/local/go/bin:${PATH}:${GOPATH}/bin

#
go get golang.org/x/crypto/bcrypt
go get github.com/jteeuwen/go-bindata
go get github.com/G-Node/gin-doi/src
go get github.com/G-Node/gin-dex/gindex
go get github.com/G-Node/git-module
go get gopkg.in/yaml.v2

cd ${GOPATH}/src/github.com/jteeuwen/go-bindata/go-bindata
go install

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
