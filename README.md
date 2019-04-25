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
instructions:
- `chainID` must be the same as lemochain-core.
- `deputyCount` must be the same as lemochain-core.
- `serverDataDir` store randomly generated nodekey.
- `dbUri` database uri.
- `dbDriver` database type.
- `logLevel` log output level.
- `coreNode` will connect the lemochain-core.
- `http and webSocket` rpc config.
- `http.disable` whether to turn on HTTP, default on.
- `http.port` http port
- `http.corsDomain` cross-domain allow list, all domain access is allowed if it is '*'.
- `http.virtualHosts` precheck request permission list, all domain access is allowed if it is '*'.
- `http.listenAddress` http listen address.
- `webSocket.disable` whether to turn on webSocket, default on.
- `webSocket.port` websocket port.
- `webSocket.corsDomain` the same as http.
- `webSocket.listenAddress` websocket listen address.

#### start
- Please click on the [wiki](https://github.com/LemoFoundationLtd/lemochain-distribution/wiki).