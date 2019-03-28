package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"time"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"strconv"
)

type AssetDao struct{
	engine *sql.DB
}

func NewAssetDao(db DBEngine) (*AssetDao) {
	return &AssetDao{engine:db.GetDB()}
}

func (dao *AssetDao) Set(asset *types.Asset) (error) {
	if asset == nil {
		log.Errorf("set asset.asset is nil.")
		return ErrArgInvalid
	}

	code := asset.AssetCode
	result, ver, err := dao.query(code)
	if err != nil{
		return err
	}

	if result == nil {
		return dao.insert(asset)
	}else{
		return dao.update(asset, ver)
	}
}

func (dao *AssetDao) Get(code common.Hash) (*types.Asset, error) {
	if code == (common.Hash{}) {
		log.Errorf("get asset.code is common.hash{}")
		return nil, ErrArgInvalid
	}

	asset, _, err := dao.query(code)
	if err != nil{
		return nil, err
	}else{
		return asset, nil
	}
}

func (dao *AssetDao) decodeAsset(val []byte)(*types.Asset, error) {
	var asset types.Asset
	err := rlp.DecodeBytes(val, &asset)
	if err != nil{
		return nil, err
	}else{
		return &asset, nil
	}
}

func (dao *AssetDao) buildAssetBatch(rows *sql.Rows) ([]*types.Asset, error) {
	result := make([]*types.Asset, 0)
	for rows.Next() {
		var val []byte
		var utcSt int64
		err := rows.Scan(&val, &utcSt)
		if err != nil {
			return nil, err
		}

		asset, err := dao.decodeAsset(val)
		if err != nil{
			return nil, err
		}else{
			result = append(result, asset)
		}
	}
	return result, nil
}

func (dao *AssetDao) GetPage(addr common.Address, start, stop int) ([]*types.Asset, error) {
	if (addr == (common.Address{}))  || (start < 0) || (stop <= 0) {
		log.Errorf("get asset by page.addr is common.address{} or start < 0 or stop <= 0")
		return nil, ErrArgInvalid
	}

	sql := "SELECT attrs, utc_st FROM t_asset WHERE addr = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildAssetBatch(rows)
}

func (dao *AssetDao) GetPageWithTotal(addr common.Address, start, stop int) ([]*types.Asset, int, error) {
	if (addr == (common.Address{}))  || (start < 0) || (stop <= 0) {
		log.Errorf("get asset by page.addr is common.address{} or start < 0 or stop <= 0")
		return nil, -1, ErrArgInvalid
	}

	sql := "SELECT count(*) as cnt FROM t_asset WHERE addr = ?"
	row := dao.engine.QueryRow(sql, addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	assets, err := dao.GetPage(addr, start, stop)
	if err != nil{
		return nil, -1, err
	}else{
		return assets, cnt, nil
	}
}

func (dao *AssetDao) query(code common.Hash) (*types.Asset, int, error) {
	row := dao.engine.QueryRow("SELECT attrs, version FROM t_asset WHERE code = ?", code.Hex())
	var val []byte
	var version int
	err := row.Scan(&val, &version)
	if ErrIsNotExist(err) {
		return nil, -1, nil
	}

	if err != nil {
		return nil, -1, err
	}

	asset, err := dao.decodeAsset(val)
	if err != nil {
		return nil, - 1, err
	}

	return asset, version, nil
}

func (dao *AssetDao) insert(asset *types.Asset) (error) {
	val, err := rlp.EncodeToBytes(asset)
	if err != nil{
		return err
	}

	code := asset.AssetCode
	addr := asset.Issuer
	sql := "INSERT INTO t_asset(code, addr, attrs, utc_st, version)VALUES(?,?,?,?,?)"
	result, err := dao.engine.Exec(sql, code.Hex(), addr.Hex(), val, time.Now().UnixNano() / 1000000, 1)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if effected != 1{
		log.Errorf("insert asset.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	}else{
		return nil
	}
}

func (dao *AssetDao) update(asset *types.Asset, version int) (error) {
	val, err := rlp.EncodeToBytes(asset)
	if err != nil{
		return err
	}

	code := asset.AssetCode
	addr := asset.Issuer
	sql := "UPDATE t_asset SET attrs = ?, version = version + 1 WHERE code = ? AND addr = ? AND version = ?"
	result, err := dao.engine.Exec(sql,  val, code.Hex(), addr.Hex(), version)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if effected != 1{
		log.Errorf("update asset.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	}else{
		return nil
	}
}
