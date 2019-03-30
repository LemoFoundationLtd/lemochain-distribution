package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/chain/account"
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	coreParams "github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-distribution/chain"
	"github.com/LemoFoundationLtd/lemochain-distribution/chain/params"
	"math/big"
	"time"
	"github.com/LemoFoundationLtd/lemochain-distribution/database"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/store"
)

const (
	MaxTxToNameLength  = 100
	MaxTxMessageLength = 1024
)

// Private
type PrivateAccountAPI struct {
	manager *account.Manager
}

// NewPrivateAccountAPI
func NewPrivateAccountAPI(m *account.Manager) *PrivateAccountAPI {
	return &PrivateAccountAPI{m}
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
	manager *account.Manager
}

// NewPublicAccountAPI
func NewPublicAccountAPI(m *account.Manager) *PublicAccountAPI {
	return &PublicAccountAPI{m}
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

	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	accountDao := database.NewAccountDao(dbEngine)
	return accountDao.Get(address)
}

// GetAllRewardValue get the value for each bonus
func (a *PublicAccountAPI) GetAllRewardValue() ([]*coreParams.Reward, error) {
	// address := coreParams.TermRewardPrecompiledContractAddress
	// acc, err := a.GetAccount(address.String())
	// if err != nil {
	// 	return nil, err
	// }
	// key := address.Hash()
	// value, err := acc.GetStorageState(key)
	// rewardMap := make(coreParams.RewardsMap)
	// json.Unmarshal(value, &rewardMap)
	// var result = make([]*coreParams.Reward, 0)
	// for _, v := range rewardMap {
	// 	result = append(result, v)
	// }
	// return result, nil
	return nil, nil
}

// GetAssetEquity returns asset equity
func (a *PublicAccountAPI) GetAssetEquityByAssetId(LemoAddress string, assetId common.Hash) (*types.AssetEquity, error) {
	address, err := common.StringToAddress(LemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	equityDao := database.NewEquityDao(dbEngine)
	return equityDao.Get(address, assetId)
}

//go:generate gencodec -type AssetEquityBatchRsp -out gen_asset_equity_rsp_json.go
type AssetEquityBatchRsp struct {
	Equities []*types.AssetEquity `json:"equities" gencodec:"required"`
	Total    uint32               `json:"total" gencodec:"required"`
}

func (a *PublicAccountAPI) GetAssetEquityByAssetCode(LemoAddress string, assetCode common.Hash, index, limit int) (*AssetEquityBatchRsp, error) {
	address, err := common.StringToAddress(LemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	equityDao := database.NewEquityDao(dbEngine)
	result, total, err := equityDao.GetPageByCodeWithTotal(address, assetCode, index, limit)
	if err != nil{
		return nil, err
	}else{
		return &AssetEquityBatchRsp{
			Equities: result,
			Total:uint32(total),
		}, nil
	}
}

func (a *PublicAccountAPI) GetAssetEquity(LemoAddress string, index, limit int)(*AssetEquityBatchRsp, error) {
	address, err := common.StringToAddress(LemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	equityDao := database.NewEquityDao(dbEngine)
	result, total, err := equityDao.GetPageWithTotal(address, index, limit)
	if err != nil{
		return nil, err
	}else{
		return &AssetEquityBatchRsp{
			Equities: result,
			Total:uint32(total),
		}, nil
	}
}

func (a *PublicAccountAPI) GetAsset(assetCode common.Hash)(*types.Asset, error){
	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	assetDao := database.NewAssetDao(dbEngine)
	return assetDao.Get(assetCode)
}

func (a *PublicAccountAPI) GetMateData(assetId common.Hash)(*database.MateData, error) {
	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	mateDataDao := database.NewMateDataDao(dbEngine)
	return mateDataDao.Get(assetId)
}

//go:generate gencodec -type CandidateInfo -out gen_candidate_info_json.go
type CandidateInfo struct {
	CandidateAddress string            `json:"address" gencodec:"required"`
	Votes            string            `json:"votes" gencodec:"required"`
	Profile          map[string]string `json:"profile"  gencodec:"required"`
}

// ChainAPI
type PublicChainAPI struct {
	chain *chain.BlockChain
}

// NewChainAPI API for access to chain information
func NewPublicChainAPI(chain *chain.BlockChain) *PublicChainAPI {
	return &PublicChainAPI{chain}
}

//go:generate gencodec -type CandidateListRes --field-override candidateListResMarshaling -out gen_candidate_list_res_json.go
type CandidateListRes struct {
	CandidateList []*CandidateInfo `json:"candidateList" gencodec:"required"`
	Total         uint32           `json:"total" gencodec:"required"`
}
type candidateListResMarshaling struct {
	Total hexutil.Uint32
}

// GetDeputyNodeList
func (c *PublicChainAPI) GetDeputyNodeList() []string {
	result := deputynode.Instance().GetLatestDeputies(c.chain.CurrentBlock().Height())
	return result
}

//
// // GetCandidateNodeList get all candidate node list information and return total candidate node
func (c *PublicChainAPI) GetCandidateList(index, size int) (*CandidateListRes, error) {
	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
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

	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
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
	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
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
	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
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
	return c.chain.ChainID()
}

// Genesis get the creation block
func (c *PublicChainAPI) Genesis() *types.Block {
	return c.chain.Genesis()
}

// CurrentBlock get the current latest block
func (c *PublicChainAPI) CurrentBlock(withBody bool) *types.Block {
	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
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
func (c *PublicChainAPI) GetServerVersion() string {
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
	err := AvailableTx(tx)
	if err != nil {
		return common.Hash{}, err
	}
	err = t.node.txPool.AddTx(tx)
	return tx.Hash(), err
}

// SendReimbursedGasTx gas代付交易 todo 测试使用
func (t *PublicTxAPI) SendReimbursedGasTx(senderPrivate, gasPayerPrivate string, to, gasPayer common.Address, amount int64, data []byte, txType uint8, toName, message string) (common.Hash, error) {
	tx := types.NewReimbursementTransaction(to, gasPayer, big.NewInt(amount), data, txType, t.node.chainID, uint64(time.Now().Unix()+1800), toName, message)
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
	err = AvailableTx(lastSignTx)
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
	profile[types.AssetStop] = "false"
	profile[types.AssetSuggestedGasLimit] = "60000"
	asset := &types.Asset{
		Category:        category,
		IsDivisible:     isDivisible,
		AssetCode:       common.Hash{},
		Decimals:        decimals,
		TotalSupply:     big.NewInt(100000),
		IsReplenishable: isReplenishable,
		Issuer:          issuer,
		Profile:         profile,
	}
	data, err := json.Marshal(asset)
	if err != nil {
		return common.Hash{}, err
	}
	tx := types.NoReceiverTransaction(nil, uint64(500000), big.NewInt(1), data, coreParams.CreateAssetTx, t.node.chainID, uint64(time.Now().Unix()+30*60), "", "create asset tx")
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
	tx := types.NewTransaction(receiver, nil, uint64(500000), big.NewInt(1), data, coreParams.IssueAssetTx, t.node.chainID, uint64(time.Now().Unix()+30*60), "", "issue asset tx")
	private, _ := crypto.HexToECDSA(prv)
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
	tx := types.NewTransaction(receiver, nil, uint64(500000), big.NewInt(1), data, coreParams.ReplenishAssetTx, t.node.chainID, uint64(time.Now().Unix()+30*60), "", "replenish asset tx")
	private, _ := crypto.HexToECDSA(prv)
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
		AssetCode: assetCode,
		Info:      info,
	}
	data, err := json.Marshal(modify)
	if err != nil {
		return common.Hash{}, err
	}
	tx := types.NoReceiverTransaction(nil, uint64(500000), big.NewInt(1), data, coreParams.ModifyAssetTx, t.node.chainID, uint64(time.Now().Unix()+30*60), "", "modify asset tx")
	private, _ := crypto.HexToECDSA(prv)
	signTx, err := types.MakeSigner().SignTx(tx, private)
	if err != nil {
		return common.Hash{}, err
	}
	return t.SendTx(signTx)
}

// 交易资产
func (t *PublicTxAPI) TradingAsset(prv string, to common.Address, assetCode, assetId common.Hash, amount *big.Int, input []byte) (common.Hash, error) {
	trading := &types.TradingAsset{
		AssetId: assetId,
		Value:   amount,
		Input:   input,
	}
	data, err := json.Marshal(trading)
	if err != nil {
		return common.Hash{}, err
	}
	tx := types.NewTransaction(to, amount, uint64(500000), big.NewInt(1), data, coreParams.TradingAssetTx, t.node.chainID, uint64(time.Now().Unix()+30*60), "", "trading asset tx")
	private, _ := crypto.HexToECDSA(prv)
	signTx, err := types.MakeSigner().SignTx(tx, private)
	if err != nil {
		return common.Hash{}, err
	}
	return t.SendTx(signTx)
}

// AvailableTx transaction parameter verification
func AvailableTx(tx *types.Transaction) error {
	toNameLength := len(tx.ToName())
	if toNameLength > MaxTxToNameLength {
		toNameErr := fmt.Errorf("the length of toName field in transaction is out of max length limit. toName length = %d. max length limit = %d. ", toNameLength, MaxTxToNameLength)
		return toNameErr
	}
	txMessageLength := len(tx.Message())
	if txMessageLength > MaxTxMessageLength {
		txMessageErr := fmt.Errorf("the length of message field in transaction is out of max length limit. message length = %d. max length limit = %d. ", txMessageLength, MaxTxMessageLength)
		return txMessageErr
	}
	switch tx.Type() {
	case coreParams.OrdinaryTx:
		if tx.To() == nil {
			if len(tx.Data()) == 0 {
				createContractErr := errors.New("The data of contract creation transaction can't be null ")
				return createContractErr
			}
		}
	case coreParams.VoteTx:
	case coreParams.RegisterTx:
		if len(tx.Data()) == 0 {
			registerTxErr := errors.New("The data of contract creation transaction can't be null ")

			return registerTxErr
		}
	default:
		txTypeErr := fmt.Errorf("transaction type error. txType = %v", tx.Type())
		return txTypeErr
	}
	return nil
}

// // PendingTx
// func (t *PublicTxAPI) PendingTx(size int) []*types.Transaction {
// 	return t.node.txPool.Pending(size)
// }

// // GetTxByHash pull the specified transaction through a transaction hash
func (t *PublicTxAPI) GetTxByHash(hash string) (*store.VTransactionDetail, error) {
	txHash := common.HexToHash(hash)

	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	txDao := database.NewTxDao(dbEngine)
	tx, err := txDao.Get(txHash)
	if err != nil {
		return nil, err
	} else {
		return &store.VTransactionDetail{
			BlockHash: tx.BHash,
			Height:    tx.Height,
			Tx:        tx.Tx,
			St:        tx.St,
		}, nil
	}
}

//
//go:generate gencodec -type TxListRes --field-override txListResMarshaling -out gen_tx_list_res_json.go
type TxListRes struct {
	VTransactions []*store.VTransaction `json:"txList" gencodec:"required"`
	Total         uint32                `json:"total" gencodec:"required"`
}
type txListResMarshaling struct {
	Total hexutil.Uint32
}

//
// // GetTxListByAddress pull the list of transactions
func (t *PublicTxAPI) GetTxListByAddress(lemoAddress string, index int, size int) (*TxListRes, error) {
	src, err := common.StringToAddress(lemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	txDao := database.NewTxDao(dbEngine)
	txes, total, err := txDao.GetByAddrWithTotal(src, index, size)
	if err != nil {
		return nil, err
	}

	result := make([]*store.VTransaction, len(txes))
	for index := 0; index < len(txes); index++ {
		result[index] = &store.VTransaction{
			Tx: txes[index].Tx,
			St: txes[index].St,
		}
	}

	return &TxListRes{
		VTransactions: result,
		Total:         uint32(total),
	}, nil
}

func (t *PublicTxAPI) GetTxListByTimestamp(lemoAddress string, beginTime int64, endTime int64, index int, size int)(*TxListRes, error){
	src, err := common.StringToAddress(lemoAddress)
	if err != nil {
		return nil, err
	}

	dbEngine := database.NewMySqlDB(database.DRIVER_MYSQL, database.DNS_MYSQL)
	defer dbEngine.Close()

	txDao := database.NewTxDao(dbEngine)
	txes, total, err := txDao.GetByTimeWithTotal(src, beginTime, endTime, index, size)
	if err != nil {
		return nil, err
	}

	result := make([]*store.VTransaction, len(txes))
	for index := 0; index < len(txes); index++ {
		result[index] = &store.VTransaction{
			Tx: txes[index].Tx,
			St: txes[index].St,
		}
	}

	return &TxListRes{
		VTransactions: result,
		Total:         uint32(total),
	}, nil
}

// ReadContract read variables in a contract includes the return value of a function.
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
