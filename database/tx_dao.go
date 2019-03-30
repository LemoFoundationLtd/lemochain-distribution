package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"time"
)

type Tx struct {
	BHash common.Hash
	THash common.Hash
	From  common.Address
	To    common.Address
	Tx  *types.Transaction
}

type TxDao struct {
	engine *sql.DB
}

func NewTxDao(db DBEngine) (*TxDao) {
	return &TxDao{engine:db.GetDB()}
}

func (dao *TxDao) Set(tx *Tx) (error) {
	if tx == nil {
		log.Errorf("set tx.tx is nil.")
		return ErrArgInvalid
	}

	sql := "REPLACE INTO t_tx(thash, bhash, faddr, taddr, tx, utc_st)VALUES(?,?,?,?,?,?)"

	val, err := rlp.EncodeToBytes(tx.Tx)
	if err != nil{
		return err
	}

	_, err = dao.engine.Exec(sql, tx.THash.Hex(), tx.BHash.Hex(), tx.From.Hex(), tx.To.Hex(), val, time.Now().UnixNano() / 1000000)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (dao *TxDao) Get(hash common.Hash) (*Tx, error) {
	if hash == (common.Hash{}) {
		log.Errorf("get tx.hash is common.hash{}")
		return nil, ErrArgInvalid
	}

	sql := "SELECT thash, bhash, faddr, taddr, tx FROM t_tx WHERE thash = ?"
	row := dao.engine.QueryRow(sql, hash.Hex())
	var thash string
	var bhash string
	var faddr string
	var taddr string
	var val []byte
	err := row.Scan(&thash, &bhash, &faddr, &taddr, &val)
	if ErrIsNotExist(err) {
		return nil, ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	tx, err := dao.encodeTx(val)
	if err != nil{
		return nil, err
	}else{
		return &Tx{
			THash: common.HexToHash(thash),
			BHash:common.HexToHash(bhash),
			From:common.HexToAddress(faddr),
			To:common.HexToAddress(taddr),
			Tx:tx,
		}, nil
	}
}


func (dao *TxDao) encodeTx(val []byte) (*types.Transaction, error) {
	var tx types.Transaction
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
		var thash string
		var bhash string
		var faddr string
		var taddr string
		var val []byte
		var utcSt int64
		err := rows.Scan(&thash, &bhash, &faddr, &taddr, &val, &utcSt)
		if err != nil {
			return nil, err
		}

		tx, err := dao.encodeTx(val)
		if err != nil{
			return nil, err
		}else{
			result = append(result, &Tx{
				THash: common.HexToHash(thash),
				BHash:common.HexToHash(bhash),
				From:common.HexToAddress(faddr),
				To:common.HexToAddress(taddr),
				Tx:tx,
			})
		}
	}
	return result, nil
}


func (dao *TxDao) GetByAddr(addr common.Address, start, limit int) ([]*Tx, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by addr.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sqlQuery := "SELECT thash, bhash, faddr, taddr, tx, utc_st FROM t_tx WHERE faddr = ? or taddr = ? ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), addr.Hex(), start, start + limit)
	if err != nil {
		return nil, err
	}

	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByAddrWithTotal(addr common.Address, start, limit int) ([]*Tx, int, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by addr with total.addr is common.address{} or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE faddr = ? or taddr = ?"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex(), addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByAddr(addr, start, limit)
	if err != nil{
		return nil, -1, err
	}else{
		return txes, cnt, nil
	}
}


func (dao *TxDao) GetByTime(addr common.Address, stStart, stStop int64, start, limit int) ([]*Tx, error) {
	if addr == (common.Address{}) || (stStart < 0) || (stStop < 0) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by time.addr is common.address{} or time stamp < 0 or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sqlQuery := "SELECT thash, bhash, faddr, taddr, tx, utc_st FROM t_tx WHERE (faddr = ? OR taddr = ?) AND utc_st > ? AND utc_st < ? ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), addr.Hex(), stStop, stStart, start, start + limit)
	if err != nil {
		return nil, err
	}

	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByTimeWithTotal(addr common.Address, stStart, stStop int64, start, limit int) ([]*Tx, int, error) {
	if addr == (common.Address{}) || (stStart < 0) || (stStop < 0) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by time with total.addr is common.address{} or time stamp < 0 or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE (faddr = ? OR taddr = ?) AND utc_st > ? AND utc_st < ?"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex(), addr.Hex(), stStop, stStart)
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByTime(addr, stStart, stStop, start, limit)
	if err != nil{
		return nil, -1, err
	}else{
		return txes, cnt, nil
	}
}


func (dao *TxDao) GetByFrom(addr common.Address, start, limit int) ([]*Tx, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by from.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sqlQuery := "SELECT thash, bhash, faddr, taddr, tx, utc_st FROM t_tx WHERE faddr = ? ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), start, start + limit)
	if err != nil {
		return nil, err
	}

	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByFromWithTotal(addr common.Address, start, limit int) ([]*Tx, int, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by from with total.addr is common.address{} or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE faddr = ?"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByFrom(addr, start, limit)
	if err != nil{
		return nil, -1, err
	}else{
		return txes, cnt, nil
	}
}


func (dao *TxDao) GetByTo(addr common.Address, start, limit int) ([]*Tx, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by to.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sqlQuery := "SELECT thash, bhash, faddr, taddr, tx, utc_st FROM t_tx WHERE taddr = ? ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), start, start + limit)
	if err != nil {
		return nil, err
	}

	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByToWithTotal(addr common.Address, start, limit int) ([]*Tx, int, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by to with total.addr is common.address{} or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE taddr = ?"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByTo(addr, start, limit)
	if err != nil{
		return nil, -1, err
	}else{
		return txes, cnt, nil
	}
}
