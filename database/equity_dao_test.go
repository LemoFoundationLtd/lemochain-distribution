package database

import (
	"testing"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"math/big"
	"github.com/stretchr/testify/assert"
)

func NewAssetEquity(code common.Hash, id common.Hash, equity int64) (*types.AssetEquity){
	return &types.AssetEquity{
		AssetCode:code,
		AssetId:id,
		Equity:new(big.Int).SetInt64(equity),
	}
}

func NewAssetEquity20() ([]*types.AssetEquity) {
	result := make([]*types.AssetEquity, 20)
	result[0] = NewAssetEquity(common.HexToHash("0x0abcd"), common.HexToHash("0x01234"), 50)
	result[1] = NewAssetEquity(common.HexToHash("0x1abcd"), common.HexToHash("0x11234"), 50)
	result[2] = NewAssetEquity(common.HexToHash("0x2abcd"), common.HexToHash("0x21234"), 50)
	result[3] = NewAssetEquity(common.HexToHash("0x3abcd"), common.HexToHash("0x31234"), 50)
	result[4] = NewAssetEquity(common.HexToHash("0x4abcd"), common.HexToHash("0x41234"), 50)
	result[5] = NewAssetEquity(common.HexToHash("0x5abcd"), common.HexToHash("0x51234"), 50)
	result[6] = NewAssetEquity(common.HexToHash("0x6abcd"), common.HexToHash("0x61234"), 50)
	result[7] = NewAssetEquity(common.HexToHash("0x7abcd"), common.HexToHash("0x71234"), 50)
	result[8] = NewAssetEquity(common.HexToHash("0x8abcd"), common.HexToHash("0x81234"), 50)
	result[9] = NewAssetEquity(common.HexToHash("0x9abcd"), common.HexToHash("0x91234"), 50)
	result[10] = NewAssetEquity(common.HexToHash("0x0abcd"), common.HexToHash("0x05678"), 50)
	result[11] = NewAssetEquity(common.HexToHash("0x1abcd"), common.HexToHash("0x15678"), 50)
	result[12] = NewAssetEquity(common.HexToHash("0x2abcd"), common.HexToHash("0x25678"), 50)
	result[13] = NewAssetEquity(common.HexToHash("0x3abcd"), common.HexToHash("0x35678"), 50)
	result[14] = NewAssetEquity(common.HexToHash("0x4abcd"), common.HexToHash("0x45678"), 50)
	result[15] = NewAssetEquity(common.HexToHash("0x5abcd"), common.HexToHash("0x55678"), 50)
	result[16] = NewAssetEquity(common.HexToHash("0x6abcd"), common.HexToHash("0x65678"), 50)
	result[17] = NewAssetEquity(common.HexToHash("0x7abcd"), common.HexToHash("0x75678"), 50)
	result[18] = NewAssetEquity(common.HexToHash("0x8abcd"), common.HexToHash("0x85678"), 50)
	result[19] = NewAssetEquity(common.HexToHash("0x9abcd"), common.HexToHash("0x95678"), 50)
	return result
}

func TestEquityDao_Get(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	equities := NewAssetEquity20()

	equityDao := NewEquityDao(db.engine)
	err := equityDao.Set(common.HexToAddress("0x01"), equities[0])
	assert.NoError(t, err)

	result, err := equityDao.Get(common.HexToAddress("0x01"), equities[0].AssetId)
	assert.NoError(t, err)
	assert.Equal(t, equities[0].Equity, result.Equity)
	assert.Equal(t, equities[0].AssetId, result.AssetId)

	err = equityDao.Set(common.HexToAddress("0x01"), equities[0])
	assert.NoError(t, err)

	result, err = equityDao.Get(common.HexToAddress("0x01"), equities[0].AssetId)
	assert.NoError(t, err)
	assert.Equal(t, equities[0].Equity, result.Equity)
	assert.Equal(t, equities[0].AssetId, result.AssetId)
}

func TestEquityDao_GetPage(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	equityDao := NewEquityDao(db.engine)

	equities := NewAssetEquity20()
	for index := 0; index < 10; index++{
		equityDao.Set(common.HexToAddress("0x01"), equities[index])
	}

	for index := 10; index < 20; index++{
		equityDao.Set(common.HexToAddress("0x02"), equities[index])
	}

	result, err := equityDao.GetPage(common.HexToAddress("0x01"), 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))

	result, err = equityDao.GetPage(common.HexToAddress("0x01"), 10, 5)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestEquityDao_GetPageWithTotal(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	equityDao := NewEquityDao(db.engine)

	equities := NewAssetEquity20()
	for index := 0; index < 10; index++{
		equityDao.Set(common.HexToAddress("0x01"), equities[index])
	}

	for index := 10; index < 20; index++{
		equityDao.Set(common.HexToAddress("0x02"), equities[index])
	}

	result, total, err := equityDao.GetPageWithTotal(common.HexToAddress("0x01"), 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 5, len(result))

	result, total, err = equityDao.GetPageWithTotal(common.HexToAddress("0x01"), 10, 1)
	assert.NoError(t, err)
	assert.Equal(t, 10, total)
	assert.Equal(t, 0, len(result))
}

func TestEquityDao_NotExist(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	equityDao := NewEquityDao(db.engine)

	equity, err := equityDao.Get(common.HexToAddress("0xab"), common.HexToHash("0x01"))
	assert.Equal(t, err, ErrNotExist)
	assert.Nil(t, equity)
}

func TestEquityDao_ArgInvaild(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	equityDao := NewEquityDao(db.engine)

	err := equityDao.Set(common.Address{}, new(types.AssetEquity))
	assert.Equal(t, err, ErrArgInvalid)

	err = equityDao.Set(common.HexToAddress("0x01"), nil)
	assert.Equal(t, err, ErrArgInvalid)

	result1, err := equityDao.Get(common.Address{}, common.HexToHash("0x01"))
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result1)

	result1, err = equityDao.Get(common.HexToAddress("0x01"), common.Hash{})
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result1)

	result2, err := equityDao.GetPage(common.Address{}, 0, 1)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result2)

	result2, err = equityDao.GetPage(common.HexToAddress("0x01"), -1, 1)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result2)

	result2, err = equityDao.GetPage(common.HexToAddress("0x01"), 0, 0)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result2)

	result3, total, err := equityDao.GetPageWithTotal(common.Address{}, 0, 1)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Equal(t, -1, total)
	assert.Nil(t, result3)

	result3, total, err = equityDao.GetPageWithTotal(common.HexToAddress("0x01"), -1, 1)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Equal(t, -1, total)
	assert.Nil(t, result3)

	result3, total, err = equityDao.GetPageWithTotal(common.HexToAddress("0x01"), 0, 0)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Equal(t, -1, total)
	assert.Nil(t, result3)
}