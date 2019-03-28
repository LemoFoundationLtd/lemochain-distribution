package database

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestKvDao_Get(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	kvDao := NewKvDao(db)

	key := []byte("key")
	val := []byte("val")
	err := kvDao.Set(key, val)
	assert.NoError(t, err)

	result, err := kvDao.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, val, result)
}

func TestKvDao_NotExist(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	kvDao := NewKvDao(db)

	result, err := kvDao.Get([]byte("lemo"))
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestKvDao_ArgInvalid(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()
	kvDao := NewKvDao(db)

	result, err := kvDao.Get(nil)
	assert.Equal(t, err, ErrArgInvalid)
	assert.Nil(t, result)

	err = kvDao.Set(nil, []byte("lemo"))
	assert.Equal(t, err, ErrArgInvalid)
}
