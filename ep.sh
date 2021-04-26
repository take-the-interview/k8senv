#!/bin/bash

set -e

function print_options {
    echo "Valid options:"
    echo "Sourcing runtime environment file:"
    echo "  -f <source_file_path> : path to .env file"
}

function source_env {
    [ ! -f "${SOURCE_FILE_PATH}" ] && { echo "no source file path provided"; exit 1; }
    source $SOURCE_FILE_PATH
}
SOURCE_FILE_PATH=/etc/runtime.env

while getopts ":f:" flag; do
    case ${flag} in
        f)
            SOURCE_FILE_PATH=$OPTARG
        ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            exit 1
        ;;
    esac
done

echo -e "sourcing ${SOURCE_FILE_PATH}"
[ -f $SOURCE_FILE_PATH ] && source_env

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


exec_command
