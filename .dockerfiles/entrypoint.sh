#!/bin/sh

random_string() {
  tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w 64 | head -n 1
}

[ -n "${PUID}" ] && usermod -u "${PUID}" todo
[ -n "${PGID}" ] && groupmod -g "${PGID}" todo

printf "Configuring todo...\n"
[ -z "${DATA}" ] && DATA="/data"
export DATA

printf "Switching UID=%s and GID=%s\n" "${PUID}" "${PGID}"
exec su-exec todo:todo "$@"
