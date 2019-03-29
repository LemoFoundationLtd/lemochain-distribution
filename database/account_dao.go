package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"math/big"
)

func NewAccountData(address common.Address) *types.AccountData {
	account := &types.AccountData{Address: address}

	if account.Balance == nil {
		account.Balance = new(big.Int)
	}

	if account.NewestRecords == nil {
		account.NewestRecords = make(map[types.ChangeLogType]types.VersionRecord)
	}

	if account.Candidate.Profile == nil {
		account.Candidate.Profile = make(types.Profile)
	}

	if account.Candidate.Votes == nil {
		account.Candidate.Votes = new(big.Int)
	}

	return account
}

type AccountDao struct{
	engine *sql.DB
}

func NewAccountDao(db DBEngine) (*AccountDao) {
	return &AccountDao{engine:db.GetDB()}
}

func (dao *AccountDao) GetDB() (*sql.DB) {
	return dao.engine
}

func (dao *AccountDao) Get(addr common.Address) (*types.AccountData, error) {
	if addr == (common.Address{}) {
		log.Errorf("get account address is common.address{}")
		return nil, ErrArgInvalid
	}

	kvDao := NewKvDao(dao)
	val, err := kvDao.Get(GetAddressKey(addr))
	if err != nil {
		log.Errorf("get account.addr: " + addr.Hex() + ".err: " + err.Error())
		return nil, err
	}

	if val == nil{
		// log.Errorf("get account.is not exist.addr: " + addr.Hex())
		return nil, ErrNotExist
	}

	var account types.AccountData
	err = rlp.DecodeBytes(val, &account)
	if err != nil{
		return nil, err
	}else{
		return &account, nil
	}
}

func (dao *AccountDao) Set(addr common.Address, account *types.AccountData) (error) {
	if addr == (common.Address{}) || account == nil {
		log.Errorf("set account address is common.address{} or account is nil.")
		return ErrArgInvalid
	}

	val, err := rlp.EncodeToBytes(account)
	if err != nil {
		return err
	}

	kvDao := NewKvDao(dao)
	return kvDao.Set(GetAddressKey(addr), val)
}
