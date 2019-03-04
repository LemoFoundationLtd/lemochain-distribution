package main

import (
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-server/common/log"
	"github.com/LemoFoundationLtd/lemochain-server/main/config"
	"github.com/LemoFoundationLtd/lemochain-server/main/node"
	"github.com/inconshreveable/log15"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var stopCh chan struct{}

func main() {
	dir := filepath.Dir(os.Args[0])
	cfg, err := config.ReadConfigFile(dir)
	if err != nil {
		panic(fmt.Sprintf("config file read error: %v", err))
	}
	initLog(int(cfg.LogLevel))
	if err := startServer(cfg); err != nil {
		panic(fmt.Sprintf("start server failed: %v", err))
	}
}

// initLog init log config
func initLog(logFlag int) {
	// flag in command is in range 1~5
	// logLevel is in range 0~4
	logLevel := log15.Lvl(logFlag)
	// default level
	if logLevel < 0 || logLevel > 4 {
		logLevel = log.LevelError // 1
	}
	showCodeLine := logLevel >= 3 // LevelInfo, LevelDebug
	log.Setup(logLevel, true, showCodeLine)
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
	for i := 5; i > 0; i-- {
		<-sigCh
		if i > 1 {
			log.Warnf("Already shutting down, interrupt more to panic. times: %d", i-1)
		}
	}
	stopCh <- struct{}{}
	panic("boom")
}
