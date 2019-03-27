package database

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestKvDao_Get(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	kvDao := NewKvDao(db.engine)

	key := []byte("key")
	val := []byte("val")
	err := kvDao.Set(key, val)
	assert.NoError(t, err)

	result, err := kvDao.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, val, result)
}
