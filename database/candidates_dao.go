package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"math/big"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
)

type CandidateDao struct{
	engine *sql.DB
}

type CandidateItem struct{
	User     common.Address
	Votes    *big.Int
}

func NewCandidateDao(engine DBEngine) (*CandidateDao){
	return &CandidateDao{engine:engine.GetDB()}
}

func (dao *CandidateDao) Set(item *CandidateItem) (error) {
	if item == nil {
		log.Errorf("set candidate. item is nil.")
		return ErrArgInvalid
	}

	if item.User == (common.Address{}) {
		log.Errorf("set candidate. user is common.address{}")
	}

	var votes int64 = 0
	if item.Votes != nil{
		votes = item.Votes.Int64()
	}

	result, err := dao.engine.Exec("REPLACE INTO t_candidates(addr, votes) VALUES (?,?)", item.User.Hex(), votes)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if affected <= 0 {
		log.Errorf("set candidate affected = 0")
		return ErrUnKnown
	}else{
		return nil
	}
}

func (dao *CandidateDao) Del(user common.Address) (error) {
	if user == (common.Address{}) {
		log.Errorf("del candidate. user is common.address{}")
		return ErrArgInvalid
	}

	result, err := dao.engine.Exec("DELETE FROM t_candidates WHERE addr = ?", user.Hex())
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if affected <= 0 {
		log.Errorf("del candidate affected = 0")
		return ErrUnKnown
	}else{
		return nil
	}
}

func (dao *CandidateDao) buildCandidateBatch(rows *sql.Rows) ([]*CandidateItem, error) {
	result := make([]*CandidateItem, 0)
	for rows.Next() {
		var addr string
		var votes int64
		err := rows.Scan(&addr, &votes)
		if err != nil {
			return nil, err
		}

		result = append(result, &CandidateItem{
			User: common.HexToAddress(addr),
			Votes:big.NewInt(votes),
		})
	}
	return result, nil
}

func (dao *CandidateDao) GetTop(size int)([]*CandidateItem, error){
	if size <= 0 {
		log.Errorf("get top candidate. size <= 0")
		return nil, ErrArgInvalid
	}

	return dao.GetPage(0, size)
}

func (dao *CandidateDao) GetPage(start, limit int) ([]*CandidateItem, error) {
	if (start < 0) || (limit <= 0) {
		log.Errorf("get candidate by page.start < 0 or limit <= 0")
		return nil, ErrArgInvalid
	}

	sqlQuery := "SELECT addr, votes FROM t_candidates ORDER BY votes DESC LIMIT ?, ?"
	stmt, err := dao.engine.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(start, start + limit)
	if err != nil {
		return nil, err
	}

	return dao.buildCandidateBatch(rows)
}

func (dao *CandidateDao) GetPageWithTotal(start, limit int) ([]*CandidateItem, int, error) {
	if (start < 0) || (limit <= 0) {
		log.Errorf("get candidate by page with total.start < 0 or limit <= 0")
		return nil, -1, ErrArgInvalid
	}

	sqlTotal := "SELECT count(*) as cnt FROM t_candidates"
	row := dao.engine.QueryRow(sqlTotal)
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return nil, -1, err
	}

	candidates, err := dao.GetPage(start, limit)
	if err != nil{
		return nil, -1, err
	}else{
		return candidates, cnt, nil
	}
}

