package chain

import (
	"encoding/json"
	"errors"
	"github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/chain/transaction"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
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
			if err := cl.Redo(engine); err != nil { // 通过changelog生成account的对应状态
				return err
			}
		}
	}
	//
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

// sealDbTx 组装需要存储到db中的Tx
func (engine *ReBuildEngine) sealDbTx(PHash, assetCode, assetId common.Hash, tx *types.Transaction) *database.Tx {
	to := common.Address{}
	if tx.To() != nil {
		to = *tx.To()
	}
	return &database.Tx{
		BHash:       engine.Block.Hash(),
		Height:      engine.Block.Height(),
		PHash:       PHash,
		THash:       tx.Hash(),
		From:        tx.From(),
		To:          to,
		Tx:          tx,
		Flag:        int(tx.Type()),
		St:          time.Now().UnixNano() / 1000000,
		PackageTime: engine.Block.Time(),
		AssetCode:   assetCode,
		AssetId:     assetId,
	}
}

// filterSaveAssetTx 过滤出资产交易并保存资产类型的交易到db,如果是资产类型的交易返回true
// 对于参数PHash,如果过滤的是BoxTx中的子交易，则PHash为BoxTx的hash,除此之外PHash == common.Hash{}
func (engine *ReBuildEngine) filterSaveAssetTx(PHash common.Hash, tx *types.Transaction, txDao *database.TxDao) (error, bool) {
	switch tx.Type() {

	case params.CreateAssetTx:
		assetCode := tx.Hash()
		dbTx := engine.sealDbTx(PHash, assetCode, common.Hash{}, tx)
		return txDao.Set(dbTx), true

	case params.IssueAssetTx:
		// 1. 获取资产交易中的assetCode和assetId
		issueAsset, err := types.GetIssueAsset(tx.Data())
		if err != nil {
			return err, true
		}
		assetCode := issueAsset.AssetCode
		assetDao := database.NewAssetDao(engine.Store)
		asset, err := assetDao.Get(assetCode)
		if err != nil {
			return err, true
		}
		AssType := asset.Category
		var assetId common.Hash
		if AssType == types.TokenAsset {
			assetId = assetCode
		} else if AssType == types.NonFungibleAsset || AssType == types.CommonAsset { // ERC721 or ERC721+20
			assetId = tx.Hash()
		} else {
			log.Errorf("Assert's Category not exist ,Category = %d ", AssType)
			return transaction.ErrAssetCategory, true
		}
		// 2. 保存交易进数据库
		dbTx := engine.sealDbTx(PHash, assetCode, assetId, tx)
		return txDao.Set(dbTx), true

	case params.ReplenishAssetTx:
		// 1. 获取资产交易中的assetCode和assetId
		repl, err := types.GetReplenishAsset(tx.Data())
		if err != nil {
			return err, true
		}
		// 2. 保存交易进数据库
		dbTx := engine.sealDbTx(PHash, repl.AssetCode, repl.AssetId, tx)
		return txDao.Set(dbTx), true

	case params.ModifyAssetTx:
		// 1. 获取资产交易中的assetCode
		modifyInfo, err := types.GetModifyAssetInfo(tx.Data())
		if err != nil {
			return err, true
		}
		assetCode := modifyInfo.AssetCode
		// 2. 保存交易进数据库
		dbTx := engine.sealDbTx(PHash, assetCode, common.Hash{}, tx)
		return txDao.Set(dbTx), true

	case params.TransferAssetTx:
		// 1. 获取资产交易中的assetId
		tradingAsset, err := types.GetTradingAsset(tx.Data())
		if err != nil {
			log.Errorf("Unmarshal transfer asset data err: %s", err)
			return err, true
		}
		assetId := tradingAsset.AssetId
		// 2. 保存交易进数据库
		dbTx := engine.sealDbTx(PHash, common.Hash{}, assetId, tx)
		return txDao.Set(dbTx), true

	default:
		return nil, false
	}
}

func (engine *ReBuildEngine) filterSaveBoxTx(boxTx *types.Transaction, txDao *database.TxDao) error {
	if boxTx.Type() != params.BoxTx {
		return errors.New("Must be box tx type ")
	}
	// 1 保存箱子中的子交易
	if box, err := types.GetBox(boxTx.Data()); err == nil {
		for _, subTx := range box.SubTxList {
			// 过滤资产相关的交易
			err, isExist := engine.filterSaveAssetTx(boxTx.Hash(), subTx, txDao)
			if err != nil {
				return err
			}
			if !isExist { // 不是资产类型的交易,单独执行保存操作
				if err := txDao.Set(engine.sealDbTx(boxTx.Hash(), common.Hash{}, common.Hash{}, boxTx)); err != nil {
					return err
				}
			}
		}
	} else {
		return err
	}
	// 2 保存箱子本身
	if err := txDao.Set(engine.sealDbTx(common.Hash{}, common.Hash{}, common.Hash{}, boxTx)); err != nil {
		return err
	}
	return nil
}

func (engine *ReBuildEngine) saveTx(tx *types.Transaction) error {
	txDao := database.NewTxDao(engine.Store)
	// 1. 过滤资产类型的交易
	err, isExist := engine.filterSaveAssetTx(common.Hash{}, tx, txDao)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}

	// 2. 过滤箱子交易
	if tx.Type() == params.BoxTx {
		return engine.filterSaveBoxTx(tx, txDao)
	}

	// 3. 其他交易
	return txDao.Set(engine.sealDbTx(common.Hash{}, common.Hash{}, common.Hash{}, tx))
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
				if asset == nil {
					return nil, errors.New("asset is nil,get asset by assetId ")
				}

				if asset.Category == types.TokenAsset {
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
