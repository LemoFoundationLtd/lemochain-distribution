package main

import (
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	coreNode "github.com/LemoFoundationLtd/lemochain-core/main/node"
	"github.com/LemoFoundationLtd/lemochain-distribution/main/config"
	"github.com/LemoFoundationLtd/lemochain-distribution/main/node"
	"os"
	"os/signal"
	"syscall"
)

var stopCh = make(chan struct{})

func main() {
	cfg, err := config.ReadConfigFile()
	if err != nil {
		panic(fmt.Sprintf("config file read error: %v", err))
	}
	coreNode.InitLogConfig(int(cfg.LogLevel))
	if err := startServer(cfg); err != nil {
		panic(fmt.Sprintf("start server failed: %v", err))
	}
}

// startServer
func startServer(cfg *config.Config) error {
	n, err := node.New(cfg)
	if err != nil {
		return fmt.Errorf("new node failed: %v", err)
	}
	if err = n.Start(); err != nil {
		return fmt.Errorf("start node failed: %v", err)
	}
	go interrupt(n.Stop)
	log.Infof("start server ok...")
	<-stopCh
	return nil
}

// interrupt listen interrupt
func interrupt(wait func() error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	<-sigCh
	log.Info("Got interrupt, shutting down...")
	go wait()
	stopCh <- struct{}{}
	for i := 5; i > 0; i-- {
		<-sigCh
		if i > 1 {
			log.Warnf("Already shutting down, interrupt more to panic. times: %d", i-1)
		}
	}
	panic("boom")
}
