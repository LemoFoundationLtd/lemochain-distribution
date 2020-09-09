package config

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/network/p2p"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ConfigGuideUrl          = ""
	NodeKeyFileName         = "nodekey"
	JsonFileName            = "distribution-config.json"
	DefaultHttpPort         = 8001
	DefaultHttpVirtualHosts = "localhost"
	DefaultWSPort           = 8002
)

var (
	ErrConfigFormat          = fmt.Errorf(`file "%s" format error. %s`, JsonFileName, ConfigGuideUrl)
	ErrChainIDInConfig       = fmt.Errorf(`file "%s" error: chainID must be in [1, 65535]`, JsonFileName)
	ErrLogLevelInConfig      = fmt.Errorf(`file "%s" error: logLevel must be in [1, 5]`, JsonFileName)
	ErrHttpPortInConfig      = fmt.Errorf(`file "%s" error: http port must be less than 65535`, JsonFileName)
	ErrWebSocketPortInConfig = fmt.Errorf(`file "%s" error: websocket port must be less than 65535`, JsonFileName)
	ErrCoreNodeInConfig      = fmt.Errorf(`file "%s" error: coreNode must be like: 5e3600755f9b512a65603b38e30885c98cbac70259c3235c9b3f42ee563b480edea351ba0ff5748a638fe0aeff5d845bf37a3b437831871b48fd32f33cd9a3c0@127.0.0.1:60001`, JsonFileName)
)

//go:generate gencodec -type RpcHttp -field-override RpcMarshaling -out gen_http_json.go
//go:generate gencodec -type RpcWS -field-override RpcMarshaling -out gen_ws_json.go
//go:generate gencodec -type Config -field-override ConfigMarshaling -out gen_config_json.go

type RpcHttp struct {
	Disable      bool   `json:"disable"`
	Port         uint32 `json:"port"  gencodec:"required"`
	CorsDomain   string `json:"corsDomain"`
	VirtualHosts string `json:"virtualHosts"`
}

type RpcMarshaling struct {
	Port hexutil.Uint32
}

type RpcWS struct {
	Disable    bool   `json:"disable"`
	Port       uint32 `json:"port"  gencodec:"required"`
	CorsDomain string `json:"corsDomain"`
}

type Config struct {
	ChainID         uint32  `json:"chainID"        gencodec:"required"`
	DeputyCount     uint32  `json:"deputyCount"    gencodec:"required"`
	TermDuration    uint64  `json:"termDuration"`
	InterimDuration uint64  `json:"interimDuration"`
	DbUri           string  `json:"dbUri"          gencodec:"required"` // sample: root:123123@tcp(localhost:3306)/lemochain?charset=utf8mb4
	DbDriver        string  `json:"dbDriver"       gencodec:"required"`
	LogLevel        uint32  `json:"logLevel"`
	CoreNode        string  `json:"coreNode"       gencodec:"required"`
	Http            RpcHttp `json:"http"`
	WebSocket       RpcWS   `json:"webSocket"`

	DataDir      string
	nodeKey      *ecdsa.PrivateKey
	coreNodeID   *p2p.NodeID
	coreEndpoint string
}

type ConfigMarshaling struct {
	ChainID         hexutil.Uint32
	DeputyCount     hexutil.Uint32
	TermDuration    hexutil.Uint64
	InterimDuration hexutil.Uint64
	LogLevel        hexutil.Uint32
}

func ReadConfigFile() (*Config, error) {
	// Try to read from command line
	dataDir := os.Args[1]
	filePath := filepath.Join(dataDir, JsonFileName)
	if _, err := os.Stat(filePath); err != nil {
		// Try to read from relative path
		filePath = JsonFileName
	}
	log.Infof("Load config file: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New(err.Error() + "\r\n" + ConfigGuideUrl)
	}

	var config Config
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return nil, ErrConfigFormat
	}
	config.DataDir = dataDir
	return &config, nil
}

func (c *Config) Check() {
	if c.ChainID > 65535 || c.ChainID < 1 {
		panic(ErrChainIDInConfig)
	}
	if c.DeputyCount == 0 {
		c.DeputyCount = 17
	}
	if c.TermDuration > 0 {
		params.TermDuration = uint32(c.TermDuration)
	}
	if c.InterimDuration > 0 {
		params.InterimDuration = uint32(c.InterimDuration)
	}
	if c.LogLevel == 0 {
		c.LogLevel = 4
	}
	if c.LogLevel > 5 {
		panic(ErrLogLevelInConfig)
	}
	if !c.Http.Disable {
		if c.Http.Port > 65535 {
			panic(ErrHttpPortInConfig)
		} else if c.Http.Port == 0 {
			c.Http.Port = DefaultHttpPort
		}

		if c.Http.VirtualHosts == "" {
			c.Http.VirtualHosts = DefaultHttpVirtualHosts
		}
	}
	if !c.WebSocket.Disable {
		if c.WebSocket.Port > 65535 {
			panic(ErrWebSocketPortInConfig)
		} else if c.WebSocket.Port == 0 {
			c.WebSocket.Port = DefaultWSPort
		}
	}
	nodeID, endpoint := parseNodeString(c.CoreNode)
	if nodeID == nil {
		panic(ErrCoreNodeInConfig)
	}
	c.coreNodeID = nodeID
	c.coreEndpoint = endpoint
}

func (c *Config) NodeKey() *ecdsa.PrivateKey {
	if c.nodeKey != nil {
		return c.nodeKey
	}

	keyFile := filepath.Join(c.DataDir, NodeKeyFileName)
	if key, err := crypto.LoadECDSA(keyFile); err == nil {
		c.nodeKey = key
		return key
	}

	key, err := crypto.GenerateKey()
	if err != nil {
		log.Critf("Failed to generate node key: %v", err)
	}
	instanceDir, _ := filepath.Abs(c.DataDir)
	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		log.Errorf("Failed to persist node key: %v", err)
		return key
	}
	keyFile = filepath.Join(instanceDir, NodeKeyFileName)
	if err := crypto.SaveECDSA(keyFile, key); err != nil {
		log.Errorf("Failed to persist node key: %v", err)
	}
	c.nodeKey = key
	return key
}

func (c *Config) CoreNodeID() *p2p.NodeID {
	return c.coreNodeID
}

func (c *Config) CoreEndpoint() string {
	return c.coreEndpoint
}

// parseNodeString verify node address
func parseNodeString(node string) (*p2p.NodeID, string) {
	tmp := strings.Split(node, "@")
	if len(tmp) != 2 {
		return nil, ""
	}
	var nodeID = tmp[0]
	var IP = tmp[1]

	// nodeID
	if len(nodeID) != 128 {
		return nil, ""
	}
	parsedNodeID := p2p.BytesToNodeID(common.FromHex(nodeID))
	_, err := parsedNodeID.PubKey()
	if err != nil {
		return nil, ""
	}
	if bytes.Compare(parsedNodeID[:], deputynode.GetSelfNodeID()) == 0 {
		return nil, ""
	}

	// IP
	if !verifyIP(IP) {
		return nil, ""
	}
	return parsedNodeID, IP
}

// verifyIP  verify ipv4
func verifyIP(input string) bool {
	tmp := strings.Split(input, ":")
	if len(tmp) != 2 {
		return false
	}
	if ip := net.ParseIP(tmp[0]); ip == nil {
		return false
	}
	p, err := strconv.Atoi(tmp[1])
	if err != nil {
		return false
	}
	if p < 1024 || p > 65535 {
		return false
	}
	return true
}
