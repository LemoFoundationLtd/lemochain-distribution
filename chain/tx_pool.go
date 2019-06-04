package chain

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/subscribe"
	"github.com/LemoFoundationLtd/lemochain-distribution/network"
)

// TxPool add filter in future
type TxPool struct {
	// todo
}

func NewTxPool() *TxPool {
	return &TxPool{}
}

func (tp *TxPool) AddTx(tx *types.Transaction) error {
	go subscribe.Send(network.GetNewTx, tx)
	return nil
}
