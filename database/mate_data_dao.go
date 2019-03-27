package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"time"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"strconv"
)

type MateData struct {
	Id   common.Hash
	Code common.Hash
	Addr common.Address
	Profile *types.Profile
}

type MateDataDao struct{
	engine *sql.DB
}

func NewMateDataDao(engine *sql.DB) (*MateDataDao){
	return &MateDataDao{engine:engine}
}

func (dao *MateDataDao) Set(mateData *MateData) (error) {
	if mateData == nil {
		log.Errorf("set mate data.mate data is nil.")
		return ErrArgInvalid
	}

	if mateData.Profile == nil {
		profile := make(types.Profile)
		mateData.Profile = &profile
	}

	result, version, err := dao.query(mateData.Id)
	if err != nil {
		return err
	}

	if result == nil{
		return dao.insert(mateData)
	}else{
		return dao.update(mateData, version)
	}
}

func (dao *MateDataDao) Get(id common.Hash) (*MateData, error) {
	if id == (common.Hash{}) {
		log.Errorf("get mate data.id is common.hash{}")
		return nil, ErrArgInvalid
	}

	data, _, err := dao.query(id)
	if err != nil{
		return nil, err
	}

	if data == nil{
		log.Errorf("get mate data.id is not exist.")
		return nil, ErrNotExist
	}else{
		return data, nil
	}
}

func (dao *MateDataDao) decodeProfile(val []byte)(*types.Profile, error){
	profile := make(types.Profile)
	err := rlp.DecodeBytes(val, &profile)
	if err != nil{
		return nil, err
	}else{
		return &profile, nil
	}
}

func (dao *MateDataDao) buildMateDataBatch(rows *sql.Rows) ([]*MateData, error) {
	result := make([]*MateData, 0)
	for rows.Next() {
		var id string
		var code string
		var addr string
		var val []byte
		var utcSt int64
		err := rows.Scan(&id, &code, &addr, &val, &utcSt)
		if err != nil {
			return nil, err
		}

		profile, err := dao.decodeProfile(val)
		if err != nil{
			return nil, err
		}else{
			mateData := &MateData{
				Code:common.HexToHash(code),
				Id:common.HexToHash(id),
				Addr:common.HexToAddress(addr),
				Profile:profile,
			}
			result = append(result, mateData)
		}
	}
	return result, nil
}

func (dao *MateDataDao) GetPage(addr common.Address, start, stop int) ([]*MateData, error ){
	if addr == (common.Address{}) || (start < 0) || (stop <= 0) {
		log.Errorf("get mate by page.addr is common.address{} or start < 0 or stop <= 0")
		return nil, ErrArgInvalid
	}

	sql := "SELECT id, code, addr, attrs, utc_st FROM t_mate_data WHERE addr = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildMateDataBatch(rows)
}

func (dao *MateDataDao) GetPageWithTotal(addr common.Address, start, stop int) ([]*MateData, int, error) {
	if addr == (common.Address{}) || (start < 0) || (stop <= 0) {
		log.Errorf("get mate by page with total.addr is common.address{} or start < 0 or stop <= 0")
		return nil, -1, ErrArgInvalid
	}

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
	if code == (common.Hash{}) || (start < 0) || (stop <= 0) {
		log.Errorf("get mate by code.addr is common.address{} or start < 0 or stop <= 0")
		return nil, ErrArgInvalid
	}

	sql := "SELECT  id, code, addr, attrs, utc_st FROM t_mate_data WHERE code = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(code.Hex(), start, stop)
	if err != nil {
		return nil, err
	}

	return dao.buildMateDataBatch(rows)
}

func (dao *MateDataDao) GetPageByCodeWithTotal(code common.Hash, start, stop int) ([]*MateData, int, error) {
	if code == (common.Hash{}) || (start < 0) || (stop <= 0) {
		log.Errorf("get mate by code with total.addr is common.address{} or start < 0 or stop <= 0")
		return nil, -1, ErrArgInvalid
	}

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
	sql := "SELECT code, addr, attrs, version FROM t_mate_data WHERE id = ?"
	row := dao.engine.QueryRow(sql, id.Hex())
	var code string
	var addr string
	var val []byte
	var version int
	err := row.Scan(&code, &addr, &val, &version)
	if ErrIsNotExist(err) {
		return nil, -1, nil
	}

	if err != nil {
		return nil, -1, err
	}

	result := &MateData{
		Id:id,
		Code:common.HexToHash(code),
		Addr:common.HexToAddress(addr),
	}

	profile := make(types.Profile)
	err = rlp.DecodeBytes(val, &profile)
	if err != nil {
		return nil, - 1, err
	}else{
		result.Profile = &profile
		return result, version, nil
	}
}

func (dao *MateDataDao) insert(mateData *MateData) (error) {
	sql := "INSERT INTO t_mate_data(code, id, addr, attrs, version, utc_st)VALUES(?,?,?,?,?,?)"
	val, err := rlp.EncodeToBytes(mateData.Profile)
	if err != nil{
		return err
	}

	result, err := dao.engine.Exec(sql, mateData.Code.Hex(), mateData.Id.Hex(), mateData.Addr.Hex(), val, 1, time.Now().UnixNano() / 1000000)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if effected != 1{
		log.Errorf("update mate data.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	}else{
		return nil
	}
}

func (dao *MateDataDao) update(mateData *MateData, version int) (error) {
	val, err := rlp.EncodeToBytes(mateData.Profile)
	if err != nil{
		return err
	}

	sql := "UPDATE t_mate_data SET attrs = ?, version = version + 1 WHERE id = ? AND code = ? AND addr = ? AND version = ?"
	result, err := dao.engine.Exec(sql,  val, mateData.Id.Hex(), mateData.Code.Hex(), mateData.Addr.Hex(), version)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if effected != 1{
		log.Errorf("update mate data.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	}else{
		return nil
	}
}


