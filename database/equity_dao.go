package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"math/big"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"strconv"
	"time"
)

type EquityDao struct{
	engine *sql.DB
}

func NewEquityDao(engine *sql.DB) (*EquityDao) {
	return &EquityDao{engine:engine}
}

func (dao *EquityDao) Set(addr common.Address, assetEquity *types.AssetEquity) (error){
	if addr == (common.Address{}) || assetEquity == nil{
		log.Errorf("set equity.addr is common.Address{} or equity is nil.")
		return ErrArgInvalid
	}

	result, version, err := dao.query(addr, assetEquity.AssetId)
	if err != nil{
		return err
	}

	if result == nil {
		return dao.insert(addr, assetEquity)
	}else{
		return dao.update(addr, assetEquity, version)
	}
}

func (dao *EquityDao) Get(addr common.Address, id common.Hash) (*types.AssetEquity, error) {
	if (addr == common.Address{}) || (id == common.Hash{}) {
		log.Errorf("get asset equity.addr is common.address{} or id is common.hash{}")
		return nil, ErrArgInvalid
	}
	equity, _, err := dao.query(addr, id)
	if err != nil {
		return nil, err
	}

	if equity == nil{
		log.Errorf("get asset equity.is not exist.")
		return nil, ErrNotExist
	}else{
		return equity, nil
	}
}

func (dao *EquityDao) GetPage(addr common.Address, start, stop int) ([]*types.AssetEquity, error) {
	sql := "SELECT code, id, equity, utc_st WHERE addr = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr, start, stop)
	if err != nil {
		return nil, err
	}

	result := make([]*types.AssetEquity, 0)
	for rows.Next() {
		var assetEquity types.AssetEquity
		var equity int64
		err := rows.Scan(&assetEquity.AssetCode, &assetEquity.AssetId, &equity)
		if err != nil {
			return nil, err
		}

		assetEquity.Equity = new(big.Int).SetInt64(equity)
		result = append(result, &assetEquity)
	}

	return result, nil
}

func (dao *EquityDao) GetPageWithTotal(addr common.Address, start, stop int) ([]*types.AssetEquity, int, error) {
	sql := "SELECT count(*) as cnt FROM t_equity WHERE addr = ?"
	row := dao.engine.QueryRow(sql, addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	result, err := dao.GetPage(addr, start, stop)
	if err != nil{
		return nil, -1, err
	}else{
		return result, cnt, nil
	}
}

func (dao *EquityDao) query(addr common.Address, id common.Hash) (*types.AssetEquity, int, error) {
	sql := "SELECT code, equity, version FROM t_equity WHERE id = ? AND addr = ?"
	row := dao.engine.QueryRow(sql, id.Hex(), addr.Hex())
	var code []byte
	var equity int64
	var version int
	err := row.Scan(&code, &equity, &version)
	if ErrIsNotExist(err) {
		return nil, -1, nil
	}

	if err != nil {
		return nil, -1, err
	}

	return &types.AssetEquity{
		AssetCode: common.BytesToHash(code),
		AssetId:id,
		Equity:new(big.Int).SetInt64(equity),
	}, version, nil
}

func (dao *EquityDao) insert(addr common.Address, assetEquity *types.AssetEquity) (error) {
	sql := "INSERT INTO t_equity(code, id, addr, equity, utc_st, version)VALUES(?,?,?,?,?,?)"
	code := assetEquity.AssetCode
	id := assetEquity.AssetId
	equity := assetEquity.Equity
	result, err := dao.engine.Exec(sql, code.Hex(), id.Hex(), addr.Hex(), equity.Int64(), time.Now().UnixNano() / 1000000, 1)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if effected != 1{
		log.Errorf("insert equity.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	}else{
		return nil
	}
}

func (dao *EquityDao) update(addr common.Address, assetEquity *types.AssetEquity, version int) (error) {
	sql := "UPDATE t_equity SET equity = ?, version = version + 1 WHERE id = ? AND version = ?"
	result, err := dao.engine.Exec(sql,  assetEquity.Equity.Int64(), assetEquity.AssetId.Hex(), version)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if effected != 1{
		log.Errorf("update equity.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	}else{
		return nil
	}
}