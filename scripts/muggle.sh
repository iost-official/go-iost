#!/bin/bash

Port1=30309
Port2=30308

Tokyo1="Tokyo 18.179.143.193 1 "$Port1
London1="London 52.56.118.10 2 "$Port1
Singapore1="Singapore 13.228.206.188 3 "$Port1
Mumbai1="Mumbai 13.232.96.221 4 "$Port1
Frankfort1="Frankfort 18.184.239.232 5 "$Port1
Seoul1="Seoul 13.124.172.86 6 "$Port1
Montreal1="Montreal 52.60.163.60 7 "$Port1
Tokyo2="Tokyo 18.179.143.193 8 "$Port2
London2="London 52.56.118.10 9 "$Port2
Singapore2="Singapore 13.228.206.188 10 "$Port2
Mumbai2="Mumbai 13.232.96.221 11 "$Port2
Frankfort2="Frankfort 18.184.239.232 12 "$Port2
Seoul2="Seoul 13.124.172.86 13 "$Port2
Montreal2="Montreal 52.60.163.60 14 "$Port2

AllNodes=("$Tokyo1" "$London1" "$Singapore1" "$Mumbai1" "$Frankfort1" "$Seoul1" "$Montreal1" "$Tokyo2" "$London2" "$Singapore2" "$Mumbai2" "$Frankfort2" "$Seoul2" "$Montreal2")

echoHelp(){
    echo 'usage:
    ./deploy.sh init        ----- 初始化节点配置
    ./deploy.sh restart     ----- 重启所有服务
    ./deploy.sh reload      ----- 重启 imonitor
    ./deploy.sh stop        ----- 关闭 iserver
    ./deploy.sh pushonline  ----- 部署 testnet 分支最新代码'
}


if [ $# -lt 1 ]; then
    echoHelp
    exit 1
fi

if [ "$1" = "restart" ]
then
    cmd="restart-iserver"
elif [ "$1" = "pushonline" ]
then
    cmd="upgrade.sh"
elif [ "$1" = "reload" ]
then
    cmd="reload"
elif [ "$1" = "init" ]
then
    cmd="init.sh"
elif [ "$1" = "stop" ]
then
        cmd="stop-iserver"
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
    if [ "$1" = "init" ]
    then
        cmd="init.sh&args=muggle/iserver_${node[2]}.yml"
    fi
    res=$(curl -s --connect-timeout 3  -XPOST ${node[1]}:${node[3]}/scripts -d "cmd=$cmd")
    if [ "$res" = "ok" -o "$res" = "SUCCESS" ]
    then
        echo -e "\033[32m ok \033[0m"
    else
        echo -e "\033[31m fail \033[0m"
        echo -e "\033[31m     result=$res \033[0m"
    fi
    if [ "$1" = "restart" ]
    then
        sleepTime=1
        echo -e "\033[34m     sleep ${sleepTime}s... \033[0m"
        sleep $sleepTime
    fi
done

