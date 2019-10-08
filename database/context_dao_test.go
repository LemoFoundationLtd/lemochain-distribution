package database

import (
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContextDao_ContextSet(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, HOST_MYSQL)
	defer db.Close()
	defer db.Clear()

	contextDao := NewContextDao(db)

	err := contextDao.ContextSet(ContextKeyCurrentBlock, common.HexToHash("0xabcdef").Bytes())
	assert.NoError(t, err)

	result, err := contextDao.ContextLoad()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, result[0].key, ContextKeyCurrentBlock)
	assert.Equal(t, result[0].val, common.HexToHash("0xabcdef").Bytes())

	contextItem, err := contextDao.ContextGet(ContextKeyCurrentBlock)
	assert.NoError(t, err)
	assert.Equal(t, contextItem, common.HexToHash("0xabcdef").Bytes())
}
