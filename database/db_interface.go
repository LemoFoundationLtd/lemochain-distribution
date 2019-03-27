package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
)

type DaoTransaction struct {
	Height uint32
	Hash   common.Hash
	TxData *types.Transaction
}

type DaoAccount struct {
	Account types.AccountData
}

type DBStore interface {
	/**
	 * （1） hash => block
	 * （2） hash => tx
	 * （3） token/code => attrs
	 * （4） address => account
	 */
	Get(key []byte) ([]byte, error)

	Set(key []byte, val []byte) (error)

	GetTxByAddress(address common.Address) (error)

	GetTxByTimeStamp(address common.Address, start int64, stop int64) (error)

	GetTokensByAddr(addr common.Address) (error)

	GetAccessKey()

	CurrentBlock() (*types.Block, error)

	Close()
}

type MySqlDB struct {
	engine *sql.DB
	driver string
	dns    string
}

func ErrIsNotExist(err error) bool {
	if err == sql.ErrNoRows{
		return true
	}else{
		return false
	}
}

func NewMySqlDB(driver string, dns string) *MySqlDB {
	db, err := Open(driver, dns)
	if err != nil {
		panic("open mysql err: " + err.Error())
	}

	return &MySqlDB{
		engine: db,
		driver: driver,
		dns:    dns,
	}
}

type Context struct {
	key []byte
	flg int
	val []byte
}

func (db *MySqlDB) ContextSet(key []byte, flg int, val []byte) (error) {
	_, err := db.engine.Exec("REPLACE INTO t_context(lm_key, lm_flg, lm_val) VALUES (?,?,?)", key, flg, val)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (db *MySqlDB) ContextLoad() ([]*Context, error) {
	rows, err := db.engine.Query("SELECT lm_key, lm_flg, lm_val FROM t_context")
	if err != nil {
		return nil, err
	}

	result := make([]*Context, 0)
	for rows.Next() {
		var context Context
		err := rows.Scan(&context.key, &context.flg, &context.val)
		if err != nil {
			return nil, err
		}

		result = append(result, &context)
	}

	return result, nil
}

func (db *MySqlDB) Get(key []byte) ([]byte, error) {
	row := db.engine.QueryRow("SELECT lm_val FROM t_kv WHERE lm_key = ?", key)
	var val []byte
	err := row.Scan(&val)
	if err != nil {
		return nil, err
	} else {
		return val, nil
	}
}

func (db *MySqlDB) Set(key []byte, val []byte) (error) {
	_, err := db.engine.Exec("REPLACE INTO t_kv(lm_key, lm_val) VALUES (?,?)", key, val)
	if err != nil {
		return err
	} else {
		return nil
	}
}


var txsql string = "SELECT tx_block_key, tx_val, tx_ver, tx_st FROM t_tx"

func (db *MySqlDB) TxSet(hash, blockHash, from, to string, val []byte, ver int64, st int64) error {
	_, err := db.engine.Exec("REPLACE INTO t_tx(tx_key, tx_block_key, tx_from, tx_to, tx_val, tx_ver, tx_st) VALUES (?,?,?,?,?,?,?)", hash, blockHash, from, to, val, ver, st)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (db *MySqlDB) TxGetByHash(key string) (string, []byte, int64, error) {
	row := db.engine.QueryRow(txsql+" WHERE tx_key = ?", key)
	var hash string
	var val []byte
	var ver int64
	var st int64
	err := row.Scan(&hash, &val, &ver, &st)
	if err == sql.ErrNoRows {
		return "", nil, -1, nil
	}

	if err != nil {
		return "", nil, -1, err
	}

	return hash, val, st, nil
}

func (db *MySqlDB) TxGetByAddr(addr string, index int, size int) ([]string, [][]byte, []int64, error) {
	stmt, err := db.engine.Prepare(txsql + " WHERE tx_from = ? or tx_to = ? ORDER BY tx_ver ASC LIMIT ?, ?")
	if err != nil {
		return nil, nil, nil, err
	}

	rows, err := stmt.Query(addr, addr, index, size)
	if err != nil {
		return nil, nil, nil, err
	}

	resultHash := make([]string, 0)
	resultVal := make([][]byte, 0)
	resultSt := make([]int64, 0)
	for rows.Next() {
		var hash string
		var val []byte
		var ver int64
		var st int64
		err := rows.Scan(&hash, &val, &ver, &st)
		if err != nil {
			return nil, nil, nil, err
		} else {
			resultHash = append(resultHash, hash)
			resultVal = append(resultVal, val)
			resultSt = append(resultSt, st)
		}
	}
	return resultHash, resultVal, resultSt, nil
}

func (db *MySqlDB) Clear() error {
	_, err := db.engine.Exec("DELETE FROM t_kv")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_tx")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_context")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_asset")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_equity")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_mate_data")
	if err != nil {
		return err
	}


	return nil
}

func (db *MySqlDB) Close() {
	db.engine.Close()
}
