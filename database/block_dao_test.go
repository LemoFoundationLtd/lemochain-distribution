package database

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"testing"
)

func TestBlockDao_GetBlock(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	blockDao := NewBlockDao(db)

	hash := common.HexToHash("0xabc")
	block := new(types.Block)
	block.Header = &types.Header{Height: 1}
	err := blockDao.SetBlock(hash, block)
	assert.NoError(t, err)

	result, err := blockDao.GetBlock(hash)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), result.Height())

	result, err = blockDao.GetBlockByHeight(1)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), result.Height())
}

func TestBlockDao_GetNotExist(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	blockDao := NewBlockDao(db)
	hash := common.HexToHash("0xabc")
	result, err := blockDao.GetBlock(hash)
	assert.Equal(t, err, ErrNotExist)
	assert.Nil(t, result)

	result, err = blockDao.GetBlockByHeight(1)
	assert.Equal(t, err, ErrNotExist)
	assert.Nil(t, result)
}

func TestBlockDao_GetArgInvalid(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	blockDao := NewBlockDao(db)
	err := blockDao.SetBlock(common.Hash{}, new(types.Block))
	assert.Equal(t, err, ErrArgInvalid)

	err = blockDao.SetBlock(common.HexToHash("0x01"), nil)
	assert.Equal(t, err, ErrArgInvalid)

	block, err := blockDao.GetBlock(common.Hash{})
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, block)
}
