package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"time"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"strconv"
)

//go:generate gencodec -type MateData -out gen_matedata_json.go
type MateData struct {
	Id      common.Hash    `json:"id" gencodec:"required"`
	Code    common.Hash    `json:"code" gencodec:"required"`
	Owner   common.Address `json:"owner" gencodec:"required"`
	Profile string         `json:"profile" gencodec:"required"`
}

type MateDataDao struct {
	engine *sql.DB
}

func NewMateDataDao(db DBEngine) (*MateDataDao) {
	return &MateDataDao{engine: db.GetDB()}
}

func (dao *MateDataDao) Set(mateData *MateData) (error) {
	if mateData == nil {
		log.Errorf("set mate data.mate data is nil.")
		return ErrArgInvalid
	}

	result, version, err := dao.query(mateData.Id)
	if err != nil {
		return err
	}

	if result == nil {
		return dao.insert(mateData)
	} else {
		return dao.update(mateData, version)
	}
}

func (dao *MateDataDao) Get(id common.Hash) (*MateData, error) {
	if id == (common.Hash{}) {
		log.Errorf("get mate data.id is common.hash{}")
		return nil, ErrArgInvalid
	}

	data, _, err := dao.query(id)
	if err != nil {
		return nil, err
	}

	if data == nil {
		log.Errorf("get mate data.id is not exist.")
		return nil, ErrNotExist
	} else {
		return data, nil
	}
}

func (dao *MateDataDao) decodeProfile(val []byte) (string, error) {
	var profile string
	err := rlp.DecodeBytes(val, &profile)
	if err != nil {
		return "", err
	} else {
		return profile, nil
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
		if err != nil {
			return nil, err
		} else {
			mateData := &MateData{
				Code:    common.HexToHash(code),
				Id:      common.HexToHash(id),
				Owner:   common.HexToAddress(addr),
				Profile: profile,
			}
			result = append(result, mateData)
		}
	}
	return result, nil
}

func (dao *MateDataDao) GetPage(addr common.Address, start, limit int) ([]*MateData, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get mate by page.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sql := "SELECT id, code, addr, attrs, utc_st FROM t_mate_data WHERE addr = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), start, start+limit)
	if err != nil {
		return nil, err
	}

	return dao.buildMateDataBatch(rows)
}

func (dao *MateDataDao) GetPageWithTotal(addr common.Address, start, limit int) ([]*MateData, int, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get mate by page with total.addr is common.address{} or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sql := "SELECT count(*) as cnt FROM t_mate_data WHERE addr = ?"
	row := dao.engine.QueryRow(sql, addr.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	data, err := dao.GetPage(addr, start, limit)
	if err != nil {
		return nil, -1, err
	} else {
		return data, cnt, nil
	}
}

func (dao *MateDataDao) GetPageByCode(code common.Hash, start, limit int) ([]*MateData, error) {
	if code == (common.Hash{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get mate by code.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sql := "SELECT  id, code, addr, attrs, utc_st FROM t_mate_data WHERE code = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(code.Hex(), start, start+limit)
	if err != nil {
		return nil, err
	}

	return dao.buildMateDataBatch(rows)
}

func (dao *MateDataDao) GetPageByCodeWithTotal(code common.Hash, start, limit int) ([]*MateData, int, error) {
	if code == (common.Hash{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get mate by code with total.addr is common.address{} or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sql := "SELECT count(*) as cnt FROM t_mate_data WHERE code = ?"
	row := dao.engine.QueryRow(sql, code.Hex())
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	data, err := dao.GetPageByCode(code, start, limit)
	if err != nil {
		return nil, -1, err
	} else {
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
		Id:    id,
		Code:  common.HexToHash(code),
		Owner: common.HexToAddress(addr),
	}

	var profile string
	err = rlp.DecodeBytes(val, &profile)
	if err != nil {
		return nil, - 1, err
	} else {
		result.Profile = profile
		return result, version, nil
	}
}

func (dao *MateDataDao) insert(mateData *MateData) (error) {
	sql := "INSERT INTO t_mate_data(code, id, addr, attrs, version, utc_st)VALUES(?,?,?,?,?,?)"
	val, err := rlp.EncodeToBytes(mateData.Profile)
	if err != nil {
		return err
	}

	result, err := dao.engine.Exec(sql, mateData.Code.Hex(), mateData.Id.Hex(), mateData.Owner.Hex(), val, 1, time.Now().UnixNano()/1000000)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if effected != 1 {
		log.Errorf("update mate data.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	} else {
		return nil
	}
}

func (dao *MateDataDao) update(mateData *MateData, version int) (error) {
	val, err := rlp.EncodeToBytes(mateData.Profile)
	if err != nil {
		return err
	}

	sql := "UPDATE t_mate_data SET attrs = ?, version = version + 1 WHERE id = ? AND code = ? AND addr = ? AND version = ?"
	result, err := dao.engine.Exec(sql, val, mateData.Id.Hex(), mateData.Code.Hex(), mateData.Owner.Hex(), version)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if effected != 1 {
		log.Errorf("update mate data.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	} else {
		return nil
	}
}
