package chain

import (
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	coreNet "github.com/LemoFoundationLtd/lemochain-core/network"
	db "github.com/LemoFoundationLtd/lemochain-core/store/protocol"
	"sync"
	"sync/atomic"
	"github.com/LemoFoundationLtd/lemochain-distribution/database"
)

type BlockChain struct {
	chainID      uint16
	// db           db.ChainDB
	// am           *account.Manager
	currentBlock atomic.Value // latest block in current chain
	stableBlock  atomic.Value // latest stable block in current chain
	genesisBlock *types.Block // genesis block

	chainForksHead map[common.Hash]*types.Block // total latest header of different fork chain
	chainForksLock sync.Mutex
	mux            sync.Mutex
	running        int32
	dbEngine       database.DBEngine
}

func NewBlockChain(chainID uint16, db db.ChainDB) (bc *BlockChain, err error) {
	bc = &BlockChain{
		chainID:        chainID,
		// db:             db,
		chainForksHead: make(map[common.Hash]*types.Block, 16),
	}
	bc.dbEngine = database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	if err := bc.loadLastState(); err != nil {
		return nil, err
	}
	return bc, nil
}

// func (bc *BlockChain) AccountManager() *account.Manager {
// 	return bc.am
// }

// Lock call by miner
func (bc *BlockChain) Lock() *sync.Mutex {
	return &bc.mux
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
	bc.currentBlock.Store(block)
	bc.stableBlock.Store(block)
	// bc.am = account.NewManager(block.Hash(), bc.db)
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
		blockDao := database.NewBlockDao(bc.dbEngine)
		block, err = blockDao.GetBlockByHeight(height)
		if err != nil {
			panic(fmt.Sprintf("can't get block. height:%d, err: %v", height, err))
		}
	} else if height <= currentBlockHeight {
		for i := currentBlockHeight - height; i > 0; i-- {
			blockDao := database.NewBlockDao(bc.dbEngine)
			block, err = blockDao.GetBlock(block.ParentHash())
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
	if bc.currentBlock.Load() == nil {
		return nil
	}
	return bc.currentBlock.Load().(*types.Block)
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
		deputynode.Instance().Add(block.Height()+params.InterimDuration+1, block.DeputyNodes)
		log.Debugf("add new term deputy nodes: %v", block.DeputyNodes)
	}
}

func (bc *BlockChain) InsertChain(block *types.Block, isSynchronising bool) (error) {
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
	if err != nil{
		return err
	}else{
		bc.updateDeputyNodes(block)
		bc.currentBlock.Store(block)
		bc.stableBlock.Store(block)
		return nil
	}
}

// InsertChain insert block of non-self to chain
// func (bc *BlockChain) InsertChain(block *types.Block, isSynchronising bool) (err error) {
// 	bc.mux.Lock()
// 	defer bc.mux.Unlock()
//
// 	hash := block.Hash()
// 	if has, _ := bc.db.IsExistByHash(hash); has {
// 		return nil
// 	}
// 	if block.Height() > 0 {
// 		// process changelog
// 		if err := bc.am.RebuildAll(block); err != nil {
// 			log.Errorf("rebuild account manager failed: %v", err)
// 			return err
// 		}
// 		if err := bc.am.Finalise(); err != nil {
// 			panic("init genesis error01")
// 		}
// 	} else {
// 		bc.initGenesis(block)
// 	}
//
// 	if err = bc.db.SetBlock(hash, block); err != nil {
// 		log.Errorf("can't insert block to cache. height:%d hash:%s", block.Height(), hash.Prefix())
// 		return coreChain.ErrSaveBlock
// 	}
// 	log.Infof("Insert block to chain. height: %d. hash: %s. time: %d. parent: %s", block.Height(), block.Hash().Prefix(), block.Time(), block.ParentHash().Prefix())
//
// 	if err := bc.am.Save(hash); err != nil {
// 		log.Errorf("save account manager failed: %v", err)
// 		return err
// 	}
// 	if err = bc.db.SetStableBlock(hash); err != nil {
// 		log.Errorf("can't SetStableBlock. height:%d hash:%s", block.Height(), hash.Prefix())
// 		return coreChain.ErrSetStableBlockToDB
// 	}
//
// 	// update deputy nodes
// 	bc.updateDeputyNodes(block)
// 	bc.currentBlock.Store(block)
// 	bc.stableBlock.Store(block)
// 	return nil
// }

// func (bc *BlockChain) initGenesis(b *types.Block) {
// 	bc.am = account.NewManager(common.Hash{}, bc.db)
// 	total, _ := new(big.Int).SetString("1600000000000000000000000000", 10) // 1.6 billion
// 	bc.am.GetAccount(b.MinerAddress()).SetBalance(total)
// 	if err := bc.am.Finalise(); err != nil {
// 		panic("init genesis error02")
// 	}
// 	b.Header.VersionRoot = bc.am.GetVersionRoot()
// 	logs := bc.am.GetChangeLogs()
// 	b.SetChangeLogs(logs)
// 	b.Header.LogRoot = types.DeriveChangeLogsSha(logs)
// }

// func (bc *BlockChain) Db() db.ChainDB {
// 	// return bc.db
// }

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
