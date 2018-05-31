# IOST prototype

## install vendor
govendor sync -v

## docker deploy

docker build -t iost .

docker run --name iost_container -p 30302:30302 -p 30303:30303 -d iost ./start.sh




2018年6月要完成TestNet！


