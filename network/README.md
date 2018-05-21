# network

## 网络层：
    Router      网络接口，通过注册信道队列(FilteredChan)的形式对接上层应用

    BaseNetwork 网络层具体实现，负责节点发现，建立连接，消息的发送、接收等管理动作

    Request     定义了网络层的数据结构，数据打包，解包，各种数据类型的处理

    Node        定义一个网络节点，邻居节点计算(xor距离最近的节点为邻居节点)，网络层的广播是从本节点向其邻居节点发送广播消息

    Peer        管理所有与本节点建立的连接



## BaseNetwork

    nodeTable 用以节点发现后，记录新节点，目前采用中心化的方式，后期再升级至gossip，全网的路由表，node启动时会将本机node注册到bootNode，再定时从bootNode中同步到本地

    peers 所有与node建立的连接，或者node发起的连接

    recentSentMap 管理所有已广播的数据，防止环形广播

    NodeHeightMap 记录其他节点广播来的区块链的高度，为下载区块任务提供候选的下载地址

    DownloadHeights 管理下载任务重试次数，超过最大重试次数 停止下载

    localNode 本机节点



## Request 网络层消息结构

    Version   消息版本
    Length    消息长度
    Timestamp 发送时间(纳秒)
    Type      网络消息类型
    FromLen   发送节点长度
    From      发送节点
    Body      消息内容


    Pack() UnPack() 消息打包 解包
    handle()        处理接收到的网络消息
    msgHandle()     处理区块高度消息收发



