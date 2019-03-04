package chain

import (
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/common/subscribe"
	"github.com/LemoFoundationLtd/lemochain-server/common/log"
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
	subscribe.Send(subscribe.NewTx, tx)
	return nil
}

func (tp *TxPool) validateTx(tx *types.Transaction) error {
	// todo
	return nil
}
