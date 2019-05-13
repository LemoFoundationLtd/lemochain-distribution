package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"strconv"
	"time"
)

//go:generate gencodec -type MetaData -out gen_metadata_json.go
type MetaData struct {
	Id      common.Hash    `json:"id" gencodec:"required"`
	Code    common.Hash    `json:"code" gencodec:"required"`
	Owner   common.Address `json:"owner" gencodec:"required"`
	Profile string         `json:"profile" gencodec:"required"`
}

type MetaDataDao struct {
	engine *sql.DB
}

func NewMetaDataDao(db DBEngine) *MetaDataDao {
	return &MetaDataDao{engine: db.GetDB()}
}

func (dao *MetaDataDao) Set(metaData *MetaData) error {
	if metaData == nil {
		log.Errorf("set meta data.meta data is nil.")
		return ErrArgInvalid
	}

	result, version, err := dao.query(metaData.Id)
	if err != nil {
		return err
	}

	if result == nil {
		return dao.insert(metaData)
	} else {
		return dao.update(metaData, version)
	}
}

func (dao *MetaDataDao) Get(id common.Hash) (*MetaData, error) {
	if id == (common.Hash{}) {
		log.Errorf("get meta data.id is common.hash{}")
		return nil, ErrArgInvalid
	}

	data, _, err := dao.query(id)
	if err != nil {
		return nil, err
	}

	if data == nil {
		log.Errorf("get meta data.id is not exist.")
		return nil, ErrNotExist
	} else {
		return data, nil
	}
}

func (dao *MetaDataDao) decodeProfile(val []byte) (string, error) {
	var profile string
	err := rlp.DecodeBytes(val, &profile)
	if err != nil {
		return "", err
	} else {
		return profile, nil
	}
}

func (dao *MetaDataDao) buildMetaDataBatch(rows *sql.Rows) ([]*MetaData, error) {
	result := make([]*MetaData, 0)
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
			metaData := &MetaData{
				Code:    common.HexToHash(code),
				Id:      common.HexToHash(id),
				Owner:   common.HexToAddress(addr),
				Profile: profile,
			}
			result = append(result, metaData)
		}
	}
	return result, nil
}

func (dao *MetaDataDao) GetPage(addr common.Address, start, limit int) ([]*MetaData, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get meta by page.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sql := "SELECT id, code, addr, attrs, utc_st FROM t_meta_data WHERE addr = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(addr.Hex(), start, start+limit)
	if err != nil {
		return nil, err
	}

	return dao.buildMetaDataBatch(rows)
}

func (dao *MetaDataDao) GetPageWithTotal(addr common.Address, start, limit int) ([]*MetaData, int, error) {
	if addr == (common.Address{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get meta by page with total.addr is common.address{} or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sql := "SELECT count(*) as cnt FROM t_meta_data WHERE addr = ?"
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

func (dao *MetaDataDao) GetPageByCode(code common.Hash, start, limit int) ([]*MetaData, error) {
	if code == (common.Hash{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get meta by code.addr is common.address{} or start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sql := "SELECT  id, code, addr, attrs, utc_st FROM t_meta_data WHERE code = ? ORDER BY utc_st LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(code.Hex(), start, start+limit)
	if err != nil {
		return nil, err
	}

	return dao.buildMetaDataBatch(rows)
}

func (dao *MetaDataDao) GetPageByCodeWithTotal(code common.Hash, start, limit int) ([]*MetaData, int, error) {
	if code == (common.Hash{}) || (start < 0) || (limit <= 0) {
		log.Errorf("get meta by code with total.addr is common.address{} or start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sql := "SELECT count(*) as cnt FROM t_meta_data WHERE code = ?"
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

func (dao *MetaDataDao) query(id common.Hash) (*MetaData, int, error) {
	sql := "SELECT code, addr, attrs, version FROM t_meta_data WHERE id = ?"
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

	result := &MetaData{
		Id:    id,
		Code:  common.HexToHash(code),
		Owner: common.HexToAddress(addr),
	}

	var profile string
	err = rlp.DecodeBytes(val, &profile)
	if err != nil {
		return nil, -1, err
	} else {
		result.Profile = profile
		return result, version, nil
	}
}

func (dao *MetaDataDao) insert(metaData *MetaData) error {
	sql := "INSERT INTO t_meta_data(code, id, addr, attrs, version, utc_st)VALUES(?,?,?,?,?,?)"
	val, err := rlp.EncodeToBytes(metaData.Profile)
	if err != nil {
		return err
	}

	result, err := dao.engine.Exec(sql, metaData.Code.Hex(), metaData.Id.Hex(), metaData.Owner.Hex(), val, 1, time.Now().UnixNano()/1000000)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if effected != 1 {
		log.Errorf("update meta data.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	} else {
		return nil
	}
}

func (dao *MetaDataDao) update(metaData *MetaData, version int) error {
	val, err := rlp.EncodeToBytes(metaData.Profile)
	if err != nil {
		return err
	}

	sql := "UPDATE t_meta_data SET attrs = ?, version = version + 1 WHERE id = ? AND code = ? AND addr = ? AND version = ?"
	result, err := dao.engine.Exec(sql, val, metaData.Id.Hex(), metaData.Code.Hex(), metaData.Owner.Hex(), version)
	if err != nil {
		return err
	}

	effected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if effected != 1 {
		log.Errorf("update meta data.affected = " + strconv.Itoa(int(effected)))
		return ErrUnKnown
	} else {
		return nil
	}
}
