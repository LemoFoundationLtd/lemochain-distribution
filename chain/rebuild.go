package chain

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-distribution/database"
	"time"
)

type ReBuildEngine struct {
	Store database.DBEngine

	Block     *types.Block
	LogsCache []*types.ChangeLog
	TxCache   []*types.Transaction

	ReBuildAccountsCache map[common.Address]*ReBuildAccount
	AssetCodeCache       map[common.Hash]*types.Asset
	AssetIdCache         map[common.Hash]string
	EquityCache          map[common.Hash]*types.AssetEquity
	StorageCache         map[common.Hash][]byte
	ChangeCandidates     map[common.Address]bool
}

func NewReBuildEngine(store database.DBEngine, block *types.Block) (*ReBuildEngine) {
	return &ReBuildEngine{
		Store:                store,
		LogsCache:            block.ChangeLogs,
		Block:                block,
		ReBuildAccountsCache: make(map[common.Address]*ReBuildAccount),
	}
}

func (engine *ReBuildEngine) GetAccount(address common.Address) (types.AccountAccessor) {
	reBuildAccount, ok := engine.ReBuildAccountsCache[address]
	if ok {
		return reBuildAccount
	}

	accountDao := database.NewAccountDao(engine.Store)
	account, err := accountDao.Get(address)
	if err != nil {
		if err == database.ErrNotExist {
			account = database.NewAccountData(address)
		} else {
			panic("get account from database err: " + err.Error())
		}
	}

	reBuildAccount = NewReBuildAccount(engine.Store, account)
	engine.ReBuildAccountsCache[address] = reBuildAccount
	return reBuildAccount
}

func (engine *ReBuildEngine) Close() {
	//
}

func (engine *ReBuildEngine) ReBuild() (error) {
	if len(engine.LogsCache) > 0 {
		for _, cl := range engine.LogsCache {
			if err := cl.Redo(engine); err != nil {
				return err
			}
		}
	}

	engine.resolve()
	return engine.Save()
}

func (engine *ReBuildEngine) Save() (error) {
	err := engine.saveBlock(engine.Block)
	if err != nil {
		return err
	}

	err = engine.saveAccountBatch(engine.ReBuildAccountsCache)
	if err != nil {
		return err
	}

	err = engine.saveTxBatch(engine.TxCache)
	if err != nil {
		return err
	}

	err = engine.saveStorageBatch(engine.StorageCache)
	if err != nil {
		return err
	}

	err = engine.saveAssetCodeBatch(engine.AssetCodeCache)
	if err != nil {
		return err
	}

	err = engine.saveAssetIdBatch(engine.AssetIdCache)
	if err != nil {
		return err
	}

	err = engine.saveEquitiesBatch(engine.EquityCache)
	if err != nil {
		return err
	}

	return engine.saveCurrentBlock(engine.Block)
}

func (engine *ReBuildEngine) saveCurrentBlock(block *types.Block) (error) {
	contextDao := database.NewContextDao(engine.Store)
	return contextDao.SetCurrentBlock(block)
}

func (engine *ReBuildEngine) saveBlock(block *types.Block) (error) {
	blockDao := database.NewBlockDao(engine.Store)
	return blockDao.SetBlock(block.Hash(), block)
}

func (engine *ReBuildEngine) saveAccountBatch(reBuildAccounts map[common.Address]*ReBuildAccount) (error) {
	for _, v := range reBuildAccounts {
		data := v.BuildAccountData()
		err := engine.saveAccount(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (engine *ReBuildEngine) saveAccount(account *types.AccountData) (error) {
	accountDao := database.NewAccountDao(engine.Store)
	return accountDao.Set(account.Address, account)
}

func (engine *ReBuildEngine) saveTxBatch(txes []*types.Transaction) (error) {
	for index := 0; index < len(txes); index++ {
		err := engine.saveTx(txes[index])
		if err != nil {
			return err
		}
	}

	return nil
}

func (engine *ReBuildEngine) saveTx(tx *types.Transaction) (error) {
	txDao := database.NewTxDao(engine.Store)

	from, err := tx.From()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return txDao.Set(&database.Tx{
		THash:  tx.Hash(),
		BHash:  engine.Block.Hash(),
		Height: engine.Block.Height(),
		From:   from,
		To:     *tx.To(),
		Tx:     tx,
		St:     time.Now().UnixNano() / 1000000,
	})
	return nil
}

func (engine *ReBuildEngine) saveStorageBatch(storages map[common.Hash][]byte) (error) {
	for k, v := range storages {
		err := engine.saveStorage(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (engine *ReBuildEngine) saveStorage(hash common.Hash, val []byte) (error) {
	return nil
}

func (engine *ReBuildEngine) saveAssetCodeBatch(assets map[common.Hash]*types.Asset) (error) {
	for k, v := range assets {
		err := engine.saveAssetCode(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (engine *ReBuildEngine) saveAssetCode(code common.Hash, asset *types.Asset) (error) {
	return nil
}

func (engine *ReBuildEngine) saveAssetIdBatch(assetIds map[common.Hash]string) (error) {
	for k, v := range assetIds {
		err := engine.saveAssetId(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (engine *ReBuildEngine) saveAssetId(id common.Hash, val string) (error) {
	return nil
}

func (engine *ReBuildEngine) saveEquitiesBatch(equities map[common.Hash]*types.AssetEquity) (error) {
	for k, v := range equities {
		err := engine.saveEquity(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (engine *ReBuildEngine) saveEquity(id common.Hash, equity *types.AssetEquity) (error) {
	return nil
}

func (engine *ReBuildEngine) saveCandidates() (error) {
	candidateDao := database.NewCandidateDao(engine.Store)
	for k, v := range engine.ChangeCandidates {
		if v {
			err := candidateDao.Set(&database.CandidateItem{
				User:  k,
				Votes: engine.ReBuildAccountsCache[k].Candidate.Votes,
			})

			if err != nil {
				return err
			}
		} else {
			// del
		}
	}

	return nil
}

func (engine *ReBuildEngine) resolve() {

	engine.LogsCache = engine.Block.ChangeLogs
	engine.TxCache = engine.Block.Txs

	engine.AssetCodeCache = make(map[common.Hash]*types.Asset)
	engine.AssetIdCache = make(map[common.Hash]string)
	engine.EquityCache = make(map[common.Hash]*types.AssetEquity)
	engine.StorageCache = make(map[common.Hash][]byte)
	for _, v := range engine.ReBuildAccountsCache {
		if len(v.AssetCodes) > 0 {
			for ak, av := range v.AssetCodes {
				engine.AssetCodeCache[ak] = av
			}
		}

		if len(v.AssetIds) > 0 {
			for ak, av := range v.AssetIds {
				engine.AssetIdCache[ak] = av
			}
		}

		if len(v.AssetEquities) > 0 {
			for ak, av := range v.AssetEquities {
				engine.EquityCache[ak] = av
			}
		}

		if len(v.Storage) > 0 {
			for ak, av := range v.Storage {
				engine.StorageCache[ak] = av
			}
		}

		isCandidate := v.isCandidate(v.Candidate.Profile)
		if isCandidate && v.IsCancelCandidate {
			panic("not exist at the same time.")
		}

		if isCandidate {
			engine.ChangeCandidates[v.Address] = true
		}

		if v.IsCancelCandidate {
			engine.ChangeCandidates[v.Address] = false
		}
	}
}
