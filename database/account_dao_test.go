package database

import (
	"testing"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/stretchr/testify/assert"
)

func TestAccountDao_Get(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	accountDao := NewAccountDao(db)
	addr := common.HexToAddress("0xabc1123")
	account := &types.AccountData{Address:addr}
	err := accountDao.Set(addr, account)
	assert.NoError(t, err)

	result, err:= accountDao.Get(addr)
	assert.NoError(t, err)
	assert.Equal(t, addr, result.Address)
}

func TestAccountDao_GetNotExist(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	accountDao := NewAccountDao(db)
	addr := common.HexToAddress("0x01")

	result, err := accountDao.Get(addr)
	assert.Equal(t, err, ErrNotExist)
	assert.Nil(t, result)
}

func TestAccountDao_GetArgInvalid(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	accountDao := NewAccountDao(db)
	result, err := accountDao.Get(common.Address{})
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result)

	err = accountDao.Set(common.Address{}, new(types.AccountData))
	assert.Equal(t, err, ErrArgInvalid)

	err = accountDao.Set(common.HexToAddress("0x01"), nil)
	assert.Equal(t, err, ErrArgInvalid)
}
