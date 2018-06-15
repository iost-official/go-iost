#! /bin/bash

redis-cli GET BlockNumber
redis-cli GET BlockHash
redis-cli HGETALL iost
