package database

import (
	"testing"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/stretchr/testify/assert"
)

func NewMateData(id common.Hash, code common.Hash, addr common.Address, isNil bool) (*MateData) {
	result := &MateData{
		Id:   id,
		Code: code,
		Addr: addr,
	}

	if !isNil {
		profile := make(types.Profile)
		profile["key1"] = "val1"
		profile["key2"] = "val2"
		profile["key3"] = "val3"
		result.Profile = "profile"
	}

	return result
}

func NewMateDataBatch1(id common.Hash, isNil bool) (*MateData) {
	return NewMateData(id, common.HexToHash("0x0abcd"), common.HexToAddress("0x01234"), isNil)
}

func NewMateDataBatch20() ([]common.Hash) {
	result := make([]common.Hash, 20)
	result[0] = common.HexToHash("0x023456789")
	result[1] = common.HexToHash("0x123456789")
	result[2] = common.HexToHash("0x223456789")
	result[3] = common.HexToHash("0x323456789")
	result[4] = common.HexToHash("0x423456789")
	result[5] = common.HexToHash("0x523456789")
	result[6] = common.HexToHash("0x623456789")
	result[7] = common.HexToHash("0x723456789")
	result[8] = common.HexToHash("0x823456789")
	result[9] = common.HexToHash("0x923456789")
	result[10] = common.HexToHash("0x1023456789")
	result[11] = common.HexToHash("0x1123456789")
	result[12] = common.HexToHash("0x1223456789")
	result[13] = common.HexToHash("0x1323456789")
	result[14] = common.HexToHash("0x1423456789")
	result[15] = common.HexToHash("0x1523456789")
	result[16] = common.HexToHash("0x1623456789")
	result[17] = common.HexToHash("0x1723456789")
	result[18] = common.HexToHash("0x1823456789")
	result[19] = common.HexToHash("0x1923456789")
	return result
}

func TestMateDataDao_Get(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	mateDataDao := NewMateDataDao(db)

	ids := NewMateDataBatch20()
	data := NewMateDataBatch1(ids[0], false)

	err := mateDataDao.Set(data)
	assert.NoError(t, err)
	result, err := mateDataDao.Get(ids[0])
	assert.NoError(t, err)
	assert.Equal(t, data, result)

	err = mateDataDao.Set(data)
	assert.NoError(t, err)
	result, err = mateDataDao.Get(ids[0])
	assert.NoError(t, err)
	assert.Equal(t, data, result)

	// profile is nil.
	data = NewMateDataBatch1(ids[0], true)
	err = mateDataDao.Set(data)
	assert.NoError(t, err)
	result, err = mateDataDao.Get(ids[0])
	assert.NoError(t, err)
	assert.Equal(t, data, result)

	err = mateDataDao.Set(data)
	assert.NoError(t, err)
	result, err = mateDataDao.Get(ids[0])
	assert.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestMateDataDao_GetPage(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	mateDataDao := NewMateDataDao(db)

	ids := NewMateDataBatch20()
	for index := 0; index < len(ids); index++{
		data := NewMateDataBatch1(ids[index], false)
		mateDataDao.Set(data)
	}

	data := NewMateDataBatch1(ids[0], false)
	result, err := mateDataDao.GetPage(data.Addr, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))

	result, err = mateDataDao.GetPage(data.Addr, 20, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestMateDataDao_GetPageWithTotal(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	mateDataDao := NewMateDataDao(db)

	ids := NewMateDataBatch20()
	for index := 0; index < len(ids); index++{
		data := NewMateDataBatch1(ids[index], false)
		mateDataDao.Set(data)
	}

	data := NewMateDataBatch1(ids[0], false)
	result, total, err := mateDataDao.GetPageWithTotal(data.Addr, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 20, total)
	assert.Equal(t, 5, len(result))

	result, total, err = mateDataDao.GetPageWithTotal(data.Addr, 20, 1)
	assert.NoError(t, err)
	assert.Equal(t, 20, total)
	assert.Equal(t, 0, len(result))
}

func TestMateDataDao_GetPageByCode(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	mateDataDao := NewMateDataDao(db)

	ids := NewMateDataBatch20()
	for index := 0; index < len(ids); index++{
		data := NewMateDataBatch1(ids[index], false)
		mateDataDao.Set(data)
	}

	data := NewMateDataBatch1(ids[0], false)
	result, err := mateDataDao.GetPageByCode(data.Code, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))

	result, err = mateDataDao.GetPageByCode(data.Code, 20, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestMateDataDao_GetPageByCodeWithTotal(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	mateDataDao := NewMateDataDao(db)

	ids := NewMateDataBatch20()
	for index := 0; index < len(ids); index++{
		data := NewMateDataBatch1(ids[index], false)
		mateDataDao.Set(data)
	}

	data := NewMateDataBatch1(ids[0], false)
	result, total, err := mateDataDao.GetPageByCodeWithTotal(data.Code, 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, 20, total)
	assert.Equal(t, 5, len(result))

	result, total, err = mateDataDao.GetPageByCodeWithTotal(data.Code, 20, 1)
	assert.NoError(t, err)
	assert.Equal(t, 20, total)
	assert.Equal(t, 0, len(result))
}

func TestMateDataDao_NotExist(t *testing.T){
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	mateDataDao := NewMateDataDao(db)

	result, err := mateDataDao.Get(common.HexToHash("0x01"))
	assert.Equal(t, err, ErrNotExist)
	assert.Nil(t, result)
}

func TestMateDataDao_ArgInvalid(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	mateDataDao := NewMateDataDao(db)

	result1, err := mateDataDao.Get(common.Hash{})
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result1)

	err = mateDataDao.Set(nil)
	assert.Equal(t, err, ErrArgInvalid)

	result2, err := mateDataDao.GetPage(common.Address{}, -1, 0)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result2)

	result2, total, err := mateDataDao.GetPageWithTotal(common.Address{}, -1, 0)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Equal(t, -1, total)
	assert.Nil(t, result2)

	result2, err = mateDataDao.GetPageByCode(common.Hash{}, -1, 0)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result2)

	result2, total, err = mateDataDao.GetPageByCodeWithTotal(common.Hash{}, -1, 0)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Equal(t, -1, total)
	assert.Nil(t, result2)
}
