#!/bin/bash

readonly RPC_URL="http://127.0.0.1:30001"
readonly VOTE_ACCOUNT="admin"
readonly VOTE_ACCOUNT_SECKEY="2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"
readonly REPLACE_WITNESS="producer01"

readonly WITNESS_NUM=33

readonly WITNESS_NAME=(

)

function err
{
    echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $@" >&2
}

function init_state
{
    iwallet account --import ${VOTE_ACCOUNT} ${VOTE_ACCOUNT_SECKEY}

    iwallet --account ${VOTE_ACCOUNT} call "vote_producer.iost" "unvote" '["'${VOTE_ACCOUNT}'", "'${REPLACE_WITNESS}'", "3000000"]'
}

function get_now_height
{
    return iwallet state | grep '"headBlock"' | tr -cd "[0-9]"
}

function get_witness_by_height
{
    return curl ${RPC_URL}/getBlockByNumber/$1/true | grep -Eo '"witness":".*?"' | cut -d: -f2 | tr -cd "[0-9,a-z,A-Z]"
}

function get_pubkey_by_witness
{
    return curl -X POST ${RPC_URL}/getContractStorage -d '{"id":"vote_producer.iost","key":"producerTable","field":"'$1'","by_longest_chain":true}' | grep -oE 'pubkey\\":\\".*?"' | cut -d: -f2 | tr -cd "[0-9,a-z,A-Z]"
}

function gen_block_success
{
    start = get_now_height
    end = start + 200

    for (( i = start; i < end; i++ ))
    do
        if [[ $(get_witness_by_height $i) == $1 ]]
        then
            return true
        if
    done
    return false
}

function check_witness
{
    for i in {0..WITNESS_NUM-1}
    do
        iwallet --account ${VOTE_ACCOUNT} call "vote_producer.iost" "vote" '["'${VOTE_ACCOUNT}'", "'${WITNESS_NAME[i]}'", "300000000"]'

        witness_pubkey = get_pubkey_by_witness ${WITNESS_NAME[i]}

        if [[ gen_block_success $witness_pubkey ]]
        then
            echo "${WITNESS_NAME[i]} gen block successful!"
        else
            err "${WITNESS_NAME[i]} gen block failed!"
        fi
        iwallet --account ${VOTE_ACCOUNT} call "vote_producer.iost" "unvote" '["'${VOTE_ACCOUNT}'", "'${WITNESS_NAME[i]}'", "300000000"]'
    done
}

function reset_state
{
    iwallet --account ${VOTE_ACCOUNT} call "vote_producer.iost" "vote" '["'${VOTE_ACCOUNT}'", "'${REPLACE_WITNESS}'", "3000000"]'
}
