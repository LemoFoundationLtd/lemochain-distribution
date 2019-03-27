package database

import (
	"database/sql"
	"github.com/meitu/go-ethereum/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
)

type MateData struct {
	Code common.Hash
	Addr common.Address
	Profile types.Profile
}

type MateDataDao struct{
	engine *sql.DB
}

func (dao *MateDataDao) Set(id common.Hash, data *MateData) (error) {
	data, version, err := dao.query(id)
	if err == sql.ErrNoRows{
		return dao.insert(id, data)
	}else{
		return dao.update(id, data, version)
	}
}

func (dao *MateDataDao) Get(id common.Hash) (*MateData, error) {
	data, _, err := dao.query(id)
	if err != nil {
		return nil, err
	}else{
		return data, nil
	}
}

func (dao *MateDataDao) decodeMateData(val []byte)(*MateData, error){
	var mateData MateData
	err := rlp.DecodeBytes(val, &mateData)
	if err != nil{
		return nil, err
	}else{
		return &mateData, nil
	}
}

func (dao *MateDataDao) buildMateDataBatch(rows *sql.Rows) ([]*MateData, error) {
	result := make([]*MateData, 0)
	for rows.Next() {
		var val []byte
		var utcSt int64
		err := rows.Scan(&val, &utcSt)
		if err != nil {
			return nil, err
		}

		mateData, err := dao.decodeMateData(val)
		if err != nil{
			return nil, err
		}else{
			result = append(result, mateData)
		}
	}
	return result, nil
}

func (dao *MateDataDao) GetPage(addr common.Address, start, stop int) ([]*MateData, error ){
	sql := "SELECT attrs, utc_st FROM t_mate_data WHERE addr = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr, start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildMateDataBatch(rows)
}

func (dao *MateDataDao) GetPageWithTotal(addr common.Address, start, stop int) ([]*MateData, int, error) {
	sql := "SELECT count(*) as cnt FROM t_mate_data WHERE addr = ?"
	row := dao.engine.QueryRow(sql, addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	data, err := dao.GetPage(addr, start, stop)
	if err != nil{
		return nil, -1, err
	}else{
		return data, cnt, nil
	}
}

func (dao *MateDataDao) GetPageByCode(code common.Hash, start, stop int)([]*MateData, error) {
	sql := "SELECT attrs, utc_st FROM t_mate_data WHERE code = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(code, start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildMateDataBatch(rows)
}

func (dao *MateDataDao) GetPageByCodeWithTotal(code common.Hash, start, stop int) ([]*MateData, int, error) {
	sql := "SELECT count(*) as cnt FROM t_mate_data WHERE code = ?"
	row := dao.engine.QueryRow(sql, code.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	data, err := dao.GetPageByCode(code, start, stop)
	if err != nil{
		return nil, -1, err
	}else{
		return data, cnt, nil
	}
}

func (dao *MateDataDao) query(id common.Hash) (*MateData, int, error) {
	sql := "SELECT code, addr, attrs, version WHERE id = ?"
	row := dao.engine.QueryRow(sql, id)
	var data MateData
	var val []byte
	var version int
	err := row.Scan(&data.Code, &data.Addr, &val, &version)
	if err != nil {
		return nil, -1, err
	}

	err = rlp.DecodeBytes(val, &data.Profile)
	if err != nil {
		return nil, - 1, err
	}

	return &data, version, nil
}

func (dao *MateDataDao) insert(id common.Hash, data *MateData) (error) {
	sql := "INSERT INTO t_mate_data(code, id, addr, attrs, version)VALUES(?,?,?,?,1)"

	val, err := rlp.EncodeToBytes(data.Profile)
	if err != nil{
		return err
	}

	_, err = dao.engine.Exec(sql, id.Hex(), data.Code.Hex(), data.Addr.Hex(), val)
	if err != nil {
		return err
	}else{
		return nil
	}
}

func (dao *MateDataDao) update(id common.Hash, data *MateData, version int) (error) {
	if (id == (common.Hash{})) || (data == nil) {
		return nil
	}

	val, err := rlp.EncodeToBytes(data.Profile)
	if err != nil{
		return err
	}

	sql := "UPDATE t_mate_data SET attrs = ?, version = version + 1 WHERE id = ? AND code = ? AND addr = ? AND version = ?"
	_, err = dao.engine.Exec(sql,  val, id.Hex(), data.Code.Hex(), data.Addr.Hex(), version)
	if err != nil {
		return err
	}else{
		return nil
	}
}


