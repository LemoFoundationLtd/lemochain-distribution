![Logo of the project](./logo.png)

# LemoChain-Distribution


[中文版](https://github.com/LemoFoundationLtd/lemochain-distribution/blob/master/README_zh.md)   
[English](https://github.com/LemoFoundationLtd/lemochain-distribution/blob/master/README.md)


#### configuration file
- File name：`distribution-config.json`,and must be put at the same level with the program executable directory;
- example
```
{
	"chainID": 100,
	"dbUri": "root:123456@tcp(127.0.0.1:3306)/lemo02?charset=utf8mb4",
	"dbDriver": "mysql",
	"logLevel": 5,
	"deputyCount": 17,
	"coreNode":"5e3600755f9b512a65603b38e30885c98cbac70259c3235c9b3f42ee563b480edea351ba0ff5748a638fe0aeff5d845bf37a3b437831871b48fd32f33cd9a3c0@120.78.132.151:7003",
	"http":{
		"disable": true,
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
instructions:
- `chainID` The ID of LemoChain.
- `dbUri` Database uri.
- `dbDriver` Database type.
- `logLevel` Log output level.
- `deputyCount` The max number of consensus nodes.
- `coreNode` Address of the lemochain-core to connect. It's looks like `nodeId@IP:Port`.
- `http and webSocket` RPC config.
- `http.disable` Whether to turn off HTTP, default on.
- `http.port` Http port
- `http.corsDomain` Cross-domain from whitelist, all domain accesses are allowed if it is '*'.
- `http.virtualHosts` Cross-domain to whitelist, all domain accesses are allowed if it is '*'.
- `webSocket.disable` Whether to turn off webSocket, default on.
- `webSocket.port` Websocket port.
- `webSocket.corsDomain` The same as http.

#### start
- Please click on the [wiki](https://github.com/LemoFoundationLtd/lemochain-distribution/wiki).