package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-go/chain/account"
	"github.com/LemoFoundationLtd/lemochain-go/chain/deputynode"
	coreParams "github.com/LemoFoundationLtd/lemochain-go/chain/params"
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-server/chain"
	"github.com/LemoFoundationLtd/lemochain-server/chain/params"
	"math/big"
	"time"
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
	accounts, err := a.GetAccount(LemoAddress)
	if err != nil {
		return "", err
	}
	balance := accounts.GetBalance().String()

	return balance, nil
}

// GetAccount return the struct of the &AccountData{}
func (a *PublicAccountAPI) GetAccount(LemoAddress string) (types.AccountAccessor, error) {
	address, err := common.StringToAddress(LemoAddress)
	if err != nil {
		return nil, err
	}

	accountData := a.manager.GetCanonicalAccount(address)
	// accountData := a.manager.GetAccount(address)
	return accountData, nil
}

// GetVoteFor
func (a *PublicAccountAPI) GetVoteFor(LemoAddress string) (string, error) {
	candiAccount, err := a.GetAccount(LemoAddress)
	if err != nil {
		return "", err
	}
	forAddress := candiAccount.GetVoteFor().String()
	return forAddress, nil
}

// GetAllRewardValue get the value for each bonus
func (a *PublicAccountAPI) GetAllRewardValue() ([]*coreParams.Reward, error) {
	address := coreParams.TermRewardPrecompiledContractAddress
	acc, err := a.GetAccount(address.String())
	if err != nil {
		return nil, err
	}
	key := address.Hash()
	value, err := acc.GetStorageState(key)
	rewardMap := make(coreParams.RewardsMap)
	json.Unmarshal(value, &rewardMap)
	var result = make([]*coreParams.Reward, 0)
	for _, v := range rewardMap {
		result = append(result, v)
	}
	return result, nil
}

// GetAssetEquity returns asset equity
func (a *PublicAccountAPI) GetAssetEquityByAssetId(LemoAddress string, assetId common.Hash) (*types.AssetEquity, error) {
	acc, err := a.GetAccount(LemoAddress)
	if err != nil {
		return nil, err
	}
	return acc.GetEquityState(assetId)
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
	return deputynode.Instance().GetLatestDeputies(c.chain.CurrentBlock().Height())
}

//
// // GetCandidateNodeList get all candidate node list information and return total candidate node
// func (c *PublicChainAPI) GetCandidateList(index, size int) (*CandidateListRes, error) {
// 	addresses, total, err := c.chain.Db().GetCandidatesPage(index, size)
// 	if err != nil {
// 		return nil, err
// 	}
// 	candidateList := make([]*CandidateInfo, 0, len(addresses))
// 	for i := 0; i < len(addresses); i++ {
// 		candidateAccount := c.chain.AccountManager().GetAccount(addresses[i])
// 		mapProfile := candidateAccount.GetCandidate()
// 		if isCandidate, ok := mapProfile[types.CandidateKeyIsCandidate]; !ok || isCandidate == coreParams.NotCandidateNode {
// 			err = fmt.Errorf("the node of %s is not candidate node", addresses[i].String())
// 			return nil, err
// 		}
//
// 		candidateInfo := &CandidateInfo{
// 			Profile: make(map[string]string),
// 		}
//
// 		candidateInfo.Profile = mapProfile
// 		candidateInfo.Votes = candidateAccount.GetVotes().String()
// 		candidateInfo.CandidateAddress = addresses[i].String()
//
// 		candidateList = append(candidateList, candidateInfo)
// 	}
// 	result := &CandidateListRes{
// 		CandidateList: candidateList,
// 		Total:         total,
// 	}
// 	return result, nil
// }

// GetCandidateTop30 get top 30 candidate node
func (c *PublicChainAPI) GetCandidateTop30() []*CandidateInfo {
	latestStableBlock := c.chain.StableBlock()
	stableBlockHash := latestStableBlock.Hash()
	storeInfos := c.chain.Db().GetCandidatesTop(stableBlockHash)
	candidateList := make([]*CandidateInfo, 0, 30)
	for _, info := range storeInfos {
		candidateInfo := &CandidateInfo{
			Profile: make(map[string]string),
		}
		CandidateAddress := info.GetAddress()
		CandidateAccount := c.chain.AccountManager().GetAccount(CandidateAddress)
		profile := CandidateAccount.GetCandidate()
		candidateInfo.Profile = profile
		candidateInfo.CandidateAddress = CandidateAddress.String()
		candidateInfo.Votes = info.GetTotal().String()
		candidateList = append(candidateList, candidateInfo)
	}
	return candidateList
}

// GetBlockByNumber get block information by height
func (c *PublicChainAPI) GetBlockByHeight(height uint32, withBody bool) *types.Block {
	if withBody {
		return c.chain.GetBlockByHeight(height)
	} else {
		block := c.chain.GetBlockByHeight(height)
		if block == nil {
			return nil
		}
		// copy only header
		onlyHeaderBlock := &types.Block{
			Header: block.Header,
		}
		return onlyHeaderBlock
	}
}

// GetBlockByHash get block information by hash
func (c *PublicChainAPI) GetBlockByHash(hash string, withBody bool) *types.Block {
	if withBody {
		return c.chain.GetBlockByHash(common.HexToHash(hash))
	} else {
		block := c.chain.GetBlockByHash(common.HexToHash(hash))
		if block == nil {
			return nil
		}
		// copy only header
		onlyHeaderBlock := &types.Block{
			Header: block.Header,
		}
		return onlyHeaderBlock
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
	if withBody {
		return c.chain.CurrentBlock()
	} else {
		currentBlock := c.chain.CurrentBlock()
		if currentBlock == nil {
			return nil
		}
		// copy only header
		onlyHeaderBlock := &types.Block{
			Header: currentBlock.Header,
		}
		return onlyHeaderBlock
	}
}

// LatestStableBlock get the latest currently agreed blocks
func (c *PublicChainAPI) LatestStableBlock(withBody bool) *types.Block {
	if withBody {
		return c.chain.StableBlock()
	} else {
		stableBlock := c.chain.StableBlock()
		if stableBlock == nil {
			return nil
		}
		// copy only header
		onlyHeaderBlock := &types.Block{
			Header: stableBlock.Header,
		}
		return onlyHeaderBlock
	}
}

// CurrentHeight
func (c *PublicChainAPI) CurrentHeight() uint32 {
	currentBlock := c.chain.CurrentBlock()
	height := currentBlock.Height()
	return height
}

// LatestStableHeight
func (c *PublicChainAPI) LatestStableHeight() uint32 {
	return c.chain.StableBlock().Height()
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
// func (t *PublicTxAPI) GetTxByHash(hash string) (*store.VTransactionDetail, error) {
// 	txHash := common.HexToHash(hash)
// 	bizDb := t.node.db.GetBizDatabase()
// 	vTxDetail, err := bizDb.GetTxByHash(txHash)
// 	return vTxDetail, err
// }
//
// //go:generate gencodec -type TxListRes --field-override txListResMarshaling -out gen_tx_list_res_json.go
// type TxListRes struct {
// 	VTransactions []*store.VTransaction `json:"txList" gencodec:"required"`
// 	Total         uint32                `json:"total" gencodec:"required"`
// }
// type txListResMarshaling struct {
// 	Total hexutil.Uint32
// }
//
// // GetTxListByAddress pull the list of transactions
// func (t *PublicTxAPI) GetTxListByAddress(lemoAddress string, index int, size int) (*TxListRes, error) {
// 	src, err := common.StringToAddress(lemoAddress)
// 	if err != nil {
// 		return nil, err
// 	}
// 	bizDb := t.node.db.GetBizDatabase()
// 	vTxs, total, err := bizDb.GetTxByAddr(src, index, size)
// 	if err != nil {
// 		return nil, err
// 	}
// 	txList := &TxListRes{
// 		VTransactions: vTxs,
// 		Total:         total,
// 	}
//
// 	return txList, nil
// }

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
