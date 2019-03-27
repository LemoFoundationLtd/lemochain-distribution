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
