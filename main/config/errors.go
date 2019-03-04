package config

import "errors"

var (
	ErrConfig      = errors.New("config format error")
	ErrChainId     = errors.New("chainID is invalid")
	ErrGenesisHash = errors.New("genesis hash is invalid")
	ErrLogLevel    = errors.New("logLevel is invalid")
	ErrPort        = errors.New("port is invalid")
	ErrCoreNode    = errors.New("coreNode is invalid")
)
