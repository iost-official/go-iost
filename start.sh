#!/bin/bash

redis-server /etc/redis/redis.conf

./iserver --config /var/lib/iserver/iserver.yml
