#!/bin/bash

redis-server /etc/redis/redis.conf

exec ./iserver --config /var/lib/iserver/iserver.yml $@
