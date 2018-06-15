#! /bin/bash

redis-cli GET BlockNum
redis-cli GET BlockHash
redis-cli HGETALL iost
