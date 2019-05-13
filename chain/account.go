package chain

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-distribution/database"
	"math/big"
	"strconv"
)

type ReBuildAccount struct {
	types.AccountData

	Store database.DBEngine
	Code  []byte

	AssetCodes    map[common.Hash]*types.Asset
	AssetIds      map[common.Hash]string
	AssetEquities map[common.Hash]*types.AssetEquity
	Storage       map[common.Hash][]byte
	Events        []*types.Event

	IsCancelCandidate bool

	NextVersion map[types.ChangeLogType]uint32
	suicided    bool
}

func NewReBuildAccount(store database.DBEngine, data *types.AccountData) *ReBuildAccount {
	reBuildAccount := &ReBuildAccount{Store: store, AccountData: *data}

	reBuildAccount.NextVersion = make(map[types.ChangeLogType]uint32)
	for k, v := range reBuildAccount.NewestRecords {
		reBuildAccount.NextVersion[k] = v.Version
	}

	reBuildAccount.AssetCodes = make(map[common.Hash]*types.Asset)
	reBuildAccount.AssetIds = make(map[common.Hash]string)
	reBuildAccount.AssetEquities = make(map[common.Hash]*types.AssetEquity)
	reBuildAccount.Storage = make(map[common.Hash][]byte)
	reBuildAccount.Events = make([]*types.Event, 0)
	return reBuildAccount
}

func (account *ReBuildAccount) isCandidate(profile types.Profile) bool {
	if len(profile) <= 0 {
		return false
	}

	result, ok := profile[types.CandidateKeyIsCandidate]
	if !ok {
		return false
	}

	val, err := strconv.ParseBool(result)
	if err != nil {
		panic("to bool err : " + err.Error())
	} else {
		return val
	}
}

func (account *ReBuildAccount) BuildAccountData() *types.AccountData {
	return &account.AccountData
}

func (account *ReBuildAccount) GetAddress() common.Address {
	return account.Address
}

func (account *ReBuildAccount) GetVersion(logType types.ChangeLogType) uint32 {
	if account.NewestRecords == nil {
		return 1
	}

	version, ok := account.NewestRecords[logType]
	if !ok {
		return 1
	} else {
		return version.Version
	}
}

func (account *ReBuildAccount) GetNextVersion(logType types.ChangeLogType) uint32 {
	version, ok := account.NextVersion[logType]
	if !ok {
		account.NextVersion[logType] = 1
		return account.NextVersion[logType]
	} else {
		account.NextVersion[logType] = version + 1
		return account.NextVersion[logType]
	}
}

func (account *ReBuildAccount) GetVoteFor() common.Address {
	return account.VoteFor
}

func (account *ReBuildAccount) SetVoteFor(addr common.Address) {
	account.VoteFor = addr
}

func (account *ReBuildAccount) GetVotes() *big.Int {
	return new(big.Int).Set(account.Candidate.Votes)
}

func (account *ReBuildAccount) SetVotes(votes *big.Int) {
	account.Candidate.Votes.Set(votes)
}

func (account *ReBuildAccount) GetCandidate() types.Profile {
	result := make(types.Profile)
	if len(account.Candidate.Profile) <= 0 {
		return result
	} else {
		for k, v := range account.Candidate.Profile {
			result[k] = v
		}
		return result
	}
}

func (account *ReBuildAccount) SetCandidate(profile types.Profile) {
	if account.isCandidate(account.Candidate.Profile) && !account.isCandidate(profile) {
		account.IsCancelCandidate = true
	}

	account.Candidate.Profile = make(types.Profile)
	if len(profile) <= 0 {
		return
	}

	for k, v := range profile {
		account.Candidate.Profile[k] = v
	}
}

func (account *ReBuildAccount) GetCandidateState(key string) string {
	val, ok := account.Candidate.Profile[key]
	if !ok {
		return ""
	} else {
		return val
	}
}

func (account *ReBuildAccount) SetCandidateState(key string, val string) {
	account.Candidate.Profile[key] = val
}

func (account *ReBuildAccount) GetBalance() *big.Int {
	return new(big.Int).Set(account.Balance)
}

func (account *ReBuildAccount) SetBalance(balance *big.Int) {
	account.Balance.Set(balance)
}

func (account *ReBuildAccount) GetCodeHash() common.Hash {
	return account.CodeHash
}

func (account *ReBuildAccount) SetCodeHash(codeHash common.Hash) {
	account.CodeHash = codeHash
}

func (account *ReBuildAccount) GetCode() (types.Code, error) {
	if account.Code == nil {
		kvDao := database.NewKvDao(account.Store)
		val, err := kvDao.Get(account.CodeHash.Bytes())
		if err != nil {
			return nil, err
		} else {
			account.Code = val
			return account.Code, nil
		}
	} else {
		return account.Code, nil
	}
}

func (account *ReBuildAccount) SetCode(code types.Code) {
	account.Code = code
}

func (account *ReBuildAccount) GetStorageRoot() common.Hash {
	return account.StorageRoot
}

func (account *ReBuildAccount) SetStorageRoot(root common.Hash) {
	account.StorageRoot = root
}

func (account *ReBuildAccount) GetAssetCodeRoot() common.Hash {
	return account.AssetCodeRoot
}

func (account *ReBuildAccount) SetAssetCodeRoot(root common.Hash) {
	account.AssetCodeRoot = root
}

func (account *ReBuildAccount) GetAssetIdRoot() common.Hash {
	return account.AssetIdRoot
}

func (account *ReBuildAccount) SetAssetIdRoot(root common.Hash) {
	account.AssetIdRoot = root
}

func (account *ReBuildAccount) GetEquityRoot() common.Hash {
	return account.EquityRoot
}

func (account *ReBuildAccount) SetEquityRoot(root common.Hash) {
	account.EquityRoot = root
}

func (account *ReBuildAccount) GetStorageState(key common.Hash) ([]byte, error) {
	val, ok := account.Storage[key]
	if !ok {
		return nil, nil
	} else {
		return val, nil
	}
}

func (account *ReBuildAccount) SetStorageState(key common.Hash, value []byte) error {
	account.Storage[key] = value
	return nil
}

func (account *ReBuildAccount) GetAssetCodeFromDB(code common.Hash) (*types.Asset, error) {
	assetDao := database.NewAssetDao(account.Store)
	return assetDao.Get(code)
}

func (account *ReBuildAccount) GetAssetCode(code common.Hash) (*types.Asset, error) {
	asset, ok := account.AssetCodes[code]
	if !ok {
		tmp, err := account.GetAssetCodeFromDB(code)
		if err != nil {
			return nil, err
		}

		account.AssetCodes[code] = tmp
		return tmp, nil
	} else {
		return asset, nil
	}
}

func (account *ReBuildAccount) SetAssetCode(code common.Hash, asset *types.Asset) error {
	if asset == nil {
		account.AssetCodes[code] = nil
		return nil
	} else {
		account.AssetCodes[code] = asset.Clone()
		return nil
	}
}

func (account *ReBuildAccount) GetAssetCodeTotalSupply(code common.Hash) (*big.Int, error) {
	asset, ok := account.AssetCodes[code]
	if !ok {
		tmp, err := account.GetAssetCodeFromDB(code)
		if err != nil {
			return nil, err
		}

		if tmp == nil {
			return nil, nil
		}

		account.AssetCodes[code] = tmp
		return new(big.Int).Set(tmp.TotalSupply), nil
	} else {
		if asset == nil {
			return nil, nil
		} else {
			return new(big.Int).Set(asset.TotalSupply), nil
		}
	}
}

func (account *ReBuildAccount) SetAssetCodeTotalSupply(code common.Hash, val *big.Int) error {
	asset, ok := account.AssetCodes[code]
	if !ok {
		tmp, err := account.GetAssetCodeFromDB(code)
		if err != nil {
			return err
		}

		if tmp == nil {
			return database.ErrNotExist
		}

		tmp.TotalSupply.Set(val)
		account.AssetCodes[code] = tmp
		return nil
	} else {
		asset.TotalSupply.Set(val)
		return nil
	}
}

func (account *ReBuildAccount) GetAssetCodeState(code common.Hash, key string) (string, error) {
	asset, ok := account.AssetCodes[code]
	if !ok {
		tmp, err := account.GetAssetCodeFromDB(code)
		if err != nil {
			return "", err
		}

		if tmp == nil {
			return "", nil
		}

		account.AssetCodes[code] = tmp
		return tmp.Profile[key], nil
	} else {
		if asset == nil {
			return "", nil
		} else {
			return account.AssetCodes[code].Profile[key], nil
		}
	}
}

func (account *ReBuildAccount) SetAssetCodeState(code common.Hash, key string, val string) error {
	asset, ok := account.AssetCodes[code]
	if !ok {
		tmp, err := account.GetAssetCodeFromDB(code)
		if err != nil {
			return err
		}

		if tmp == nil {
			return database.ErrNotExist
		}

		tmp.Profile[key] = val
		account.AssetCodes[code] = tmp
		return nil
	} else {
		if asset == nil {
			return database.ErrNotExist
		} else {
			account.AssetCodes[code].Profile[key] = val
			return nil
		}
	}
}

func (account *ReBuildAccount) GetAssetId(id common.Hash) (string, error) {
	metaDataDao := database.NewMetaDataDao(account.Store)
	val, err := metaDataDao.Get(id)
	if err != nil {
		return "", err
	} else {
		return val.Profile, nil
	}
}

func (account *ReBuildAccount) GetAssetIdState(id common.Hash) (string, error) {
	assetId, ok := account.AssetIds[id]
	if !ok {
		metaDataDao := database.NewMetaDataDao(account.Store)
		metaData, err := metaDataDao.Get(id)
		if err != nil {
			return "", err
		} else {
			account.AssetIds[id] = metaData.Profile
			return metaData.Profile, nil
		}
	} else {
		return assetId, nil
	}
}

func (account *ReBuildAccount) SetAssetIdState(id common.Hash, data string) error {
	account.AssetIds[id] = data
	return nil
}

func (account *ReBuildAccount) GetEquity(id common.Hash) (*types.AssetEquity, error) {
	equityDao := database.NewEquityDao(account.Store)
	return equityDao.Get(account.Address, id)
}

func (account *ReBuildAccount) GetEquityState(id common.Hash) (*types.AssetEquity, error) {
	equity, ok := account.AssetEquities[id]
	if !ok {
		tmp, err := account.GetEquity(id)
		if err != nil {
			return nil, err
		} else {
			account.AssetEquities[id] = tmp
			return tmp, nil
		}
	} else {
		return equity.Clone(), nil
	}
}

func (account *ReBuildAccount) SetEquityState(id common.Hash, equity *types.AssetEquity) error {
	if equity == nil {
		account.AssetEquities[id] = nil
		return nil
	} else {
		account.AssetEquities[id] = equity.Clone()
		return nil
	}
}

func (account *ReBuildAccount) PushEvent(event *types.Event) {
	account.Events = append(account.Events, event)
}

func (account *ReBuildAccount) PopEvent() error {
	size := len(account.Events)
	if size == 0 {
		return database.ErrNoEvents
	}
	account.Events = account.Events[:size-1]
	return nil
}

func (account *ReBuildAccount) GetEvents() []*types.Event {
	return account.Events
}

func (account *ReBuildAccount) GetSuicide() bool {
	return account.suicided
}

func (account *ReBuildAccount) SetSuicide(suicided bool) {
	if suicided {
		account.SetBalance(new(big.Int))
		account.SetCodeHash(common.Hash{})
		account.SetStorageRoot(common.Hash{})
		account.SetAssetCodeRoot(common.Hash{})
		account.SetAssetIdRoot(common.Hash{})
	}
	account.suicided = suicided
}

func (account *ReBuildAccount) IsEmpty() bool {
	for _, record := range account.NewestRecords {
		if record.Version != 0 {
			return false
		}
	}
	return true
}

func (account *ReBuildAccount) MarshalJSON() ([]byte, error) {
	panic("implement me")
}
