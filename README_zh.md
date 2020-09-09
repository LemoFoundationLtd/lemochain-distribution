![Logo of the project](./logo.png)

# LemoChain-Distribution


[中文版](https://github.com/LemoFoundationLtd/lemochain-distribution/blob/master/README_zh.md)   
[English](https://github.com/LemoFoundationLtd/lemochain-distribution/blob/master/README.md)


#### 运行
运行时可以带一个参数，指定数据目录的位置
```shell script
lemo-distribution ./lemoserver-data
```

#### 配置文件
- 文件名：distribution-config.json，放在数据目录下
- 文件范例
```
{
	"chainID": 100,
	"dbUri": "root:123456@tcp(127.0.0.1:3306)/lemo02?charset=utf8mb4",
	"dbDriver": "mysql",
	"logLevel": 5,
	"deputyCount": 17,
	"coreNode":"5e3600755f9b512a65603b38e30885c98cbac70259c3235c9b3f42ee563b480edea351ba0ff5748a638fe0aeff5d845bf37a3b437831871b48fd32f33cd9a3c0@120.78.132.151:7003",
	"http":{
		"disable": false,
		"port": 5001,
		"corsDomain": "*",
		"virtualHosts": "www.lemochain.com"
	},
	"webSocket":{
		"disable": true,
		"port": 5005,
		"corsDomain": "*"
	}
}
```
其中：
- `chainID` LemoChain的ID
- `dbUri` 数据库连接字符串
- `dbDriver` 数据库类型
- `logLevel` 日志输出级别
- `deputyCount` 区块链的最大共识节点数
- `coreNode` 要连接的lemochain-core节点地址，格式为`nodeId@IP:Port`
- `http、webSocket` rpc配置
- `http.disable` 是否禁止http服务，默认开启
- `http.port` http服务器端口
- `http.corsDomain` http跨域允许列表，"*"表示允许所有域名访问
- `http.virtualHosts` 允许部署的域名列表，"*"表示允许部署在任意域名
- `webSocket.disable` 是否禁止websocket服务，默认开启
- `webSocket.port` websocket服务器端口
- `webSocket.corsDomain` websocket允许跨域域名列表，"*"表示允许所有域名访问

#### 启动流程
- 启动流程请转到[wiki](https://github.com/LemoFoundationLtd/lemochain-distribution/wiki).
