package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	coreParams "github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/store"
	"github.com/LemoFoundationLtd/lemochain-distribution/chain/params"
	"github.com/LemoFoundationLtd/lemochain-distribution/database"
	"math/big"
	"time"
)

const (
	MaxTxToNameLength  = 100
	MaxTxMessageLength = 1024
)

var (
	ErrToName         = errors.New("the length of toName field in transaction is out of max length limit")
	ErrTxMessage      = errors.New("the length of message field in transaction is out of max length limit")
	ErrCreateContract = errors.New("the data of create contract transaction can't be null")
	ErrSpecialTx      = errors.New("the data of special transaction can't be null")
	ErrTxType         = errors.New("the transaction type does not exit")
)

// Private
type PrivateAccountAPI struct {
	node *Node
}

// NewPrivateAccountAPI
func NewPrivateAccountAPI(node *Node) *PrivateAccountAPI {
	return &PrivateAccountAPI{node: node}
}

// NewAccount get lemo address api
func (a *PrivateAccountAPI) NewKeyPair() (*crypto.AccountKey, error) {
	accountKey, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}
	return accountKey, nil
}

// PublicAccountAPI API for access to account information
type PublicAccountAPI struct {
	node *Node
}

// NewPublicAccountAPI
func NewPublicAccountAPI(node *Node) *PublicAccountAPI {
	return &PublicAccountAPI{node: node}
}

// GetBalance get balance in mo
func (a *PublicAccountAPI) GetBalance(LemoAddress string) (string, error) {
	account, err := a.GetAccount(LemoAddress)
	if err != nil {
		return "", err
	}
	balance := account.Balance.String()

	return balance, nil
}

// GetAccount return the struct of the &AccountData{}
func (a *PublicAccountAPI) GetAccount(LemoAddress string) (*types.AccountData, error) {
	address, err := common.StringToAddress(LemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(a.node.config.DbDriver, a.node.config.DbUri)
	defer dbEngine.Close()

	accountDao := database.NewAccountDao(dbEngine)
	accountData, err := accountDao.Get(address)
	if err == database.ErrNotExist {
		return &types.AccountData{Address: address, Balance: big.NewInt(0), Candidate: types.Candidate{Votes: big.NewInt(0)}}, nil
	}
	return accountData, err
}

// GetAssetEquity returns asset equity
func (a *PublicAccountAPI) GetAssetEquityByAssetId(LemoAddress string, assetId common.Hash) (*types.AssetEquity, error) {
	address, err := common.StringToAddress(LemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(a.node.config.DbDriver, a.node.config.DbUri)
	defer dbEngine.Close()

	equityDao := database.NewEquityDao(dbEngine)
	return equityDao.Get(address, assetId)
}

//go:generate gencodec -type AssetEquityBatchRsp --field-override AssetEquityBatchRspMarshaling -out gen_asset_equity_rsp_json.go
type AssetEquityBatchRsp struct {
	Equities []*types.AssetEquity `json:"equities" gencodec:"required"`
	Total    uint32               `json:"total" gencodec:"required"`
}

type AssetEquityBatchRspMarshaling struct {
	Total hexutil.Uint32
}

func (a *PublicAccountAPI) GetAssetEquityByAssetCode(LemoAddress string, assetCode common.Hash, index, limit int) (*AssetEquityBatchRsp, error) {
	address, err := common.StringToAddress(LemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(a.node.config.DbDriver, a.node.config.DbUri)
	defer dbEngine.Close()

	equityDao := database.NewEquityDao(dbEngine)
	result, total, err := equityDao.GetPageByCodeWithTotal(address, assetCode, index, limit)
	if err != nil {
		return nil, err
	} else {
		return &AssetEquityBatchRsp{
			Equities: result,
			Total:    uint32(total),
		}, nil
	}
}

func (a *PublicAccountAPI) GetAssetEquity(LemoAddress string, index, limit int) (*AssetEquityBatchRsp, error) {
	address, err := common.StringToAddress(LemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(a.node.config.DbDriver, a.node.config.DbUri)
	defer dbEngine.Close()

	equityDao := database.NewEquityDao(dbEngine)
	result, total, err := equityDao.GetPageWithTotal(address, index, limit)
	if err != nil {
		return nil, err
	} else {
		return &AssetEquityBatchRsp{
			Equities: result,
			Total:    uint32(total),
		}, nil
	}
}

func (a *PublicAccountAPI) GetAsset(assetCode common.Hash) (*types.Asset, error) {
	dbEngine := database.NewMySqlDB(a.node.config.DbDriver, a.node.config.DbUri)
	defer dbEngine.Close()

	assetDao := database.NewAssetDao(dbEngine)
	return assetDao.Get(assetCode)
}

func (a *PublicAccountAPI) GetMetaData(assetId common.Hash) (*database.MetaData, error) {
	dbEngine := database.NewMySqlDB(a.node.config.DbDriver, a.node.config.DbUri)
	defer dbEngine.Close()

	metaDataDao := database.NewMetaDataDao(dbEngine)
	return metaDataDao.Get(assetId)
}

//go:generate gencodec -type CandidateInfo -out gen_candidate_info_json.go
type CandidateInfo struct {
	CandidateAddress string            `json:"address" gencodec:"required"`
	Votes            string            `json:"votes" gencodec:"required"`
	Profile          map[string]string `json:"profile"  gencodec:"required"`
}

// ChainAPI
type PublicChainAPI struct {
	node *Node
}

// NewChainAPI API for access to chain information
func NewPublicChainAPI(node *Node) *PublicChainAPI {
	return &PublicChainAPI{node: node}
}

//go:generate gencodec -type CandidateListRes --field-override candidateListResMarshaling -out gen_candidate_list_res_json.go
type CandidateListRes struct {
	CandidateList []*CandidateInfo `json:"candidateList" gencodec:"required"`
	Total         uint32           `json:"total" gencodec:"required"`
}
type candidateListResMarshaling struct {
	Total hexutil.Uint32
}

//go:generate gencodec -type DeputyNodeInfo --field-override deputyNodeInfoMarshaling -out gen_deputyNode_info_json.go
type DeputyNodeInfo struct {
	MinerAddress  common.Address `json:"minerAddress"   gencodec:"required"` // candidate account address
	IncomeAddress common.Address `json:"incomeAddress" gencodec:"required"`
	NodeID        []byte         `json:"nodeID"         gencodec:"required"`
	Rank          uint32         `json:"rank"           gencodec:"required"` // start from 0
	Votes         *big.Int       `json:"votes"          gencodec:"required"`
	Host          string         `json:"host"          gencodec:"required"`
	Port          string         `json:"port"          gencodec:"required"`
	DepositAmount string         `json:"depositAmount"  gencodec:"required"` // 质押金额
	Introduction  string         `json:"introduction"  gencodec:"required"`  // 节点介绍
	P2pUri        string         `json:"p2pUri"  gencodec:"required"`        // p2p 连接用的定位符
}

type deputyNodeInfoMarshaling struct {
	NodeID hexutil.Bytes
	Rank   hexutil.Uint32
	Votes  *hexutil.Big10
}

// GetAllRewardValue get the value for each bonus
func (c *PublicChainAPI) GetAllRewardValue() (coreParams.RewardsMap, error) {
	address := coreParams.TermRewardContract
	dbEngine := database.NewMySqlDB(c.node.config.DbDriver, c.node.config.DbUri)
	defer dbEngine.Close()
	kvDao := database.NewKvDao(dbEngine)

	value, err := kvDao.Get(database.GetStorageKey(address.Hash()))
	if err != nil {
		return nil, err
	}
	rewardMap := make(coreParams.RewardsMap)
	// return empty map if the reward not exist
	if len(value) == 0 {
		return rewardMap, nil
	}
	err = json.Unmarshal(value, &rewardMap)
	return rewardMap, err
}

//go:generate gencodec -type TermRewardInfo --field-override termRewardInfoMarshaling -out gen_termReward_info_json.go
type TermRewardInfo struct {
	Term         uint32   `json:"term" gencodec:"required"`
	Value        *big.Int `json:"value" gencodec:"required"`
	RewardHeight uint32   `json:"rewardHeight" gencodec:"required"`
}
type termRewardInfoMarshaling struct {
	Term         hexutil.Uint32
	Value        *hexutil.Big10
	RewardHeight hexutil.Uint32
}

func (c *PublicChainAPI) GetTermReward(height uint32) (*TermRewardInfo, error) {
	term := deputynode.GetTermIndexByHeight(height)
	termValueMap, err := c.GetAllRewardValue()
	if err != nil {
		return nil, err
	}
	if reward, ok := termValueMap[term]; ok {
		return &TermRewardInfo{
			Term:         reward.Term,
			Value:        reward.Value,
			RewardHeight: (term+1)*coreParams.TermDuration + coreParams.InterimDuration + 1,
		}, nil
	} else {
		return &TermRewardInfo{
			Term:         term,
			Value:        new(big.Int),
			RewardHeight: (term+1)*coreParams.TermDuration + coreParams.InterimDuration + 1,
		}, nil
	}
}

// GetDeputyNodeList get deputy nodes who are in charge
func (c *PublicChainAPI) GetDeputyNodeList() []*DeputyNodeInfo {
	nodes := c.node.chain.DeputyManager().GetDeputiesByHeight(c.node.chain.StableBlock().Height())
	dbEngine := database.NewMySqlDB(c.node.config.DbDriver, c.node.config.DbUri)
	defer dbEngine.Close()

	accountDao := database.NewAccountDao(dbEngine)

	var result []*DeputyNodeInfo
	for _, n := range nodes {
		candidateAcc, err := accountDao.Get(n.MinerAddress)
		if err != nil {
			log.Errorf("Get minerAddress accountData error: %v", err)
			continue
		}
		profile := candidateAcc.Candidate.Profile
		incomeAddress, err := common.StringToAddress(profile[types.CandidateKeyIncomeAddress])
		if err != nil {
			log.Errorf("incomeAddress string to address type.incomeAddress: %s.error: %v", profile[types.CandidateKeyIncomeAddress], err)
			continue
		}
		host := profile[types.CandidateKeyHost]
		port := profile[types.CandidateKeyPort]
		nodeAddrString := fmt.Sprintf("%x@%s:%s", n.NodeID, host, port)
		deputyNodeInfo := &DeputyNodeInfo{
			MinerAddress:  n.MinerAddress,
			IncomeAddress: incomeAddress,
			NodeID:        n.NodeID,
			Rank:          n.Rank,
			Votes:         n.Votes,
			Host:          host,
			Port:          port,
			DepositAmount: profile[types.CandidateKeyDepositAmount],
			Introduction:  profile[types.CandidateKeyIntroduction],
			P2pUri:        nodeAddrString,
		}
		result = append(result, deputyNodeInfo)
	}
	return result
}

// // GetCandidateNodeList get all candidate node list information and return total candidate node
func (c *PublicChainAPI) GetCandidateList(index, size int) (*CandidateListRes, error) {
	dbEngine := database.NewMySqlDB(c.node.config.DbDriver, c.node.config.DbUri)
	defer dbEngine.Close()

	candidateDao := database.NewCandidateDao(dbEngine)
	candidates, total, err := candidateDao.GetPageWithTotal(index, size)
	if err != nil {
		return nil, err
	}

	accountDao := database.NewAccountDao(dbEngine)
	result := make([]*CandidateInfo, len(candidates))
	for index := 0; index < len(candidates); index++ {
		account, err := accountDao.Get(candidates[index].User)
		if err != nil {
			return nil, err
		}

		result[index] = &CandidateInfo{
			Votes:            candidates[index].Votes.String(),
			Profile:          account.Candidate.Profile,
			CandidateAddress: candidates[index].User.String(),
		}
	}

	return &CandidateListRes{
		CandidateList: result,
		Total:         uint32(total),
	}, nil
}

// GetCandidateTop30 get top 30 candidate node
func (c *PublicChainAPI) GetCandidateTop30() []*CandidateInfo {
	result := make([]*CandidateInfo, 0)

	dbEngine := database.NewMySqlDB(c.node.config.DbDriver, c.node.config.DbUri)
	defer dbEngine.Close()

	candidateDao := database.NewCandidateDao(dbEngine)
	candidateItems, err := candidateDao.GetTop(20)
	if err != nil {
		return result
	}

	accountDao := database.NewAccountDao(dbEngine)
	for _, info := range candidateItems {
		account, err := accountDao.Get(info.User)
		if err != nil {
			return result
		}

		result = append(result, &CandidateInfo{
			Votes:            account.Candidate.Votes.String(),
			Profile:          account.Candidate.Profile,
			CandidateAddress: account.Address.String(),
		})
	}
	return result
}

// GetBlockByNumber get block information by height
func (c *PublicChainAPI) GetBlockByHeight(height uint32, withBody bool) *types.Block {
	dbEngine := database.NewMySqlDB(c.node.config.DbDriver, c.node.config.DbUri)
	defer dbEngine.Close()

	blockDao := database.NewBlockDao(dbEngine)
	block, err := blockDao.GetBlockByHeight(height)
	if err != nil {
		log.Errorf("get block by height.err: " + err.Error())
		return nil
	}

	if withBody {
		return block
	} else {
		return &types.Block{
			Header: block.Header,
		}
	}
}

// GetBlockByHash get block information by hash
func (c *PublicChainAPI) GetBlockByHash(hash string, withBody bool) *types.Block {
	dbEngine := database.NewMySqlDB(c.node.config.DbDriver, c.node.config.DbUri)
	defer dbEngine.Close()

	blockDao := database.NewBlockDao(dbEngine)
	block, err := blockDao.GetBlock(common.HexToHash(hash))
	if err != nil {
		log.Errorf("get block by hash.err: " + err.Error())
		return nil
	}

	if withBody {
		return block
	} else {
		return &types.Block{
			Header: block.Header,
		}
	}
}

// ChainID get chain id
func (c *PublicChainAPI) ChainID() uint16 {
	return c.node.chain.ChainID()
}

// Genesis get the creation block
func (c *PublicChainAPI) Genesis() *types.Block {
	return c.node.chain.Genesis()
}

// CurrentBlock get the current latest block
func (c *PublicChainAPI) CurrentBlock(withBody bool) *types.Block {
	dbEngine := database.NewMySqlDB(c.node.config.DbDriver, c.node.config.DbUri)
	defer dbEngine.Close()

	contextDao := database.NewContextDao(dbEngine)
	block, err := contextDao.GetCurrentBlock()
	if err != nil {
		log.Errorf("get current block. err: " + err.Error())
		return nil
	}

	if withBody {
		return block
	} else {
		return &types.Block{
			Header: block.Header,
		}
	}
}

// GasPriceAdvice get suggest gas price
func (c *PublicChainAPI) GasPriceAdvice() *big.Int {
	// todo
	return big.NewInt(100000000)
}

// GetServerVersion
func (c *PublicChainAPI) NodeVersion() string {
	return params.Version
}

// TXAPI
type PublicTxAPI struct {
	// txpool *chain.TxPool
	node *Node
}

// NewTxAPI API for send a transaction
func NewPublicTxAPI(node *Node) *PublicTxAPI {
	return &PublicTxAPI{node}
}

// Send send a transaction
func (t *PublicTxAPI) SendTx(tx *types.Transaction) (common.Hash, error) {
	err := tx.VerifyTxBody(t.node.chain.ChainID(), uint64(time.Now().Unix()), false)
	if err != nil {
		return common.Hash{}, err
	}
	err = t.node.txPool.AddTx(tx)
	return tx.Hash(), err
}

// SendReimbursedGasTx gas代付交易 todo 测试使用
func (t *PublicTxAPI) SendReimbursedGasTx(senderPrivate, gasPayerPrivate string, from, to, gasPayer common.Address, amount int64, data []byte, txType uint16, toName, message string) (common.Hash, error) {
	tx := types.NewReimbursementTransaction(from, to, gasPayer, big.NewInt(amount), data, txType, t.node.chain.ChainID(), uint64(time.Now().Unix()+1800), toName, message)
	senderPriv, _ := crypto.HexToECDSA(senderPrivate)
	gasPayerPriv, _ := crypto.HexToECDSA(gasPayerPrivate)
	firstSignTx, err := types.MakeReimbursementTxSigner().SignTx(tx, senderPriv)
	if err != nil {
		return common.Hash{}, err
	}
	signTx := types.GasPayerSignatureTx(firstSignTx, common.Big1, uint64(60000))
	lastSignTx, err := types.MakeGasPayerSigner().SignTx(signTx, gasPayerPriv)
	if err != nil {
		return common.Hash{}, err
	}
	err = lastSignTx.VerifyTxBody(t.node.chain.ChainID(), uint64(time.Now().Unix()), false)
	if err != nil {
		return common.Hash{}, err
	}
	err = t.node.txPool.AddTx(lastSignTx)
	return lastSignTx.Hash(), err
}

// CreateAsset 创建资产
func (t *PublicTxAPI) CreateAsset(prv string, category, decimals uint32, isReplenishable, isDivisible bool) (common.Hash, error) {
	private, _ := crypto.HexToECDSA(prv)
	issuer := crypto.PubkeyToAddress(private.PublicKey)
	profile := make(types.Profile)
	profile[types.AssetName] = "Demo Token"
	profile[types.AssetSymbol] = "DT"
	profile[types.AssetDescription] = "test issue token"
	profile[types.AssetFreeze] = "false"
	profile[types.AssetSuggestedGasLimit] = "60000"
	asset := &types.Asset{
		Category:        category,
		IsDivisible:     isDivisible,
		AssetCode:       common.Hash{},
		Decimal:         decimals,
		TotalSupply:     big.NewInt(1000000),
		IsReplenishable: isReplenishable,
		Issuer:          issuer,
		Profile:         profile,
	}
	data, err := json.Marshal(asset)
	if err != nil {
		return common.Hash{}, err
	}
	tx := types.NoReceiverTransaction(issuer, nil, uint64(500000), big.NewInt(1), data, coreParams.CreateAssetTx, t.node.chain.ChainID(), uint64(time.Now().Unix()+30*60), "", "create asset tx")
	signTx, err := types.MakeSigner().SignTx(tx, private)
	if err != nil {
		return common.Hash{}, err
	}
	return t.SendTx(signTx)
}

// 发行资产
func (t *PublicTxAPI) IssueAsset(prv string, receiver common.Address, assetCode common.Hash, amount *big.Int, metaData string) (common.Hash, error) {
	issue := &types.IssueAsset{
		AssetCode: assetCode,
		MetaData:  metaData,
		Amount:    amount,
	}
	data, err := json.Marshal(issue)
	if err != nil {
		return common.Hash{}, err
	}
	private, _ := crypto.HexToECDSA(prv)
	sender := crypto.PubkeyToAddress(private.PublicKey)
	tx := types.NewTransaction(sender, receiver, nil, uint64(500000), big.NewInt(1), data, coreParams.IssueAssetTx, t.node.chain.ChainID(), uint64(time.Now().Unix()+30*60), "", "issue asset tx")

	signTx, err := types.MakeSigner().SignTx(tx, private)
	if err != nil {
		return common.Hash{}, err
	}
	return t.SendTx(signTx)
}

// 增发资产
func (t *PublicTxAPI) ReplenishAsset(prv string, receiver common.Address, assetCode, assetId common.Hash, amount *big.Int) (common.Hash, error) {
	repl := &types.ReplenishAsset{
		AssetCode: assetCode,
		AssetId:   assetId,
		Amount:    amount,
	}
	data, err := json.Marshal(repl)
	if err != nil {
		return common.Hash{}, err
	}
	private, _ := crypto.HexToECDSA(prv)
	sender := crypto.PubkeyToAddress(private.PublicKey)
	tx := types.NewTransaction(sender, receiver, nil, uint64(500000), big.NewInt(1), data, coreParams.ReplenishAssetTx, t.node.chain.ChainID(), uint64(time.Now().Unix()+30*60), "", "replenish asset tx")

	signTx, err := types.MakeSigner().SignTx(tx, private)
	if err != nil {
		return common.Hash{}, err
	}
	return t.SendTx(signTx)
}

// ModifyAsset 修改资产信息
func (t *PublicTxAPI) ModifyAsset(prv string, assetCode common.Hash) (common.Hash, error) {
	info := make(types.Profile)
	info["name"] = "Modify"
	info["stop"] = "true"
	modify := &types.ModifyAssetInfo{
		AssetCode:     assetCode,
		UpdateProfile: info,
	}
	data, err := json.Marshal(modify)
	if err != nil {
		return common.Hash{}, err
	}
	private, _ := crypto.HexToECDSA(prv)
	sender := crypto.PubkeyToAddress(private.PublicKey)
	tx := types.NoReceiverTransaction(sender, nil, uint64(500000), big.NewInt(1), data, coreParams.ModifyAssetTx, t.node.chain.ChainID(), uint64(time.Now().Unix()+30*60), "", "modify asset tx")
	signTx, err := types.MakeSigner().SignTx(tx, private)
	if err != nil {
		return common.Hash{}, err
	}
	return t.SendTx(signTx)
}

// 交易资产
func (t *PublicTxAPI) TransferAsset(prv string, to common.Address, assetCode, assetId common.Hash, amount *big.Int, input []byte) (common.Hash, error) {
	transfer := &types.TransferAsset{
		AssetId: assetId,
		Amount:  amount,
		Input:   input,
	}
	data, err := json.Marshal(transfer)
	if err != nil {
		return common.Hash{}, err
	}
	private, _ := crypto.HexToECDSA(prv)
	sender := crypto.PubkeyToAddress(private.PublicKey)
	tx := types.NewTransaction(sender, to, amount, uint64(500000), big.NewInt(1), data, coreParams.TransferAssetTx, t.node.chain.ChainID(), uint64(time.Now().Unix()+30*60), "", "trading asset tx")
	signTx, err := types.MakeSigner().SignTx(tx, private)
	if err != nil {
		return common.Hash{}, err
	}
	return t.SendTx(signTx)
}

// // PendingTx
// func (t *PublicTxAPI) PendingTx(size int) []*types.Transaction {
// 	return t.node.txPool.Pending(size)
// }

// // GetTxByHash pull the specified transaction through a transaction hash
func (t *PublicTxAPI) GetTxByHash(hash string) (*store.VTransactionDetail, error) {
	txHash := common.HexToHash(hash)

	dbEngine := database.NewMySqlDB(t.node.config.DbDriver, t.node.config.DbUri)
	defer dbEngine.Close()

	txDao := database.NewTxDao(dbEngine)
	tx, err := txDao.Get(txHash)
	if err != nil {
		if database.ErrIsNotExist(err) {
			return nil, nil
		}
		return nil, err
	} else {
		return &store.VTransactionDetail{
			BlockHash:   tx.BHash,
			PHash:       tx.PHash,
			Height:      tx.Height,
			Tx:          tx.Tx,
			PackageTime: tx.PackageTime,
			AssetCode:   tx.AssetCode,
			AssetId:     tx.AssetId,
		}, nil
	}
}

//go:generate gencodec -type TxListRes --field-override txListResMarshaling -out gen_tx_list_res_json.go
type TxListRes struct {
	VTransactions []*store.VTransaction `json:"txList" gencodec:"required"`
	Total         uint32                `json:"total" gencodec:"required"`
}
type txListResMarshaling struct {
	Total hexutil.Uint32
}

// // GetTxListByAddress pull the list of transactions
func (t *PublicTxAPI) GetTxListByAddress(lemoAddress string, index int, size int) (*TxListRes, error) {
	src, err := common.StringToAddress(lemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(t.node.config.DbDriver, t.node.config.DbUri)
	defer dbEngine.Close()

	txDao := database.NewTxDao(dbEngine)
	txes, total, err := txDao.GetByAddrWithTotal(src, index, size)
	if err != nil {
		return nil, err
	}

	result := make([]*store.VTransaction, len(txes))
	for index := 0; index < len(txes); index++ {
		result[index] = &store.VTransaction{
			Tx:          txes[index].Tx,
			PHash:       txes[index].PHash,
			PackageTime: txes[index].PackageTime,
			AssetCode:   txes[index].AssetCode,
			AssetId:     txes[index].AssetId,
		}
	}

	return &TxListRes{
		VTransactions: result,
		Total:         uint32(total),
	}, nil
}

func (t *PublicTxAPI) GetTxListByTimestamp(lemoAddress string, beginTime int64, endTime int64, index int, size int) (*TxListRes, error) {
	src, err := common.StringToAddress(lemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(t.node.config.DbDriver, t.node.config.DbUri)
	defer dbEngine.Close()

	txDao := database.NewTxDao(dbEngine)
	txes, total, err := txDao.GetByTimeWithTotal(src, beginTime, endTime, index, size)
	if err != nil {
		return nil, err
	}

	result := make([]*store.VTransaction, len(txes))
	for index := 0; index < len(txes); index++ {
		result[index] = &store.VTransaction{
			Tx:          txes[index].Tx,
			PHash:       txes[index].PHash,
			PackageTime: txes[index].PackageTime,
			AssetCode:   txes[index].AssetCode,
			AssetId:     txes[index].AssetId,
		}
	}

	return &TxListRes{
		VTransactions: result,
		Total:         uint32(total),
	}, nil
}

// GetAssetTxList 通过assetCode或者assetId获取与此地址相关的交易列表
func (t *PublicTxAPI) GetAssetTxList(lemoAddress string, assetCodeOrId common.Hash, index int, size int) (*TxListRes, error) {
	addr, err := common.StringToAddress(lemoAddress)
	if err != nil {
		return nil, err
	}
	dbEngine := database.NewMySqlDB(t.node.config.DbDriver, t.node.config.DbUri)
	defer dbEngine.Close()

	txDao := database.NewTxDao(dbEngine)
	txes, total, err := txDao.GetByAddressAndAssetCodeOrAssetIdWithTotal(addr, assetCodeOrId, index, size)
	if err != nil {
		return nil, err
	}
	result := make([]*store.VTransaction, len(txes))
	for index := 0; index < len(txes); index++ {
		result[index] = &store.VTransaction{
			Tx:          txes[index].Tx,
			PHash:       txes[index].PHash,
			PackageTime: txes[index].PackageTime,
			AssetCode:   txes[index].AssetCode,
			AssetId:     txes[index].AssetId,
		}
	}

	return &TxListRes{
		VTransactions: result,
		Total:         uint32(total),
	}, nil
}

// // ReadContract read variables in a contract includes the return value of a function.
// func (t *PublicTxAPI) ReadContract(to *common.Address, data hexutil.Bytes) (string, error) {
// 	ctx := context.Background()
// 	result, _, err := t.doCall(ctx, to, coreParams.OrdinaryTx, data, 5*time.Second)
// 	return common.ToHex(result), err
// }
//
// // EstimateGas returns an estimate of the amount of gas needed to execute the given transaction.
// func (t *PublicTxAPI) EstimateGas(to *common.Address, txType uint8, data hexutil.Bytes) (string, error) {
// 	var costGas uint64
// 	var err error
// 	ctx := context.Background()
// 	_, costGas, err = t.doCall(ctx, to, txType, data, 5*time.Second)
// 	strCostGas := strconv.FormatUint(costGas, 10)
// 	return strCostGas, err
// }
//
// // EstimateContractGas returns an estimate of the amount of gas needed to create a smart contract.
// // todo will delete
// func (t *PublicTxAPI) EstimateCreateContractGas(data hexutil.Bytes) (uint64, error) {
// 	ctx := context.Background()
// 	_, costGas, err := t.doCall(ctx, nil, coreParams.OrdinaryTx, data, 5*time.Second)
// 	return costGas, err
// }

// // doCall
// func (t *PublicTxAPI) doCall(ctx context.Context, to *common.Address, txType uint8, data hexutil.Bytes, timeout time.Duration) ([]byte, uint64, error) {
// 	t.node.lock.Lock()
// 	defer t.node.lock.Unlock()
//
// 	defer func(start time.Time) { log.Debug("Executing EVM call finished", "runtime", time.Since(start)) }(time.Now())
// 	// get latest stableBlock
// 	stableBlock := t.node.chain.StableBlock()
// 	log.Infof("stable block height = %v", stableBlock.Height())
// 	stableHeader := stableBlock.Header
//
// 	p := t.node.chain.TxProcessor()
// 	ret, costGas, err := p.CallTx(ctx, stableHeader, to, txType, data, common.Hash{}, timeout)
//
// 	return ret, costGas, err
// }
