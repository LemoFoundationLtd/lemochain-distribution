![Logo of the project](./logo.png)

# LemoChain-Distribution


#### 配置文件
- 文件名：distribution-config.json，且与程序可执行文件放在同级目录下；
- 文件范例
```
{
	"chainID": 100,
	"genesisHash": "0x531d9cccfc39cdb1957a4eac21f0154eb6c192a76123ca786adebf54821d53b4",
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
		"listenAddress": "127.0.0.1"
	},
	"webSocket":{
		"disable": true,
		"port": 5005,
		"corsDomain": "*",
		"listenAddress": "127.0.0.1"
	}
}
```
其中：
- [ ] chainID: 与要连接的lemochain-core一致
- [ ] genesisHash： 与要连接的lemochain-core一致
- [ ] serverDataDir： 区块等相关数据存放目录
- [ ] dbUri： 数据库连接字符串
- [ ] dbDriver： 数据库类型
- [ ] logLevel： 日志输出级别
- [ ] coreNode： 被连接的lemochain-core相关NodeID与IP端口
- [ ] http、webSocket：rpc配置
- [ ] http.disable： 是否禁止http服务，默认开启
- [ ] http.port：http服务器端口
- [ ] http.corsDomain：http跨域允许列表
- [ ] http.virtualHosts：http虚拟主机
- [ ] http.listenAddress：http监听地址
- [ ] webSocket.disable：是否禁止websocket服务，默认开启
- [ ] webSocket.port；websocket服务器端口
- [ ] webSocket.corsDomain：websocket允许跨域域名列表
- [ ] webSocket.listenAddress：websocket监听地址
