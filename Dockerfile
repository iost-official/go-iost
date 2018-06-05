FROM centos:7

ENV GOVERSION 1.10.1

## install git
RUN yum update -y && yum install wget git make gcc gcc-c++ kernel-devel redis -y
RUN git config --global user.name "IOST" && git config --global user.email "chain@iost.io"

# Install Redis.
RUN \
  cd /tmp && \
  wget http://download.redis.io/redis-stable.tar.gz && \
  tar xvzf redis-stable.tar.gz && \
  cd redis-stable && \
  make && \
  make install && \
  cp -f src/redis-sentinel /usr/local/bin && \
  mkdir -p /etc/redis && \
  cp -f *.conf /etc/redis && \
  rm -rf /tmp/redis-stable* && \
  sed -i 's/^\(bind .*\)$/# \1/' /etc/redis/redis.conf && \
  sed -i 's/^\(daemonize .*\)$/# \1/' /etc/redis/redis.conf && \
  sed -i 's/^\(logfile .*\)$/# \1/' /etc/redis/redis.conf

EXPOSE 6379

## install go
RUN mkdir /goroot && \
    mkdir /gopath && \
    curl https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz | \
         tar xzf - -C /goroot --strip-components=1

ENV CGO_ENABLED 1
ENV GOPATH /gopath
ENV GOROOT /goroot
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH

# Install project
RUN mkdir -p $GOPATH/src/github.com/iost-official && cd $GOPATH/src/github.com/iost-official && \
git clone https://445789ea93ff81d814c78fccae8e25000f96e539@github.com/iost-official/prototype && \
cd prototype && git checkout develop && go get github.com/kardianos/govendor && govendor sync -v && \
cd iserver && go build && cd ..


EXPOSE 30302
EXPOSE 30303

WORKDIR $GOPATH/src/github.com/iost-official/prototype/iserver

## docker deploy
## docker build -t iost .
## docker run --name iost_container -p 30302:30302 -p 30303:30303 -d iost ./start.sh

