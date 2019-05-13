package chain

import (
	"encoding/json"
	"github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-distribution/database"
	"time"
)

type ReBuildEngine struct {
	Store                database.DBEngine
	Block                *types.Block
	ReBuildAccountsCache map[common.Address]*ReBuildAccount
}

func NewReBuildEngine(store database.DBEngine, block *types.Block) *ReBuildEngine {
	return &ReBuildEngine{
		Store:                store,
		Block:                block,
		ReBuildAccountsCache: make(map[common.Address]*ReBuildAccount),
	}
}

func (engine *ReBuildEngine) GetAccount(address common.Address) types.AccountAccessor {
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

func (engine *ReBuildEngine) ReBuild() error {
	logs := engine.Block.ChangeLogs
	if len(logs) > 0 {
		for _, cl := range logs {
			if err := cl.Redo(engine); err != nil {
				return err
			}
		}
	}

	err := engine.resolve()
	if err != nil {
		return err
	}

	return engine.Save()
}

func (engine *ReBuildEngine) Save() error {
	err := engine.saveBlock(engine.Block)
	if err != nil {
		return err
	}

	err = engine.saveAccountBatch(engine.ReBuildAccountsCache)
	if err != nil {
		return err
	}

	err = engine.saveTxBatch(engine.Block.Txs)
	if err != nil {
		return err
	}

	return engine.saveCurrentBlock(engine.Block)
}

func (engine *ReBuildEngine) saveCurrentBlock(block *types.Block) error {
	contextDao := database.NewContextDao(engine.Store)
	return contextDao.SetCurrentBlock(block)
}

func (engine *ReBuildEngine) saveBlock(block *types.Block) error {
	blockDao := database.NewBlockDao(engine.Store)
	return blockDao.SetBlock(block.Hash(), block)
}

func (engine *ReBuildEngine) saveAccountBatch(reBuildAccounts map[common.Address]*ReBuildAccount) error {
	for _, v := range reBuildAccounts {
		data := v.BuildAccountData()
		err := engine.saveAccount(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (engine *ReBuildEngine) saveAccount(account *types.AccountData) error {
	accountDao := database.NewAccountDao(engine.Store)
	return accountDao.Set(account.Address, account)
}

func (engine *ReBuildEngine) saveTxBatch(txes []*types.Transaction) error {
	for index := 0; index < len(txes); index++ {
		err := engine.saveTx(txes[index])
		if err != nil {
			return err
		}
	}

	return nil
}

func (engine *ReBuildEngine) saveTx(tx *types.Transaction) error {
	txDao := database.NewTxDao(engine.Store)

	from, err := tx.From()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	to := tx.To()
	if to == nil {
		return txDao.Set(&database.Tx{
			THash:  tx.Hash(),
			BHash:  engine.Block.Hash(),
			Height: engine.Block.Height(),
			From:   from,
			To:     common.Address{},
			Tx:     tx,
			St:     time.Now().UnixNano() / 1000000,
		})
	} else {
		return txDao.Set(&database.Tx{
			THash:  tx.Hash(),
			BHash:  engine.Block.Hash(),
			Height: engine.Block.Height(),
			From:   from,
			To:     *to,
			Tx:     tx,
			St:     time.Now().UnixNano() / 1000000,
		})
	}

	return nil
}

func (engine *ReBuildEngine) saveStorageBatch(storage map[common.Hash][]byte) error {
	for k, v := range storage {
		err := engine.saveStorage(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (engine *ReBuildEngine) saveStorage(hash common.Hash, val []byte) error {
	kvDao := database.NewKvDao(engine.Store)
	return kvDao.Set(hash.Bytes(), val)
}

func (engine *ReBuildEngine) saveAssetCodeBatch(assets map[common.Hash]*types.Asset) error {
	for _, v := range assets {
		err := engine.saveAssetCode(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (engine *ReBuildEngine) saveAssetCode(asset *types.Asset) error {
	assetCodeDao := database.NewAssetDao(engine.Store)
	return assetCodeDao.Set(asset)
}

func (engine *ReBuildEngine) saveAssetIdBatch(address common.Address, assetIds map[common.Hash]*types.IssueAsset) error {
	for k, v := range assetIds {
		err := engine.saveAssetId(address, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (engine *ReBuildEngine) saveAssetId(address common.Address, hash common.Hash, assetId *types.IssueAsset) error {
	assetIdDao := database.NewMetaDataDao(engine.Store)
	return assetIdDao.Set(&database.MetaData{
		Id:      hash,
		Code:    assetId.AssetCode,
		Owner:   address,
		Profile: assetId.MetaData,
	})
}

func (engine *ReBuildEngine) saveEquitiesBatch(address common.Address, equities map[common.Hash]*types.AssetEquity) error {
	for _, v := range equities {
		err := engine.saveEquity(address, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (engine *ReBuildEngine) saveEquity(address common.Address, equity *types.AssetEquity) error {
	equityDao := database.NewEquityDao(engine.Store)
	return equityDao.Set(address, equity)
}

func (engine *ReBuildEngine) getAssetIds() (map[common.Hash]*types.IssueAsset, error) {
	assetDao := database.NewAssetDao(engine.Store)
	txes := engine.Block.Txs
	assetIds := make(map[common.Hash]*types.IssueAsset)
	for index := 0; index < len(txes); index++ {
		tx := txes[index]
		if tx.Type() == params.IssueAssetTx {
			extendData := tx.Data()
			if len(extendData) <= 0 {
				panic("tx is issue asset. but data is nil.")
			}

			issueAsset := &types.IssueAsset{}
			err := json.Unmarshal(extendData, issueAsset)
			if err != nil {
				return nil, err
			} else {
				asset, err := assetDao.Get(issueAsset.AssetCode)
				if err != nil {
					return nil, err
				}

				if asset.Category == types.Asset01 {
					assetIds[asset.AssetCode] = issueAsset
				} else {
					assetIds[tx.Hash()] = issueAsset
				}
			}
		}
	}

	return assetIds, nil
}

func (engine *ReBuildEngine) resolve() error {
	assetIds, err := engine.getAssetIds()
	if err != nil {
		return err
	}

	for _, v := range engine.ReBuildAccountsCache {
		if len(v.AssetCodes) > 0 {
			AssetCodeCache := make(map[common.Hash]*types.Asset)
			for ak, av := range v.AssetCodes {
				AssetCodeCache[ak] = av
			}

			engine.saveAssetCodeBatch(AssetCodeCache)
		}

		if len(v.AssetIds) > 0 {
			AssetIdCache := make(map[common.Hash]*types.IssueAsset)
			for ak, av := range v.AssetIds {
				AssetCode := assetIds[ak].AssetCode
				AssetIdCache[ak] = &types.IssueAsset{
					AssetCode: AssetCode,
					MetaData:  av,
				}
			}

			engine.saveAssetIdBatch(v.Address, AssetIdCache)
		}

		if len(v.AssetEquities) > 0 {
			EquityCache := make(map[common.Hash]*types.AssetEquity)
			for ak, av := range v.AssetEquities {
				EquityCache[ak] = av
			}

			engine.saveEquitiesBatch(v.Address, EquityCache)
		}

		if len(v.Storage) > 0 {
			StorageCache := make(map[common.Hash][]byte)
			for ak, av := range v.Storage {
				StorageCache[ak] = av
			}

			engine.saveStorageBatch(StorageCache)
		}

		isCandidate := v.isCandidate(v.Candidate.Profile)
		if isCandidate && v.IsCancelCandidate {
			panic("not exist at the same time.")
		}

		if isCandidate {
			candidateDao := database.NewCandidateDao(engine.Store)
			err := candidateDao.Set(&database.CandidateItem{
				User:  v.Address,
				Votes: v.Candidate.Votes,
			})
			if err != nil {
				return err
			}
		}

		if v.IsCancelCandidate {
			candidateDao := database.NewCandidateDao(engine.Store)
			err := candidateDao.Del(v.Address)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
