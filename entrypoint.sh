#!/bin/bash

set -e

function print_options {
  echo "Valid options:"
  echo "  Readiness/Liveness probe via file"
  echo "  -h <time>           : heartbeat file expiration"
  echo "  -f <heartbeat_file> : defaults to /dev/shm/heartbeat"
  echo
  echo "  -v                  : Be verbose"
}

function heart_beat {
  [ ! -f "${HEARTBEAT_FILE}" ] && { echo "${HEARTBEAT_FILE} does not exist."; exit 1; }
  if [ $(( $(date +%s) - $(date -r $HEARTBEAT_FILE +%s) )) -gt $HEARTBEAT_TIMEOUT ];then
    echo "No heartbeat in more than ${HEARTBEAT_TIMEOUT} seconds"
    exit 2
  fi
  exit 0
}
HEARTBEAT_FILE="/dev/shm/heartbeat"

while getopts "vh:f:" flag; do
  case "$flag" in
    f)
      HEARTBEAT_FILE=$OPTARG
    ;;
    h)
      HEARTBEAT_TIMEOUT=$OPTARG
    ;;
    v) VERBOSE="-verbose";;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      print_options
      exit 1
    ;;
  esac
done
[ ! -z $HEARTBEAT_TIMEOUT ] && heart_beat

shift $((OPTIND-1))

CMD="${@}"

if [ -z "${CMD}" ];then
  echo "Usage: $0 [options] <cmd>"
  print_options
  exit 1
fi

function exec_command {
  if [ "${CMD}" = "dummy" ];then
    while true; do
      echo -e "HTTP/1.1 200 OK\n\n $(date)" | nc -lNn 0.0.0.0 8080
    done
    exit 0
  elif [ -f "Startupfile" ];then
    while IFS= read line
    do
      local label=${line%%:*}
      local cmd=${line#*:}
      if [ "${label}" = "${CMD}" ];then
        echo "----> Executing: LABEL: ${label} CMD: ${cmd}"
        eval "$cmd"
        exit $?
      fi
    done <"Startupfile"
  fi
  echo "----> Executing: ${CMD}"
  eval "${CMD}"
  EXIT_CODE=$?
  echo "Done!"
  exit $EXIT_CODE
}

eval "$(k8senv -e ${VERBOSE})"
exec_command
