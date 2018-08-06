#!/bin/bash

set -e

function print_options {
  echo "Valid options:"
  echo "  -v                : Be verbose"
}

while getopts "v" flag; do
  case "$flag" in
    v) VERBOSE="-verbose";;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      print_options
      exit 1
    ;;
  esac
done
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
      echo -e "HTTP/1.1 200 OK\n\n $(date)" | nc -l 0.0.0.0 8080 
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
