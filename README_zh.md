![Logo of the project](./logo.png)

# LemoChain-Distribution


[中文版](https://github.com/LemoFoundationLtd/lemochain-distribution/blob/master/README_zh.md)   
[English](https://github.com/LemoFoundationLtd/lemochain-distribution/blob/master/README.md)



#### 配置文件
- 文件名：distribution-config.json，且与程序可执行文件放在同级目录下；
- 文件范例
```
{
	"chainID": 100,
	"deputyCount": 17,
	"serverDataDir": "./lemo-distribution",
	"dbUri": "root:123456@tcp(127.0.0.1:3306)/lemo02?charset=utf8mb4",
	"dbDriver": "mysql",
	"logLevel": 5,
	"coreNode":"5e3600755f9b512a65603b38e30885c98cbac70259c3235c9b3f42ee563b480edea351ba0ff5748a638fe0aeff5d845bf37a3b437831871b48fd32f33cd9a3c0@120.78.132.151:7003",
	"http":{
		"disable": true,
		"port": 5001,
		"corsDomain": "*",
		"virtualHosts": "www.lemochain.com",
		"listenAddress": "0.0.0.0"
	},
	"webSocket":{
		"disable": true,
		"port": 5005,
		"corsDomain": "*",
		"listenAddress": "0.0.0.0"
	}
}
```
其中：
- `chainID` 与要连接的lemochain-core一致
- `deputyCount` 与要连接的lemochain-core一致
- `serverDataDir` 存放随机生成的nodekey
- `dbUri` 数据库连接字符串
- `dbDriver` 数据库类型
- `logLevel` 日志输出级别
- `coreNode` 被连接的lemochain-core相关NodeID与IP端口，配置格式为`nodeId@IP:Port`
- `http、webSocket` rpc配置
- `http.disable` 是否禁止http服务，默认开启
- `http.port` http服务器端口
- `http.corsDomain` http跨域允许列表,配置为"*"表示允许所有域名访问。
- `http.virtualHosts` http跨域限制预检请求允许列表，配置为"*"表示允许所有域名访问。
- `http.listenAddress` http监听地址
- `webSocket.disable` 是否禁止websocket服务，默认开启
- `webSocket.port` websocket服务器端口
- `webSocket.corsDomain` websocket允许跨域域名列表
- `webSocket.listenAddress` websocket监听地址

#### 启动流程
- 启动流程请转到[wiki](https://github.com/LemoFoundationLtd/lemochain-distribution/wiki).
