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
    ./deploy.sh init        ----- 初始化节点配置
    ./deploy.sh restart     ----- 重启所有服务
    ./deploy.sh reload     	----- 重启 imonitor
    ./deploy.sh pushonline  ----- 部署 testnet 分支最新代码
    ./deploy.sh start  	    ----- 启动所有服务
    ./deploy.sh stop  		----- 停止所有服务并删除数据'
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
elif [ "$1" = "start" ]
then
    cmd="start.sh"
elif [ "$1" = "stop" ]
then
    cmd="stop.sh"
elif [ "$1" = "reset" ]
then
    cmd="reset.sh"

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
        cmd="init.sh&args=iserver_local_${node[2]}.yml"
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
        sleepTime=24
        echo -e "\033[34m     sleep ${sleepTime}s... \033[0m"
        sleep $sleepTime
    fi
done

