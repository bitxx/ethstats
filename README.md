# eth-stats
可用于实时监控分布自不同区域的ethereum节点，操作简单，一目了然。  
该项目可监控基于ethereum的大多数项目，包括L2、L3等等，具体细节，就需要自行探索了

## 功能
1. 每个节点名称不得重复
2. 支持实时上传节点信息
3. 节点异常时，实时邮件反馈。若同一个节点频繁出异常，则一个小时内只发送一份邮件，避免频繁发送造成邮箱上限异常
4. 定时邮件发送节点简报
5. server和client强稳定性，运行期间，除非强制杀进程或者bug，否则程序不会因为任何逻辑问题停止运行，降低了运维复杂度
6. 可通过命令行传入参或者通过配置文件启动`client、server`，不建议同时使用两种方式，选择其中一种即可
7. 本项目没有前端页面，主要是不会用前端语言，也设计不了。。。本项目在server/app/service/api中提供了socket数据出口，只要前端使用socket调用，即可渲染在前端。
   1. 前端通过socket的emit可读取：`stats 节点信息`、`latency 延迟`、`node-ping ping`三类数据

## 使用方式
分为客户端和服务器端，客户端安装在每台需要监控的节点上，服务器端找台有ip的稳定机子部署就行。
需要先在根目录执行：
```shell
go mod tidy
```

### client
```shell
cd client
go build -o client ./client.go 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o client client.go

# 配置方式启动
client start -c config/setting.yml

# 命令行方式启动
./client start --name test --secret 123456 --server-url 链地址，如：ws://127.0.0.1:30303
```

### server
```shell
cd server
go build server.go -o server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o server server.go

# 配置方式启动
server start -c config/setting.yml

# 命令行方式启动
./server start --name ethereum-server --secret 123456 --host 0.0.0.0 --port 3000 --email-subject-prefix ethereum --email-host 邮箱服务地址 --email-port 465 --email-username 发件邮箱账户 --email-password 邮箱密钥 --email-from 发件邮箱账户--email-to 收件邮箱账户(多个逗号隔开)
```

## 参考
[1] [goerli-ethstats-server](https://github.com/goerli/ethstats-server)  
[2] [goerli-ethstats-client](https://github.com/goerli/ethstats-client)  
[3] [AvileneRausch2001-ethstats](https://github.com/AvileneRausch2001/ethstats)  
[4] [AvileneRausch2001-ethstats](https://github.com/AvileneRausch2001/ethstats)  
[5] [maticnetwork-ethstats-backend](https://github.com/maticnetwork/ethstats-backend)  


