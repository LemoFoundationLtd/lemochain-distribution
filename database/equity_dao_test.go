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