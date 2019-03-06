package chain

import (
	"fmt"
	coreChain "github.com/LemoFoundationLtd/lemochain-go/chain"
	"github.com/LemoFoundationLtd/lemochain-go/chain/account"
	"github.com/LemoFoundationLtd/lemochain-go/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-go/chain/params"
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	coreNet "github.com/LemoFoundationLtd/lemochain-go/network"
	"github.com/LemoFoundationLtd/lemochain-go/store"
	db "github.com/LemoFoundationLtd/lemochain-go/store/protocol"
	"github.com/LemoFoundationLtd/lemochain-server/common/log"
	"math/big"
	"sync"
	"sync/atomic"
)

type BlockChain struct {
	chainID      uint16
	db           db.ChainDB
	am           *account.Manager
	currentBlock atomic.Value // latest block in current chain
	stableBlock  atomic.Value // latest stable block in current chain
	genesisBlock *types.Block // genesis block

	chainForksHead map[common.Hash]*types.Block // total latest header of different fork chain
	chainForksLock sync.Mutex
	mux            sync.Mutex
	running        int32
}

func NewBlockChain(chainID uint16, db db.ChainDB) (bc *BlockChain, err error) {
	bc = &BlockChain{
		chainID:        chainID,
		db:             db,
		chainForksHead: make(map[common.Hash]*types.Block, 16),
	}
	if err := bc.loadLastState(); err != nil {
		return nil, err
	}
	return bc, nil
}

func (bc *BlockChain) AccountManager() *account.Manager {
	return bc.am
}

// Lock call by miner
func (bc *BlockChain) Lock() *sync.Mutex {
	return &bc.mux
}

// loadLastState load latest state in starting
func (bc *BlockChain) loadLastState() error {
	block, err := bc.db.LoadLatestBlock()
	if err == store.ErrNotExist {
		return nil
	} else if err != nil {
		log.Errorf("Can't load last state: %v", err)
		return err
	}
	bc.currentBlock.Store(block)
	bc.stableBlock.Store(block)
	bc.am = account.NewManager(block.Hash(), bc.db)
	return nil
}

// ChainID
func (bc *BlockChain) ChainID() uint16 {
	return bc.chainID
}

// Genesis genesis block
func (bc *BlockChain) Genesis() *types.Block {
	return bc.genesisBlock
}

// HasBlock has special block in local
func (bc *BlockChain) HasBlock(hash common.Hash) bool {
	if ok, _ := bc.db.IsExistByHash(hash); ok {
		return true
	}
	return false
}

func (bc *BlockChain) getGenesisFromDb() *types.Block {
	block, err := bc.db.GetBlockByHeight(0)
	if err != nil {
		panic("can't get genesis block")
	}
	return block
}

func (bc *BlockChain) GetBlockByHeight(height uint32) *types.Block {
	// genesis block
	if height == 0 {
		return bc.getGenesisFromDb()
	}

	// not genesis block
	block := bc.currentBlock.Load().(*types.Block)
	currentBlockHeight := block.Height()
	stableBlockHeight := bc.stableBlock.Load().(*types.Block).Height()
	var err error
	if stableBlockHeight >= height {
		block, err = bc.db.GetBlockByHeight(height)
		if err != nil {
			panic(fmt.Sprintf("can't get block. height:%d, err: %v", height, err))
		}
	} else if height <= currentBlockHeight {
		for i := currentBlockHeight - height; i > 0; i-- {
			block, err = bc.db.GetBlockByHash(block.ParentHash())
			if err != nil {
				panic(fmt.Sprintf("can't get block. height:%d, err: %v", height, err))
			}
		}
	} else {
		return nil
	}
	return block
}

func (bc *BlockChain) GetBlockByHash(hash common.Hash) *types.Block {
	block, err := bc.db.GetBlockByHash(hash)
	if err != nil {
		log.Debugf("can't get block. hash:%s", hash.Hex())
		return nil
	}
	return block
}

// CurrentBlock get latest current block
func (bc *BlockChain) CurrentBlock() *types.Block {
	return bc.currentBlock.Load().(*types.Block)
}

// StableBlock get latest stable block
func (bc *BlockChain) StableBlock() *types.Block {
	return bc.stableBlock.Load().(*types.Block)
}

// updateDeputyNodes update deputy nodes map
func (bc *BlockChain) updateDeputyNodes(block *types.Block) {
	if block.Height()%params.TermDuration == 0 {
		deputynode.Instance().Add(block.Height()+params.InterimDuration+1, block.DeputyNodes)
		log.Debugf("add new term deputy nodes: %v", block.DeputyNodes)
	}
}

// InsertChain insert block of non-self to chain
func (bc *BlockChain) InsertChain(block *types.Block, isSynchronising bool) (err error) {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	hash := block.Hash()
	if has, _ := bc.db.IsExistByHash(hash); has {
		return nil
	}
	if err = bc.db.SetBlock(hash, block); err != nil {
		log.Errorf("can't insert block to cache. height:%d hash:%s", block.Height(), hash.Prefix())
		return coreChain.ErrSaveBlock
	}
	if block.Height() == 0 {
		bc.initGenesis(block)
	}
	log.Infof("Insert block to chain. height: %d. hash: %s. time: %d. parent: %s", block.Height(), block.Hash().Prefix(), block.Time(), block.ParentHash().Prefix())
	// process changelog
	if err := bc.am.RebuildAll(block); err != nil {
		log.Errorf("rebuild account manager failed: %v", err)
		return err
	}
	if err := bc.am.Save(hash); err != nil {
		log.Errorf("save account manager failed: %v", err)
		return err
	}
	// update deputy nodes
	bc.updateDeputyNodes(block)
	bc.currentBlock.Store(block)
	bc.stableBlock.Store(block)
	return nil
}

func (bc *BlockChain) initGenesis(b *types.Block) {
	bc.am = account.NewManager(common.Hash{}, bc.db)
	total, _ := new(big.Int).SetString("1600000000000000000000000000", 10) // 1.6 billion
	bc.am.GetAccount(b.MinerAddress()).SetBalance(total)
}

func (bc *BlockChain) Db() db.ChainDB {
	return bc.db
}

// not used. just for implement interface
func (bc *BlockChain) SetStableBlock(hash common.Hash, height uint32) error {
	return nil
}
func (bc *BlockChain) Verify(block *types.Block) error {
	return nil
}
func (bc *BlockChain) ReceiveConfirm(info *coreNet.BlockConfirmData) error {
	return nil
}
func (bc *BlockChain) GetConfirms(query *coreNet.GetConfirmInfo) []types.SignData {
	return nil
}
func (bc *BlockChain) ReceiveConfirms(pack coreNet.BlockConfirms) {

}
