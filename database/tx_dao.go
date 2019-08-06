package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"time"
)

type Tx struct {
	BHash       common.Hash
	Height      uint32
	PHash       common.Hash // 为箱子交易的时候的箱子的hash
	THash       common.Hash
	From        common.Address
	To          common.Address
	Tx          *types.Transaction
	Flag        int
	St          int64       // 交易保存进db的时间
	PackageTime uint32      // 打包交易的时间
	AssetCode   common.Hash // 如果是资产交易则对应资产code
	AssetId     common.Hash // 对应资产交易的资产id
}

type TxDao struct {
	engine *sql.DB
}

func NewTxDao(db DBEngine) *TxDao {
	return &TxDao{engine: db.GetDB()}
}

func (dao *TxDao) Set(tx *Tx) error {
	if tx == nil {
		log.Errorf("set tx.tx is nil.")
		return ErrArgInvalid
	}

	sql := "REPLACE INTO t_tx(thash, phash, bhash, height, faddr, taddr, tx, flag, utc_st, package_time,asset_code,asset_id)VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"

	val, err := rlp.EncodeToBytes(tx.Tx)
	if err != nil {
		return err
	}

	height := int64(tx.Height)
	_, err = dao.engine.Exec(sql, tx.THash.Hex(), tx.PHash.Hex(), tx.BHash.Hex(), height, tx.From.Hex(), tx.To.Hex(), val, tx.Tx.Type(), time.Now().UnixNano()/1000000, tx.PackageTime, tx.AssetCode.Hex(), tx.AssetId.Hex())
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

	sql := "SELECT thash, phash, bhash, height, faddr, taddr, tx, flag, utc_st, package_time,asset_code,asset_id FROM t_tx WHERE thash = ?"
	row := dao.engine.QueryRow(sql, hash.Hex())
	var thash string
	var phash string
	var bhash string
	var height int64
	var faddr string
	var taddr string
	var flag int
	var st int64
	var packageTime uint32
	var assetCode string
	var assetId string
	var val []byte
	err := row.Scan(&thash, &phash, &bhash, &height, &faddr, &taddr, &val, &flag, &st, &packageTime, &assetCode, &assetId)
	if ErrIsNotExist(err) {
		return nil, ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	tx, err := dao.encodeTx(val)
	if err != nil {
		return nil, err
	} else {
		return &Tx{
			BHash:       common.HexToHash(bhash),
			Height:      uint32(height),
			PHash:       common.HexToHash(phash),
			THash:       common.HexToHash(thash),
			From:        common.HexToAddress(faddr),
			To:          common.HexToAddress(taddr),
			Tx:          tx,
			Flag:        flag,
			St:          st,
			PackageTime: packageTime,
			AssetCode:   common.HexToHash(assetCode),
			AssetId:     common.HexToHash(assetId),
		}, nil
	}
}

func (dao *TxDao) encodeTx(val []byte) (*types.Transaction, error) {
	var tx types.Transaction
	err := rlp.DecodeBytes(val, &tx)
	if err != nil {
		return nil, err
	} else {
		return &tx, nil
	}
}

func (dao *TxDao) buildTxBatch(rows *sql.Rows) ([]*Tx, error) {
	result := make([]*Tx, 0)
	for rows.Next() {
		var thash string
		var phash string
		var bhash string
		var height int64
		var faddr string
		var taddr string
		var flag int
		var st int64
		var packageTime uint32
		var assetCode string
		var assetId string
		var val []byte
		err := rows.Scan(&thash, &phash, &bhash, &height, &faddr, &taddr, &val, &flag, &st, &packageTime, &assetCode, &assetId)
		if err != nil {
			return nil, err
		}

		tx, err := dao.encodeTx(val)
		if err != nil {
			return nil, err
		} else {
			result = append(result, &Tx{
				BHash:       common.HexToHash(bhash),
				Height:      uint32(height),
				PHash:       common.HexToHash(phash),
				THash:       common.HexToHash(thash),
				From:        common.HexToAddress(faddr),
				To:          common.HexToAddress(taddr),
				Tx:          tx,
				Flag:        flag,
				St:          st,
				PackageTime: packageTime,
				AssetCode:   common.HexToHash(assetCode),
				AssetId:     common.HexToHash(assetId),
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

	sqlQuery := "SELECT thash, phash, bhash, height, faddr, taddr, tx, flag, utc_st, package_time,asset_code,asset_id FROM t_tx WHERE faddr = ? or taddr = ? ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), addr.Hex(), start, start+limit)
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
	if err != nil {
		return nil, -1, err
	} else {
		return txes, cnt, nil
	}
}

func (dao *TxDao) GetByTime(addr common.Address, stStart, stStop int64, start, limit int) ([]*Tx, error) {
	if addr == (common.Address{}) || (stStart < 0) || (stStop < 0) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by time.addr is common.address{} or time stamp < 0 or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sqlQuery := "SELECT thash, phash, bhash, height, faddr, taddr, tx, flag, utc_st, package_time,asset_code,asset_id FROM t_tx WHERE (faddr = ? OR taddr = ?) AND utc_st > ? AND utc_st < ? ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), addr.Hex(), stStop, stStart, start, start+limit)
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
	if err != nil {
		return nil, -1, err
	} else {
		return txes, cnt, nil
	}
}

func (dao *TxDao) GetByFrom(addr common.Address, start, limit int) ([]*Tx, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by from.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sqlQuery := "SELECT thash, phash, bhash, height, faddr, taddr, tx, flag, utc_st ,package_time,asset_code,asset_id FROM t_tx WHERE faddr = ? ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), start, start+limit)
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
	if err != nil {
		return nil, -1, err
	} else {
		return txes, cnt, nil
	}
}

func (dao *TxDao) GetByTo(addr common.Address, start, limit int) ([]*Tx, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by to.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sqlQuery := "SELECT thash, phash, bhash, height, faddr, taddr, tx, flag, utc_st, package_time,asset_code,asset_id FROM t_tx WHERE taddr = ? ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), start, start+limit)
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
	if err != nil {
		return nil, -1, err
	} else {
		return txes, cnt, nil
	}
}

// 通过 address 和 assetCode或者assetId查询交易
func (dao *TxDao) GetByAddressAndAssetCodeOrAssetId(addr common.Address, assetCodeOrId common.Hash, start, limit int) ([]*Tx, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by addr.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}
	sqlQuery := "SELECT thash, phash, bhash, height, faddr, taddr, tx, flag, utc_st, package_time,asset_code,asset_id FROM t_tx WHERE (faddr = ? OR taddr = ?) AND (asset_code = ? OR asset_id = ?) ORDER BY utc_st DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(addr.Hex(), addr.Hex(), assetCodeOrId.Hex(), assetCodeOrId.Hex(), start, start+limit)
	if err != nil {
		return nil, err
	}
	return dao.buildTxBatch(rows)
}

func (dao *TxDao) GetByAddressAndAssetCodeOrAssetIdWithTotal(addr common.Address, assetCodeOrId common.Hash, start, limit int) ([]*Tx, int, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get tx by addr with total.addr is common.address{} or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}
	sqlTotal := "SELECT count(*) as cnt FROM t_tx WHERE (faddr = ? OR taddr = ?) AND (asset_code = ? OR asset_id = ?)"
	row := dao.engine.QueryRow(sqlTotal, addr.Hex(), addr.Hex(), assetCodeOrId.Hex(), assetCodeOrId.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	txes, err := dao.GetByAddressAndAssetCodeOrAssetId(addr, assetCodeOrId, start, limit)
	if err != nil {
		return nil, -1, err
	} else {
		return txes, cnt, nil
	}
}
