#!/usr/bin/env bash

set -e

NAME="clickhouse-goose"

FORCE=0
VERSION=""

for i in "$@"
do
    case ${i} in
        -f|--force)
        FORCE=1
        shift
        ;;
        *)
        ;;
    esac
done

function get_version() {
    V=$(grep 'const version' cmd/${NAME}/main.go | awk '{print $4}')
    VERSION="${V%\"}"
    VERSION="${VERSION#\"}"

    if ! [[ ${VERSION} =~ [0-9]+\.[0-9]+\.[0-9] ]]; then
     echo "Invalid version from cmd/${NAME}/main.go"
     exit
    fi
}

function already_built() {
    OS=$1
    if [[ -f ./releases/${OS}/${VERSION}/${NAME} ]]; then
        echo "1"
    else
        echo "0"
    fi
}


function build() {
    OS=$1

    built=$(already_built ${OS})

    if [[ "$built" == "1" ]] && [[ ${FORCE} -eq 0 ]]; then
        echo "$OS: binary already in ./releases/${OS}/${VERSION}/${NAME}. Use -f to rebuild"
        return
    fi

    mkdir -p ./releases/${OS}/${VERSION}
    docker build -t ${NAME}-${OS}:${VERSION} -f ${OS}.dockerfile .
    docker run -d --name ${NAME}-${OS} ${NAME}-${OS}:${VERSION}
    docker cp ${NAME}-${OS}:/build/cmd/${NAME}/${NAME} ./releases/${OS}/${VERSION}/
    docker stop ${NAME}-${OS}
    docker rm ${NAME}-${OS}
    echo "$OS: binary built and placed in ./releases/${OS}/${VERSION}/${NAME}"
}

get_version
build alpine
build debian
build ubuntu