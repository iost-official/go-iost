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

update_container_restart_policy() {
    # IB-613 fix container restart policy
    sed -i.bak 's/on-failure/unless-stopped/' $PREFIX/docker-compose.yml
    rm -f $PREFIX/docker-compose.yml.bak
}

update_container_ulimit() {
    # IB-638 increase ulimit
    grep -q nofile $PREFIX/docker-compose.yml && return
    _SYS=$(uname)
    if [ x$_SYS = x"Linux" ]; then
        sed -i -e '/volumes/{n;a\    ulimits:\n      nofile: 51200' -e '}' $PREFIX/docker-compose.yml
    elif [ x$_SYS = x"Darwin" ]; then
        sed -i '' -e '/volumes/{n;a\
            \    ulimits:' -e ';a\
            \      nofile: 51200' -e '}' $PREFIX/docker-compose.yml
    else
        >&2 echo System not recognized !
    fi
}

update_container_security_opt() {
  grep -q seccomp $PREFIX/docker-compose.yml && return
  uu="Ubuntu"
  sys=$(lsb_release -a 2> /dev/null | grep "Distributor ID:" | cut -d ":" -f2)
  result=$(echo $sys | grep "${uu}")
  if [ "$result" != ""  ]; then
      ver=$(lsb_release -r --short)
      if [ `echo "$ver < 22"|bc` -eq 1 ]; then
          sed -i -e '/ulimits/{n;a\    security_opt:\n      - seccomp:unconfined' -e '}' $PREFIX/docker-compose.yml
      fi
  fi
}

#
# main
#

cd $PREFIX

update_container_restart_policy
update_container_ulimit
update_container_security_opt
docker-compose pull
docker-compose up -d

until $($CURL localhost:30001/getNodeInfo &>/dev/null); do
    >&2 printf '.'
    sleep 2
done

print_bye
