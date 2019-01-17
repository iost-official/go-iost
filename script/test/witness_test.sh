#!/bin/bash

readonly HTTP_URL="http://127.0.0.1:30001"
readonly GRPC_URL="127.0.0.1:30002"
readonly VOTE_ACCOUNT="admin"
readonly VOTE_ACCOUNT_SECKEY="2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"
readonly REPLACE_WITNESS="producer000"

readonly WITNESS_NUM=0

readonly WITNESS_NAME=(
)

function err
{
    echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $@" >&2
}

function init_state
{
    iwallet -s ${GRPC_URL} account --import ${VOTE_ACCOUNT} ${VOTE_ACCOUNT_SECKEY}
    iwallet -s ${GRPC_URL} --account ${VOTE_ACCOUNT} call "vote_producer.iost" "unvote" '["'${VOTE_ACCOUNT}'", "'${REPLACE_WITNESS}'", "3000000"]'
}

function get_now_height
{
    echo $(iwallet -s ${GRPC_URL} state | grep '"headBlock"' | tr -cd "[0-9]")
}

function get_witness_by_height
{
    echo $(curl -s ${HTTP_URL}/getBlockByNumber/$1/true | grep -Eo '"witness":".*?"' | cut -d: -f2 | tr -cd "[0-9,a-z,A-Z]")
}

function get_pubkey_by_witness
{
    echo $(curl -s -X POST ${HTTP_URL}/getContractStorage -d '{"id":"vote_producer.iost","key":"producerTable","field":"'$1'","by_longest_chain":true}' | grep -oE 'pubkey\\":\\".*?"' | cut -d: -f2 | tr -cd "[0-9,a-z,A-Z]")
}

function gen_block_success
{
    local i
    start=$(get_now_height)
    end=$((${start} + 2200))

    for (( i = start; i < end; i++ ))
    do
        witness=$(get_witness_by_height $i)
        echo "check block from $start to $end, wait for witness $1, now height is $i, witness is $witness"
        if [[ $witness == $1 ]]
        then
            return 0
        fi
    done
    return 1
}

function check_witness
{
    local i
    for (( i = 0; i < ${WITNESS_NUM}; i++ ))
    do
        iwallet -s ${GRPC_URL} --account ${VOTE_ACCOUNT} call "vote_producer.iost" "vote" '["'${VOTE_ACCOUNT}'", "'${WITNESS_NAME[i]}'", "3000000"]'

        witness_pubkey=$(get_pubkey_by_witness ${WITNESS_NAME[i]})
        echo "${WITNESS_NAME[i]}'s pubkey is $witness_pubkey"

        if
        gen_block_success $witness_pubkey
        then
            err "${WITNESS_NAME[i]} gen block successful!"
        else
            err "${WITNESS_NAME[i]} gen block failed!"
        fi
        iwallet -s ${GRPC_URL} --account ${VOTE_ACCOUNT} call "vote_producer.iost" "unvote" '["'${VOTE_ACCOUNT}'", "'${WITNESS_NAME[i]}'", "3000000"]'
    done
}

function reset_state
{
    iwallet -s ${GRPC_URL} --account ${VOTE_ACCOUNT} call "vote_producer.iost" "vote" '["'${VOTE_ACCOUNT}'", "'${REPLACE_WITNESS}'", "3000000"]'
}

init_state
check_witness
reset_state
