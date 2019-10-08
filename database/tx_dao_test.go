package database

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewTx(thash common.Hash) *Tx {
	return &Tx{
		BHash:     common.HexToHash("0xabcde"),
		THash:     thash,
		PHash:     common.HexToHash("0x10001"),
		From:      common.HexToAddress("0x12345"),
		To:        common.HexToAddress("0x54321"),
		Tx:        new(types.Transaction),
		AssetCode: common.HexToHash("0x10002"),
		AssetId:   common.HexToHash("0x10003"),
	}
}

func NewTx10() []*Tx {
	result := make([]*Tx, 10)
	result[0] = NewTx(common.HexToHash("0x01"))
	result[1] = NewTx(common.HexToHash("0x02"))
	result[2] = NewTx(common.HexToHash("0x03"))
	result[3] = NewTx(common.HexToHash("0x04"))
	result[4] = NewTx(common.HexToHash("0x05"))
	result[5] = NewTx(common.HexToHash("0x06"))
	result[6] = NewTx(common.HexToHash("0x07"))
	result[7] = NewTx(common.HexToHash("0x08"))
	result[8] = NewTx(common.HexToHash("0x09"))
	result[9] = NewTx(common.HexToHash("0x10"))
	return result
}

func TestTxDao_Get(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, HOST_MYSQL)
	defer db.Close()
	defer db.Clear()

	txDao := NewTxDao(db)

	tx10 := NewTx10()
	err := txDao.Set(tx10[0])
	assert.NoError(t, err)

	result, err := txDao.Get(tx10[0].THash)
	assert.NoError(t, err)
	assert.Equal(t, result.BHash, tx10[0].BHash)

	err = txDao.Set(tx10[0])
	assert.NoError(t, err)

	result, err = txDao.Get(tx10[0].THash)
	assert.NoError(t, err)
	assert.Equal(t, result.BHash, tx10[0].BHash)
}

func TestTxDao_GetPage(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, HOST_MYSQL)
	defer db.Close()
	defer db.Clear()
	txDao := NewTxDao(db)

	tx10 := NewTx10()
	for index := 0; index < len(tx10); index++ {
		txDao.Set(tx10[index])
	}

	result, err := txDao.GetByAddr(tx10[0].From, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))

	result, err = txDao.GetByAddr(tx10[0].From, 10, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))

	result, err = txDao.GetByFrom(tx10[0].From, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))

	result, err = txDao.GetByFrom(tx10[0].From, 10, 5)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))

	result, err = txDao.GetByTo(tx10[0].To, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))

	result, err = txDao.GetByTo(tx10[0].To, 10, 5)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))

	result, err = txDao.GetByTime(tx10[0].To, 2553675430*1000, 0, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))

	result, err = txDao.GetByTime(tx10[0].To, 2553675430*1000, 0, 10, 5)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestTxDao_GetPateWithTotal(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, HOST_MYSQL)
	defer db.Close()
	defer db.Clear()
	txDao := NewTxDao(db)

	tx10 := NewTx10()
	for index := 0; index < len(tx10); index++ {
		txDao.Set(tx10[index])
	}

	result, total, err := txDao.GetByAddrWithTotal(tx10[0].From, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 5, len(result))

	result, total, err = txDao.GetByAddrWithTotal(tx10[0].From, 10, 1)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 0, len(result))

	result, total, err = txDao.GetByFromWithTotal(tx10[0].From, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 5, len(result))

	result, total, err = txDao.GetByFromWithTotal(tx10[0].From, 10, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 0, len(result))

	result, total, err = txDao.GetByToWithTotal(tx10[0].To, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 5, len(result))

	result, total, err = txDao.GetByToWithTotal(tx10[0].To, 10, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 0, len(result))

	result, total, err = txDao.GetByTimeWithTotal(tx10[0].To, 2553675430*1000, 0, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 5, len(result))

	result, total, err = txDao.GetByTimeWithTotal(tx10[0].To, 2553675430*1000, 0, 10, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 0, len(result))
}

func TestTxDao_NotExist(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, HOST_MYSQL)
	defer db.Close()
	defer db.Clear()
	txDao := NewTxDao(db)

	result, err := txDao.Get(common.HexToHash("0x01"))
	assert.Equal(t, ErrNotExist, err)
	assert.Nil(t, result)
}

func TestTxDao_ArgInvalid(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, HOST_MYSQL)
	defer db.Close()
	defer db.Clear()
	txDao := NewTxDao(db)

	result, err := txDao.Get(common.Hash{})
	assert.Equal(t, ErrArgInvalid, err)
	assert.Nil(t, result)

	err = txDao.Set(nil)
	assert.Equal(t, ErrArgInvalid, err)

	//
	result1, err := txDao.GetByAddr(common.Address{}, -1, 0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Nil(t, result1)

	result1, total, err := txDao.GetByAddrWithTotal(common.Address{}, -1, 0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Equal(t, -1, total)
	assert.Nil(t, result1)

	//
	result1, err = txDao.GetByFrom(common.Address{}, -1, 0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Nil(t, result1)

	result1, total, err = txDao.GetByFromWithTotal(common.Address{}, -1, 0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Equal(t, -1, total)
	assert.Nil(t, result1)

	//
	result1, err = txDao.GetByTo(common.Address{}, -1, 0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Nil(t, result1)

	result1, total, err = txDao.GetByToWithTotal(common.Address{}, -1, 0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Equal(t, -1, total)
	assert.Nil(t, result1)

	//
	result1, err = txDao.GetByTime(common.Address{}, -1, -1, -1, 0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Nil(t, result1)

	result1, total, err = txDao.GetByTimeWithTotal(common.Address{}, -1, -1, -1, 0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Equal(t, -1, total)
	assert.Nil(t, result1)
}

func TestTxDao_GetByAddressAndAssetCodeOrAssetIdWithTotal(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, HOST_MYSQL)
	defer db.Close()
	defer db.Clear()
	txDao := NewTxDao(db)

	tx10 := NewTx10()
	for index := 0; index < len(tx10); index++ {
		txDao.Set(tx10[index])
	}

	for i := 0; i < len(tx10); i++ {
		txs, total, err := txDao.GetByAddressAndAssetCodeOrAssetIdWithTotal(tx10[i].From, tx10[i].AssetCode, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 10, total)
		assert.Equal(t, 10, len(txs))
	}

}
