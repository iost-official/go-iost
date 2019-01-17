#!/bin/bash

readonly GRPC_URL="127.0.0.1:30002"
readonly VOTER_ACCOUNT="admin"
readonly VOTER_ACCOUNT_SECKEY="2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"

readonly WITNESS_NUM=1

readonly WITNESS_NAME=(
producer000
)

iwallet account --import ${VOTER_ACCOUNT} ${VOTER_ACCOUNT_SECKEY}
for (( i = 0; i < ${WITNESS_NUM}; i++ ))
do
    iwallet -s ${GRPC_URL} --account ${VOTER_ACCOUNT} call "vote_producer.iost" "vote" '["'${VOTER_ACCOUNT}'", "'${WITNESS_NAME[i]}'", "3000000"]'
done
