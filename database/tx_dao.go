package database

import (
	"database/sql"
	"github.com/meitu/go-ethereum/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
)

type Tx struct {
	BHash common.Hash
	THash common.Hash
	From  common.Address
	To    common.Address
	St    int64
	Tx  *types.Transaction
}

type TxDao struct {
	engine *sql.DB
}

func (dao *TxDao) Set(hash common.Hash, tx *Tx) (error) {
	if hash == (common.Hash{}) || (tx == nil) {
		return nil
	}

	sql := "REPLACE INTO t_tx(thash, bhash, from, to, tx, utc_st)VALUES(?,?,?,?,?,?)"

	val, err := rlp.EncodeToBytes(tx)
	if err != nil{
		return err
	}

	_, err = dao.engine.Exec(sql, hash.Hex(), tx.BHash.Hex(), tx.From.Hex(), tx.To.Hex(), val, tx.St)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (dao *TxDao) Get(hash common.Hash) (*Tx, error) {
	if hash == (common.Hash{}) {
		return nil, nil
	}

	sql := "SELECT tx FROM t_tx WHERE thash = ?"
	row := dao.engine.QueryRow(sql, hash.Hex())
	var val []byte
	err := row.Scan(&val)
	if err != nil {
		return nil, err
	}

	return dao.encodeTx(val)
}


func (dao *TxDao) encodeTx(val []byte) (*Tx, error) {
	var tx Tx
	err := rlp.DecodeBytes(val, &tx)
	if err != nil {
		return nil, err
	}else{
		return &tx, nil
	}
}

func (dao *TxDao) buildTxBatch(rows *sql.Rows) ([]*Tx, error) {
	result := make([]*Tx, 0)
	for rows.Next() {
		var val []byte
		var utcSt int64
		err := rows.Scan(&val, &utcSt)
		if err != nil {
			return nil, err
		}

		tx, err := dao.encodeTx(val)
		if err != nil{
			return nil, err
		}else{
			result = append(result, tx)
		}
	}
	return result, nil
}


func (dao *TxDao) GetByAddr(addr common.Address, start, stop int) ([]*Tx, error) {
	sqlQuery := "SELECT tx, utc_st FROM t_tx WHERE from = ? or to = ? ORDER BY utc_st DESC DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr, addr, start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByAddrWithTotal(addr common.Address, start, stop int) ([]*Tx, int, error) {
	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE from = ? or to = ?"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex(), addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByAddr(addr, start, stop)
	if err != nil{
		return nil, -1, err
	}else{
		return txes, cnt, nil
	}
}


func (dao *TxDao) GetByTime(addr common.Address, stStart, stStop int64, start, stop int) ([]*Tx, error) {
	sqlQuery := "SELECT tx, utc_st FROM t_tx WHERE (from = ? OR to = ?) AND utc_st > ? AND utc_st < ? ORDER BY utc_st DESC DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr, addr, stStart, stStop, start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByTimeWithTotal(addr common.Address, stStart, stStop int64, start, stop int) ([]*Tx, int, error) {
	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE (from = ? OR to = ?) AND utc_st > ? AND utc_st < ?"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex(), addr.Hex(), stStart, stStop)
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByTime(addr, stStart, stStop, start, stop)
	if err != nil{
		return nil, -1, err
	}else{
		return txes, cnt, nil
	}
}


func (dao *TxDao) GetByFrom(addr common.Address, start, stop int) ([]*Tx, error) {
	sqlQuery := "SELECT tx, utc_st FROM t_tx WHERE from = ? ORDER BY utc_st DESC DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr, addr, start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByFromWithTotal(addr common.Address, start, stop int) ([]*Tx, int, error) {
	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE from = ?"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByFrom(addr, start, stop)
	if err != nil{
		return nil, -1, err
	}else{
		return txes, cnt, nil
	}
}


func (dao *TxDao) GetByTo(addr common.Address, start, stop int) ([]*Tx, error) {
	sqlQuery := "SELECT tx, utc_st FROM t_tx WHERE to = ? ORDER BY utc_st DESC DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr, addr, start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByToWithTotal(addr common.Address, start, stop int) ([]*Tx, int, error) {
	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE to = ?"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByFrom(addr, start, stop)
	if err != nil{
		return nil, -1, err
	}else{
		return txes, cnt, nil
	}
}
