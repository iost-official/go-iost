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
USR_LOCAL_BIN=${USR_LOCAL_BIN:=/usr/local/bin}
export PATH=$PATH:$USR_LOCAL_BIN

#
# function
#

print_bye() {
    {
        echo Upgrade done.
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
