# IOST prototype

## install vendor
govendor sync -v

## 启动本地测试网络
cd network/main/main && go run main.go --mode private [public,committee]

cd iserver && go run main.go --config iserver_local.yml



2018年6月要完成TestNet！


