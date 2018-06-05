#!/bin/bash

redis-server /etc/redis/redis.conf

cd $GOPATH/src/github.com/iost-official/prototype/iserver && ./iserver

