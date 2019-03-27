package database

import (
	"testing"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/stretchr/testify/assert"
)

func NewAsset(code common.Hash, addr common.Address) (*types.Asset) {
	return &types.Asset{
		AssetCode: code,
		Issuer:addr,
	}
}

func NewAsset10() ([]*types.Asset) {
	result := make([]*types.Asset, 10)
	result[0] = NewAsset(common.HexToHash("0x0bcde"), common.HexToAddress("0x02345"))
	result[1] = NewAsset(common.HexToHash("0x1bcde"), common.HexToAddress("0x02345"))
	result[2] = NewAsset(common.HexToHash("0x2bcde"), common.HexToAddress("0x02345"))
	result[3] = NewAsset(common.HexToHash("0x3bcde"), common.HexToAddress("0x02345"))
	result[4] = NewAsset(common.HexToHash("0x4bcde"), common.HexToAddress("0x02345"))
	result[5] = NewAsset(common.HexToHash("0x5bcde"), common.HexToAddress("0x52345"))
	result[6] = NewAsset(common.HexToHash("0x6bcde"), common.HexToAddress("0x52345"))
	result[7] = NewAsset(common.HexToHash("0x7bcde"), common.HexToAddress("0x52345"))
	result[8] = NewAsset(common.HexToHash("0x8bcde"), common.HexToAddress("0x52345"))
	result[9] = NewAsset(common.HexToHash("0x9bcde"), common.HexToAddress("0x52345"))
	return result
}

func TestAssetDao_Set(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	assetes := NewAsset10()

	assetDao := NewAssetDao(db.engine)
	err := assetDao.Set(assetes[0])		// insert
	assert.NoError(t, err)

	result, err := assetDao.Get(assetes[0].AssetCode)
	assert.NoError(t, err)
	assert.Equal(t, result.Issuer, assetes[0].Issuer)

	err = assetDao.Set(assetes[0])		// update
	assert.NoError(t, err)

	err = assetDao.Set(assetes[0])		// update
	assert.NoError(t, err)
}

func TestAssetDao_GetPage(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	assetDao := NewAssetDao(db.engine)

	assetes := NewAsset10()
	for index := 0; index < len(assetes); index++{
		err := assetDao.Set(assetes[index])
		assert.NoError(t, err)
	}

	result, err := assetDao.GetPage(assetes[0].Issuer, 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	result, err = assetDao.GetPage(assetes[0].Issuer, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))

	result, err = assetDao.GetPage(assetes[5].Issuer, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))
}

func TestAssetDao_GetPageWithTotal(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	assetDao := NewAssetDao(db.engine)

	assetes := NewAsset10()
	for index := 0; index < len(assetes); index++ {
		err := assetDao.Set(assetes[index])
		assert.NoError(t, err)
	}

	result, total, err := assetDao.GetPageWithTotal(assetes[0].Issuer, 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 5, total)

	result, total, err = assetDao.GetPageWithTotal(assetes[0].Issuer, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))
	assert.Equal(t, 5, total)

	result, total, err = assetDao.GetPageWithTotal(assetes[5].Issuer, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))
	assert.Equal(t, 5, total)

	result, total, err = assetDao.GetPageWithTotal(common.HexToAddress("0x123456789"), 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, 0, total)
}

func TestAssetDao_GetNotExist(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	assetDao := NewAssetDao(db.engine)

	assetes := NewAsset10()
	for index := 0; index < len(assetes); index++{
		err := assetDao.Set(assetes[index])
		assert.NoError(t, err)
	}

	result, err := assetDao.GetPage(common.HexToAddress("0x01"), 0, 2)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))

	result, total, err := assetDao.GetPageWithTotal(common.HexToAddress("0x01"), 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestAssetDao_GetArgInvalid(t *testing.T){
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	assetDao := NewAssetDao(db.engine)

	err := assetDao.Set(nil)
	assert.Equal(t, err, ErrArgInvalid)

	result1, err := assetDao.Get(common.Hash{})
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result1)

	result2, err := assetDao.GetPage(common.Address{}, 0, 2)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result2)

	result2, err = assetDao.GetPage(common.HexToAddress("0x01"), -1, 2)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result2)

	result2, err = assetDao.GetPage(common.HexToAddress("0x01"), 1, 0)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result2)

	result3, total, err := assetDao.GetPageWithTotal(common.Address{}, 0, 2)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result3)
	assert.Equal(t, -1, total)

	result3, total, err = assetDao.GetPageWithTotal(common.HexToAddress("0x01"), -1, 2)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result3)
	assert.Equal(t, -1, total)

	result3, total, err = assetDao.GetPageWithTotal(common.HexToAddress("0x01"), 0, 0)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result3)
	assert.Equal(t, -1, total)
}