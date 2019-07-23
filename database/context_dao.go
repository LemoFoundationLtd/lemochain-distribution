package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"strconv"
)

type Context struct {
	key string
	val []byte
}

var (
	ContextKeyCurrentBlock = "context.chain.current_block"
)

type ContextDao struct {
	engine *sql.DB
}

func NewContextDao(db DBEngine) *ContextDao {
	return &ContextDao{engine: db.GetDB()}
}

func (dao *ContextDao) GetCurrentBlock() (*types.Block, error) {
	val, err := dao.ContextGet(ContextKeyCurrentBlock)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, ErrNotExist
	}

	var block types.Block
	err = rlp.DecodeBytes(val, &block)
	if err != nil {
		log.Errorf("rlp block[%d] error.", block.Height())
		return nil, err
	} else {
		return &block, nil
	}
}

func (dao *ContextDao) SetCurrentBlock(block *types.Block) error {
	if block == nil {
		log.Errorf("set current block. block is ni.")
		return ErrArgInvalid
	}

	val, err := rlp.EncodeToBytes(block)
	if err != nil {
		return err
	} else {
		return dao.ContextSet(ContextKeyCurrentBlock, val)
	}
}

func (dao *ContextDao) ContextSet(key string, val []byte) error {
	result, err := dao.engine.Exec("REPLACE INTO t_context(lm_key, lm_val) VALUES (?,?)", key, val)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if effected < 1 {
		log.Errorf("insert context.affected: " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	} else {
		return nil
	}
}

func (dao *ContextDao) ContextGet(key string) ([]byte, error) {
	row := dao.engine.QueryRow("SELECT lm_val FROM t_context WHERE lm_key = ?", key)
	var val []byte
	err := row.Scan(&val)
	if ErrIsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	} else {
		return val, nil
	}
}

func (dao *ContextDao) ContextLoad() ([]*Context, error) {
	rows, err := dao.engine.Query("SELECT lm_key, lm_val FROM t_context")
	if err != nil {
		return nil, err
	}

	result := make([]*Context, 0)
	for rows.Next() {
		var key string
		var val []byte
		err := rows.Scan(&key, &val)
		if err != nil {
			return nil, err
		}

		result = append(result, &Context{
			key: key,
			val: val,
		})
	}

	return result, nil
}
