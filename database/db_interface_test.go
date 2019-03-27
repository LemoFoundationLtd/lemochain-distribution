package database

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCreateDB(t *testing.T) {
	// db, _ := Open(DRIVER_MYSQL, DNS_MYSQL)
	// _, err := CreateDB(db)
	// assert.NoError(t, err)
}

func TestMySqlDB_Context(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	err := db.ContextSet([]byte("key"), 1, []byte("lemochain"))
	assert.NoError(t, err)

	result, err := db.ContextLoad()
	assert.Equal(t, 1, len(result))
	assert.Equal(t, []byte("lemochain"), result[0].val)
}

func TestMySqlDB_KeyVal(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	key := []byte("key")
	val := []byte("val")
	err := db.Set(key, val)
	assert.NoError(t, err)

	result, err := db.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, val, result)
}


