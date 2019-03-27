package database

import (
	"testing"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/stretchr/testify/assert"
)

func TestContextDao_ContextSet(t *testing.T) {
	db := NewMySqlDB(DRIVER_MYSQL, DNS_MYSQL)
	defer db.Close()
	defer db.Clear()

	contextDao := NewContextDao(db.engine)

	err := contextDao.ContextSet(ContextFlgCurrentBlock, common.HexToHash("0x123456").Bytes(), common.HexToHash("0xabcdef").Bytes())
	assert.NoError(t, err)

	result, err := contextDao.ContextLoad()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, result[0].key, common.HexToHash("0x123456").Bytes())
	assert.Equal(t, result[0].val, common.HexToHash("0xabcdef").Bytes())
	assert.Equal(t, result[0].flg, ContextFlgCurrentBlock)
}
