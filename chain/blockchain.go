package chain

import (
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	coreNet "github.com/LemoFoundationLtd/lemochain-core/network"
	"github.com/LemoFoundationLtd/lemochain-core/store"
	"github.com/LemoFoundationLtd/lemochain-distribution/database"
	"sync"
	"sync/atomic"
)

type BlockChain struct {
	chainID      uint16
	dm           *deputynode.Manager
	stableBlock  atomic.Value // latest stable block in current chain
	genesisBlock *types.Block // genesis block

	chainForksHead map[common.Hash]*types.Block // total latest header of different fork chain
	chainForksLock sync.Mutex
	mux            sync.Mutex
	running        int32
	dbEngine       database.DBEngine
}

// blockLoader BlockLoader implement for deputynode.Manager
type blockLoader struct {
	bc *BlockChain
}

func (b *blockLoader) GetBlockByHeight(height uint32) (*types.Block, error) {
	block := b.bc.GetBlockByHeight(height)
	if block == nil {
		return nil, store.ErrNotExist
	} else {
		return block, nil
	}
}

func NewBlockChain(chainID uint16, deputyCount int, dbEngine database.DBEngine) (bc *BlockChain, err error) {
	bc = &BlockChain{
		chainID:        chainID,
		chainForksHead: make(map[common.Hash]*types.Block, 16),
		dbEngine:       dbEngine,
	}

	if err := bc.loadGenesis(); err != nil {
		return nil, err
	}

	if err := bc.loadLastState(); err != nil {
		return nil, err
	}

	bc.dm = deputynode.NewManager(deputyCount, &blockLoader{bc})

	return bc, nil
}

// func (bc *BlockChain) AccountManager() *account.Manager {
// 	return bc.am
// }

func (bc *BlockChain) DeputyManager() *deputynode.Manager {
	return bc.dm
}

func (bc *BlockChain) loadGenesis() error {
	blockDao := database.NewBlockDao(bc.dbEngine)
	block, err := blockDao.GetBlockByHeight(0)
	if err == database.ErrNotExist {
		return nil
	}

	if err != nil {
		log.Errorf("get genesis err: " + err.Error())
		return err
	} else {
		bc.genesisBlock = block
		return nil
	}
}

// loadLastState load latest state in starting
func (bc *BlockChain) loadLastState() error {
	contextDao := database.NewContextDao(bc.dbEngine)
	block, err := contextDao.GetCurrentBlock()
	if err == database.ErrNotExist {
		return nil
	} else if err != nil {
		log.Errorf("Can't load last state: %v", err)
		return err
	}
	bc.stableBlock.Store(block)
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
	blockDao := database.NewBlockDao(bc.dbEngine)
	if ok, _ := blockDao.IsExist(hash); ok {
		return true
	}
	return false
}

func (bc *BlockChain) getGenesisFromDb() *types.Block {
	blockDao := database.NewBlockDao(bc.dbEngine)
	block, err := blockDao.GetBlockByHeight(0)
	if err != nil && err == database.ErrNotExist {
		return nil
	} else if err != nil && err != database.ErrNotExist {
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
	stable := bc.StableBlock()
	if stable == nil { // 还未初始化到genesis block
		return nil
	}
	stableBlockHeight := stable.Height()
	if stableBlockHeight >= height {
		blockDao := database.NewBlockDao(bc.dbEngine)
		block, err := blockDao.GetBlockByHeight(height)
		if err != nil {
			panic(fmt.Sprintf("can't get block. height:%d, err: %v", height, err))
		}
		return block
	}
	return nil
}

func (bc *BlockChain) GetBlockByHash(hash common.Hash) *types.Block {
	blockDao := database.NewBlockDao(bc.dbEngine)
	block, err := blockDao.GetBlock(hash)
	if err != nil {
		log.Debugf("can't get block. hash:%s", hash.Hex())
		return nil
	}
	return block
}

// CurrentBlock get latest current block
func (bc *BlockChain) CurrentBlock() *types.Block {
	return bc.StableBlock()
}

// StableBlock get latest stable block
func (bc *BlockChain) StableBlock() *types.Block {
	if bc.stableBlock.Load() == nil {
		return nil
	}
	return bc.stableBlock.Load().(*types.Block)
}

// updateDeputyNodes update deputy nodes map
func (bc *BlockChain) updateDeputyNodes(block *types.Block) {
	if block.Height()%params.TermDuration == 0 {
		bc.dm.SaveSnapshot(block.Height(), block.DeputyNodes)
		log.Debugf("save new term deputy nodes: %v", block.DeputyNodes)
	}
}

func (bc *BlockChain) InsertBlock(block *types.Block) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	hash := block.Hash()
	blockDao := database.NewBlockDao(bc.dbEngine)
	has, err := blockDao.IsExist(hash)
	if err != nil || has {
		return err
	}

	reBuildEngine := NewReBuildEngine(bc.dbEngine, block)
	err = reBuildEngine.ReBuild()
	if err != nil {
		return err
	} else {
		bc.updateDeputyNodes(block)
		bc.stableBlock.Store(block)
		if block.Height() == 0 {
			bc.genesisBlock = block
		}

		log.Debugf("insert block success. Height:%d", block.Height())
		return nil
	}
}

// not used. just for implement interface
func (bc *BlockChain) InsertConfirm(info *coreNet.BlockConfirmData) {
}
func (bc *BlockChain) InsertStableConfirms(pack coreNet.BlockConfirms) {
}
func (bc *BlockChain) IsInBlackList(b *types.Block) bool {
	return false
}
