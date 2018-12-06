FROM golang:alpine AS binarybuilder
# Install build deps
RUN apk --no-cache --no-progress add --virtual build-deps build-base git linux-pam-dev python py-pip
WORKDIR /go/src/github.com/G-Node/gogs
COPY . .
RUN make build TAGS="sqlite cert pam"

# RUN apt-get update &&                                   \
#     apt-get install -y --no-install-recommends          \
#                        gcc g++ libc6-dev make golang    \
#                        git git-annex openssh-server     \
#                        python-pip python-setuptools     \
#                        socat tzdata patch    \
#                        libpam0g-dev node-less \
#     && rm -rf /var/lib/apt/lists/*



FROM alpine:latest
# Install system utils & Gogs runtime dependencies
ADD https://github.com/tianon/gosu/releases/download/1.10/gosu-amd64 /usr/sbin/gosu
RUN chmod +x /usr/sbin/gosu \
  && echo http://dl-2.alpinelinux.org/alpine/edge/community/ >> /etc/apk/repositories \
  && apk --no-cache --no-progress add \
    bash \
    ca-certificates \
    curl \
    git \
    linux-pam \
    openssh \
    s6 \
    shadow \
    socat \
    tzdata \
    python \
    py-pip

RUN pip install supervisor pyyaml

ENV GOGS_CUSTOM /data/gogs

#COPY . /app/gogs/build
#WORKDIR /app/gogs/build

#RUN ./docker/build-go.sh
#RUN ./docker/build.sh
#RUN ./docker/finalize.sh

# Configure LibC Name Service
COPY docker/nsswitch.conf /etc/nsswitch.conf

WORKDIR /app/gogs
COPY docker ./docker
COPY templates ./templates
COPY public ./public
COPY --from=binarybuilder /go/src/github.com/G-Node/gogs/gogs .

RUN ./docker/finalize.sh

# Configure Docker Container
VOLUME ["/data"]
#VOLUME ["/tmp"]
EXPOSE 22 3000
ENTRYPOINT ["/app/gogs/docker/start.sh"]
CMD ["/bin/s6-svscan", "/app/gogs/docker/s6/"]
