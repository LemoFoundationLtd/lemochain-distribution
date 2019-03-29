package database

import (
	"testing"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"strconv"
	"math/big"
	"github.com/stretchr/testify/assert"
)

func NewCandidates50() []*CandidateItem{
	result := make([]*CandidateItem, 0)
	for index := 0; index < 50; index++{
		result = append(result, &CandidateItem{
			User: common.HexToAddress(strconv.Itoa(100 + index)),
			Votes: new(big.Int).SetInt64(int64(index)),
		})
	}
	return result
}

func TestCandidateDao_Set(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	candidateDao := NewCandidateDao(db)

	candidates := NewCandidates50()
	err := candidateDao.Set(candidates[0])
	assert.NoError(t, err)

	result, total, err := candidateDao.GetPageWithTotal(0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, total)

	err = candidateDao.Del(candidates[0].User)
	assert.NoError(t, err)

	result, total, err = candidateDao.GetPageWithTotal(0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, 0, total)


	for index := 0; index < len(candidates); index++{
		candidateDao.Set(candidates[index])
	}

	result, err = candidateDao.GetTop(30)
	assert.NoError(t, err)
	assert.Equal(t, 30, len(result))
	assert.Equal(t, result[0], candidates[49])

	result, total, err = candidateDao.GetPageWithTotal(0, 40)
	assert.NoError(t, err)
	assert.Equal(t, 40, len(result))
	assert.Equal(t, 50, total)
	assert.Equal(t, result[0], candidates[49])

	result, total, err = candidateDao.GetPageWithTotal(40, 60)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(result))
	assert.Equal(t, 50, total)
	assert.Equal(t, result[0], candidates[9])

	result, total, err = candidateDao.GetPageWithTotal(60, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, 50, total)
}

func TestCandidateDao_ArgInvalid(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	candidateDao := NewCandidateDao(db)

	err := candidateDao.Set(nil)
	assert.Equal(t, ErrArgInvalid, err)

	err = candidateDao.Del(common.Address{})
	assert.Equal(t, ErrArgInvalid, err)

	result, err := candidateDao.GetTop(0)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Nil(t, result)

	result, err = candidateDao.GetPage(-1, -1)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Nil(t, result)

	result, total, err := candidateDao.GetPageWithTotal(-1, -1)
	assert.Equal(t, ErrArgInvalid, err)
	assert.Nil(t, result)
	assert.Equal(t, -1, total)
}
