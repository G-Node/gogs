#!/bin/sh

create_socat_links() {
    # Bind linked docker container to localhost socket using socat
    USED_PORT="3000:22"
    while read NAME ADDR PORT; do
        if test -z "$NAME$ADDR$PORT"; then
            continue
        elif echo $USED_PORT | grep -E "(^|:)$PORT($|:)" > /dev/null; then
            echo "init:socat  | Can't bind linked container ${NAME} to localhost, port ${PORT} already in use" 1>&2
        else
            SERV_FOLDER=/app/gogs/docker/s6/SOCAT_${NAME}_${PORT}
            mkdir -p ${SERV_FOLDER}
            CMD="socat -ls TCP4-LISTEN:${PORT},fork,reuseaddr TCP4:${ADDR}:${PORT}"
            echo -e "#!/bin/sh\nexec $CMD" > ${SERV_FOLDER}/run
            chmod +x ${SERV_FOLDER}/run
            USED_PORT="${USED_PORT}:${PORT}"
            echo "init:socat  | Linked container ${NAME} will be binded to localhost on port ${PORT}" 1>&2
        fi
    done << EOT
    $(env | sed -En 's|(.*)_PORT_([0-9]+)_TCP=tcp://(.*):([0-9]+)|\1 \3 \4|p')
EOT
}

cleanup() {
    # Cleanup SOCAT services and s6 event folder
    # On start and on shutdown in case container has been killed
    rm -rf $(find /app/gogs/docker/s6/ -name 'event')
    rm -rf /app/gogs/docker/s6/SOCAT_*
}

create_volume_subfolder() {
    # Create VOLUME subfolder
    for f in /data/gogs/data /data/gogs/conf /data/gogs/log /data/git /data/ssh; do
        if ! test -d $f; then
            mkdir -p $f
        fi
    done
}

setids() {
    PUID=${PUID:-1000}
    PGID=${PGID:-1000}
    groupmod -o -g "$PGID" git
    usermod -o -u "$PUID" git
}

setids
cleanup
create_volume_subfolder

LINK=$(echo "$SOCAT_LINK" | tr '[:upper:]' '[:lower:]')
if [ "$LINK" = "false" -o "$LINK" = "0" ]; then
    echo "init:socat  | Will not try to create socat links as requested" 1>&2
else
    create_socat_links
fi

CROND=$(echo "$RUN_CROND" | tr '[:upper:]' '[:lower:]')
if [ "$CROND" = "true" -o "$CROND" = "1" ]; then
    echo "init:crond  | Cron Daemon (crond) will be run as requested by s6" 1>&2
    rm -f /app/gogs/docker/s6/crond/down
else
    # Tell s6 not to run the crond service
    touch /app/gogs/docker/s6/crond/down
fi

#

#SSH setup
# Check if host keys are present, else create them
if ! test -f /data/ssh/ssh_host_rsa_key; then
    ssh-keygen -q -f /data/ssh/ssh_host_rsa_key -N '' -t rsa
fi

if ! test -f /data/ssh/ssh_host_dsa_key; then
    ssh-keygen -q -f /data/ssh/ssh_host_dsa_key -N '' -t dsa
fi

if ! test -f /data/ssh/ssh_host_ecdsa_key; then
    ssh-keygen -q -f /data/ssh/ssh_host_ecdsa_key -N '' -t ecdsa
fi

if ! test -f /data/ssh/ssh_host_ed25519_key; then
    ssh-keygen -q -f /data/ssh/ssh_host_ed25519_key -N '' -t ed25519
fi

if ! test -d ~git/.ssh; then
    mkdir -p ~git/.ssh
    chmod 700 ~git/.ssh
fi

#Gogs setup
if ! test -f ~git/.ssh/environment; then
    echo "GOGS_CUSTOM=${GOGS_CUSTOM}" > ~git/.ssh/environment
    chmod 600 ~git/.ssh/environment
fi

cd /app/gogs

# Link volumed data with app data
ln -sf /data/gogs/log  ./log
ln -sf /data/gogs/data ./data

# Backward Compatibility with Gogs Container v0.6.15
ln -sf /data/git /home/git

chown -R git:git /data /app/gogs ~git/
chmod 0755 /data /data/gogs ~git/

# Set correct right to ssh keys
chown -R root:root /data/ssh/*
chmod 0700 /data/ssh
chmod 0600 /data/ssh/*
# Exec CMD or S6 by default if nothing present
if [ $# -gt 0 ];then
    exec "$@"
else
    exec supervisord -c /app/gogs/docker/supervisord.conf
fi
