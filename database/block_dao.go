package database

import (
	"database/sql"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"strconv"
)

type BlockDao struct{
	engine *sql.DB
}

func NewBlockDao(engine *sql.DB) (*BlockDao){
	return &BlockDao{engine:engine}
}

func (dao *BlockDao) SetBlock(hash common.Hash, block *types.Block) (error) {
	if (hash == common.Hash{}) || (block == nil) {
		log.Errorf("set block.hash is common.hash{} or block is nil.")
		return ErrArgInvalid
	}

	val, err := rlp.EncodeToBytes(block)
	if err != nil{
		return err
	}else{
		kvDao := NewKvDao(dao.engine)
		err = kvDao.Set(GetCanonicalKey(block.Height()), hash.Bytes())
		if err != nil{
			return err
		}

		return kvDao.Set(GetBlockHashKey(hash), val)
	}
}

func (dao *BlockDao) GetBlock(hash common.Hash) (*types.Block, error) {
	if hash == (common.Hash{}) {
		log.Errorf("get block.hash is common.hash{}")
		return nil, ErrArgInvalid
	}

	kvDao := NewKvDao(dao.engine)
	val, err := kvDao.Get(GetBlockHashKey(hash))
	if err != nil {
		return nil, err
	}

	if val == nil{
		log.Errorf("get block by hash.is not exist.hash: " + hash.Hex())
		return nil, ErrNotExist
	}

	var block types.Block
	err = rlp.DecodeBytes(val, &block)
	if err != nil {
		return nil, err
	}else{
		return &block, nil
	}
}

func (dao *BlockDao) GetBlockByHeight(height uint32) (*types.Block, error) {
	kvDao := NewKvDao(dao.engine)
	val, err := kvDao.Get(GetCanonicalKey(height))
	if err != nil {
		return nil, err
	}

	if val == nil{
		log.Errorf("get block by height.is not exist.height: " + strconv.Itoa(int(height)))
		return nil, ErrNotExist
	}

	return dao.GetBlock(common.BytesToHash(val))
}
