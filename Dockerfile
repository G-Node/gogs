FROM ubuntu:16.04

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update &&                                   \
    apt-get install -y --no-install-recommends          \
                       gcc g++ libc6-dev make golang    \
                       git git-annex openssh-server     \
                       python-pip python-setuptools     \
                       socat tzdata patch    \
                       libpam0g-dev node-less \
    && rm -rf /var/lib/apt/lists/*

RUN pip install supervisor pyyaml


ENV GOGS_CUSTOM /data/gogs

COPY . /app/gogs/build
WORKDIR /app/gogs/build

RUN ./docker/build-go.sh
RUN ./docker/build.sh
RUN ./docker/finalize.sh

# Configure LibC Name Service
COPY docker/nsswitch.conf /etc/nsswitch.conf

# Configure Docker Container
VOLUME ["/data"]
EXPOSE 22 3000
ENTRYPOINT ["/app/gogs/docker/start.sh"]
