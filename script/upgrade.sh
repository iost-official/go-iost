#!/bin/sh
#
# upgrade.sh
# Author Yuanyi<yuanyi@iost.io>
#
# Distributed under terms of the LGPLv3 license.
#

set -ue

PREFIX=${PREFIX:="/data/iserver"}

CURL="curl -fsSL"
PYTHON=${PYTHON:=python}

USR_LOCAL_BIN=${USR_LOCAL_BIN:=/usr/local/bin}
export PATH=$PATH:$USR_LOCAL_BIN

#
# function
#

print_bye() {
    {
        echo
        echo Upgrade done. iServer info:
        set +e; docker inspect iserver | $PYTHON -c "import json,sys;d=json.load(sys.stdin)[0];print(d['Config']['Image']);print(d['Image'])"; set -e
        echo
        echo Happy hacking !
        echo
    }>&2
}

#
# main
#

cd $PREFIX

docker-compose pull
docker-compose up -d

until $($CURL localhost:30001/getNodeInfo &>/dev/null); do
    >&2 printf '.'
    sleep 2
done

print_bye
