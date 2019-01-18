#!/bin/bash

readonly HTTP_URL="http://127.0.0.1:30001"

readonly WITNESS_NUM=1

readonly WITNESS_NAME=(
producer000
)

for (( i = 0; i < ${WITNESS_NUM}; i++ ))
do
    echo "$i: $(curl -s -X POST ${HTTP_URL}/getContractStorage -d '{"id":"vote.iost","owner":"","field":"'${WITNESS_NAME[i]}'","key":"v_1"}')"
done
