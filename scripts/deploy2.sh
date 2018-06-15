#!/bin/bash

Port=30310

Tokyo="Tokyo 18.179.143.193 1 "$Port
London="London 52.56.118.10 2 "$Port
Singapore="Singapore 13.228.206.188 3 "$Port
Mumbai="Mumbai 13.232.96.221 4 "$Port
Frankfort="Frankfort 18.184.239.232 5 "$Port
Seoul="Seoul 13.124.172.86 6 "$Port
Montreal="Montreal 52.60.163.60 7 "$Port

AllNodes=("$Tokyo" "$London" "$Singapore" "$Mumbai" "$Frankfort" "$Seoul" "$Montreal")

echoHelp(){
    echo 'usage:
    ./deploy.sh redis        ----- 初始化节点配置'
}


if [ $# -lt 1 ]; then
    echoHelp
    exit 1
fi

if [ "$1" = "redis" ]
then
    cmd="checkredis.sh"
else
    echoHelp
    exit 1
fi


nodeCnt=${#AllNodes[*]}
echo "start $1 $nodeCnt servers:"

for ((i=0; i<$nodeCnt; i++))
do
    node=(${AllNodes[$i]})
    echo -e "  "${node[0]}" "${node[1]}":  \c"
    res=$(curl -s --connect-timeout 3  -XPOST ${node[1]}:${node[3]}/scripts -d "cmd=$cmd")
    echo -e "\033[31m     result=$res \033[0m"
done

