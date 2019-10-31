# NGE websocket tester

## 客户端

> - 支持指定参数连接上游服务，目前支持的系统有NGE & BitMEX
>
> - 支持公有流及私有流的订阅接收
>
> - 支持 SQL 对接收的流数据进行过滤显示
>
> - 默认订阅 orderBookL2，trade，instrument 三个公有流数据
>
> - 断线自动重连，并记录本次连接时长
>
> - Ctrl+C 中止程序运行，并显示程序启动时间、运行时长、断连次数及最长连接时间
>
> - 可使用 `-d`，`--deadline`  参数指定程序运行时长，超时自动退出，支持的时间单位：
>
>   > ```bash
>   > # 指定程序运行 5 小时后退出，支持的周期字段 h, m, s
>   > $ cd examples/client
>   > $ go run *.go --deadline 5h
> > ```
>   
>   1. h 小时
>   2. m 分钟
>   3. s 秒

### HELP

```bash
$ cd examples/client
$ go run *.go --help
Usage of /tmp/go-build247188547/b001/exe/filter:
      --append              Wether append topic list to default subscrib.
  -d, --deadline duration   Deadline duration, -1 means infinity. (default -1ns)
      --delay int           Delay seconds per binary expect backoff algorithm's delay slot. (default 3)
      --fail int            Heartbeat fail count. (default 3)
      --heartbeat int       Heartbeat interval in seconds. (default 15)
  -H, --host string         Host addreses to connect. (default "www.btcmex.com")
      --key string          API Key for authentication request.
      --max-count int       Max slot count in binary expect backoff algorithm. (default 6)
      --max-retry int       Max reconnect count, -1 means infinity. (default -1)
      --output string       SQL for output.
  -p, --port int            Host port to connect. (default 443)
      --scheme string       Websocket scheme. (default "wss")
      --secret string       API Secret for authentication request.
      --symbol string       Symbol name. (default "XBTUSD")
      --topics strings      Topic names for subscribe. (default [trade,orderBookL2,instrument])
      --uri string          URI for realtime push data. (default "/realtime")
      --url string          Connection URL.
  -v, --verbose count       Debug level, turn on for detail info.
pflag: help requested
exit status 2
```

### SQL FILTER EXAMPLE

> - SQL 字段名为 [ngerest](https://github.com/frozenpine/ngerest) 工程中定义的结构字段名，字段名不区分大小写
> - SQL 表名区分大小写，与公有流推送数据中 "table" 字段的大小写一致
> - [ngerest](https://github.com/frozenpine/ngerest) 为  [BitMEX/api-connectors](https://github.com/BitMEX/api-connectors/tree/master/auto-generated/go)  工程的一个 Fork，修正了部分 NGE 中同字段不同数据类型的兼容性

```bash
$ cd examples/client

# Example for trade
# 输出显示 trade 数据流中的指定字段
$ go run *.go --output 'SELECT symbol,price,side,size,tickDirection,timestamp FROM trade'


# Example for orderBookL2
# 输出显示 MBL 数据流中所有字段内容，同时将 orderBookL2 表名更改为 mbl
$ go run *.go --output 'SELECT * FROM orderBookL2 AS mbl'

# Example for instrument
# 输出显示 instrument 数据流中指数价格及标记价格大于0的记录
$ go run *.go --output 'SELECT indicativeSettlePrice, markPrice FROM instrument WHERE indicativeSettlePrice > 0 AND markPrice > 0'
```

## 服务端

> - 支持 orderBookL2、orderBookL2_25、instrument、trade 的公有流数据传输
>
> - 支持 trade 数据流的 Mock（暂时需通过调整代码实现，详见 server/server.go 中的 FIXME）
>
> - orderBook 及 instrument 数据流目前仅支持通过 Upstream 级联上级数据源
>
>   > 后续版本将引入 orderbook 模块用于支持模拟撮合，将实现所有公有流数据的mock

### HELP

目前未加入命令行参数的支持，程序默认监听 0.0.0.0:9988，支持两个 endpoint：

1. /realtime websocket入口点

2. /status 服务端简单的状态信息

   > ```bash
   > $ curl -s localhost:9988/status
   > {"startup":"2019-10-31T07:09:04.2256833Z","clients":2,"uptime":"3h47m35.6323026s"}
   > ```

### STARTUP EXAMPLE

```bash
$ cd examples/server
$ go run main.go
```

