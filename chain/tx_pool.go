package chain

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
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
	if err := tp.validateTx(tx); err != nil {
		log.Error("transaction is invalid: %v", err)
		return err
	}
	subscribe.Send(network.GetNewTx, tx)
	return nil
}

func (tp *TxPool) validateTx(tx *types.Transaction) error {
	// todo
	return nil
}
