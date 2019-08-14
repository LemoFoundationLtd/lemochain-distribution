package node

import (
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/chain/account"
	"github.com/LemoFoundationLtd/lemochain-core/common/flock"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	coreNode "github.com/LemoFoundationLtd/lemochain-core/main/node"
	"github.com/LemoFoundationLtd/lemochain-core/network/rpc"
	"github.com/LemoFoundationLtd/lemochain-core/store"
	"github.com/LemoFoundationLtd/lemochain-core/store/protocol"
	"github.com/LemoFoundationLtd/lemochain-distribution/chain"
	"github.com/LemoFoundationLtd/lemochain-distribution/database"
	"github.com/LemoFoundationLtd/lemochain-distribution/main/config"
	. "github.com/LemoFoundationLtd/lemochain-distribution/network"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Node struct {
	config *config.Config

	db     protocol.ChainDB
	accMan *account.Manager
	chain  *chain.BlockChain
	pm     *ProtocolManager

	txPool *chain.TxPool

	instanceDirLock flock.Releaser

	rpcAPIs []rpc.API

	httpEndpoint  string
	httpWhitelist []string
	httpListener  net.Listener
	httpHandler   *rpc.Server

	wsEndpoint string
	wsListener net.Listener
	wsHandler  *rpc.Server
}

func initDb(dataDir string, driver string, dns string) protocol.ChainDB {
	dir := filepath.Join(dataDir, "chaindata")
	return store.NewChainDataBase(dir, driver, dns)
}

func New(cfg *config.Config) (*Node, error) {
	bc, err := chain.NewBlockChain(uint16(cfg.ChainID), int(cfg.DeputyCount), database.NewMySqlDB(cfg.DbDriver, cfg.DbUri))
	if err != nil {
		return nil, err
	}
	pm := NewProtocolManager(uint16(cfg.ChainID), cfg.CoreNodeID(), cfg.CoreEndpoint(), bc)

	n := &Node{
		config: cfg,
		chain:  bc,
		// accMan: bc.AccountManager(),
		pm:     pm,
		txPool: chain.NewTxPool(),
	}

	return n, nil
}

func (n *Node) Start() error {
	if err := n.openDataDir(); err != nil {
		log.Errorf("%v", err)
		return coreNode.ErrOpenFileFailed
	}
	n.pm.Start()
	if err := n.startRPC(); err != nil {
		log.Errorf("%v", err)
		return coreNode.ErrRpcStartFailed
	}
	return nil
}

func (n *Node) Stop() error {
	n.stopRPC()
	if err := n.accMan.Stop(true); err != nil {
		log.Errorf("stop account manager failed: %v", err)
		return err
	}
	log.Debug("stop account manager ok...")
	if n.instanceDirLock != nil {
		if err := n.instanceDirLock.Release(); err != nil {
			log.Errorf("Can't release datadir lock: %v", err)
		}
		n.instanceDirLock = nil
	}
	return nil
}

func (n *Node) openDataDir() error {
	if n.config.DataDir == "" {
		return nil
	}
	if err := os.MkdirAll(n.config.DataDir, 0700); err != nil {
		return err
	}
	release, _, err := flock.New(filepath.Join(n.config.DataDir, "LOCK"))
	if err != nil {
		return err
	}
	n.instanceDirLock = release
	return nil
}

func (n *Node) startRPC() error {
	apis := n.apis()
	if err := n.startHttp(apis); err != nil {
		return err
	}
	if err := n.startWS(apis); err != nil {
		n.stopHttp()
		return err
	}
	n.rpcAPIs = apis
	return nil
}

func (n *Node) startHttp(apis []rpc.API) error {
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()
	for _, api := range apis {
		if api.Public {
			if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
				return err
			}
		}
	}
	// All APIs registered, start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	endpoint := fmt.Sprintf("%s:%d", n.config.Http.ListenAddress, n.config.Http.Port)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		return err
	}
	cors := strings.Split(n.config.Http.CorsDomain, ",")
	vhosts := strings.Split(n.config.Http.VirtualHosts, ",")
	go rpc.NewHTTPServer(cors, vhosts, handler).Serve(listener)
	log.Info("HTTP endpoint opened", "url", fmt.Sprintf("http://%s", endpoint), "cors", strings.Join(cors, ","), "vhosts", strings.Join(vhosts, ","))
	// All listeners booted successfully
	n.httpEndpoint = endpoint
	n.httpListener = listener
	n.httpHandler = handler

	return nil
}

func (n *Node) startWS(apis []rpc.API) error {
	// todo
	return nil
}

func (n *Node) stopRPC() {
	n.stopHttp()
	n.stopWS()
}

func (n *Node) stopHttp() {
	if n.httpListener != nil {
		if err := n.httpListener.Close(); err != nil {
			log.Errorf("close httpListener failed: %v", err)
		}
		n.httpListener = nil

		log.Info("HTTP endpoint closed", "url", fmt.Sprintf("http://%s", n.httpEndpoint))
	}
	if n.httpHandler != nil {
		n.httpHandler.Stop()
		n.httpHandler = nil
	}
}

func (n *Node) stopWS() {
	// todo
}

func (n *Node) apis() []rpc.API {
	return []rpc.API{
		{
			Namespace: "chain",
			Version:   "1.0",
			Service:   NewPublicChainAPI(n),
			Public:    true,
		},
		{
			Namespace: "account",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(n),
			Public:    true,
		},
		{
			Namespace: "account",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(n),
			Public:    false,
		},
		{
			Namespace: "tx",
			Version:   "1.0",
			Service:   NewPublicTxAPI(n),
			Public:    true,
		},
	}
}
