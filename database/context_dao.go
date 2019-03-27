package database

import (
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"strconv"
	"database/sql"
	"github.com/meitu/go-ethereum/common"
)

type Context struct {
	key []byte
	flg int
	val []byte
}

var (
	ContextFlgCurrentBlock = 1
)

type ContextDao struct{
	engine *sql.DB
}

func NewContextDao(engine *sql.DB) (*ContextDao) {
	return &ContextDao{engine:engine}
}

func (dao *ContextDao) ContextSet(flg int, key []byte, val []byte) (error) {
	result, err := dao.engine.Exec("REPLACE INTO t_context(lm_key, lm_flg, lm_val) VALUES (?,?,?)", common.ToHex(key), flg, val)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if effected < 1{
		log.Errorf("insert context flg: " + strconv.Itoa(flg) + "|affected: " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	}else{
		return nil
	}
}

func (dao *ContextDao) ContextLoad() ([]*Context, error) {
	rows, err := dao.engine.Query("SELECT lm_key, lm_flg, lm_val FROM t_context")
	if err != nil {
		return nil, err
	}

	result := make([]*Context, 0)
	for rows.Next() {
		var key string
		var val []byte
		var flg int
		err := rows.Scan(&key, &flg, &val)
		if err != nil {
			return nil, err
		}

		result = append(result, &Context{
			key:common.FromHex(key),
			flg:flg,
			val:val,
		})
	}

	return result, nil
}
