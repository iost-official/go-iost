#!/bin/bash

nohup redis-server >> /var/lib/iserver/redis.log 2>&1 &
#redis-server /etc/redis/redis.conf
exec ./iserver --config /var/lib/iserver/iserver.yml $@
