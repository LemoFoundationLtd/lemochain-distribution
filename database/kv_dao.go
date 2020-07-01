package database

import (
	"database/sql"
	"encoding/binary"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
)

var (
	hashPrefix = []byte("B")
	hashSuffix = []byte("b")

	heightPrefix = []byte("H")
	heightSuffix = []byte("h") // // headerPrefix + height (uint64 big endian) + heightSuffix -> hash

	accountPrefix = []byte("A")
	accountSuffix = []byte("a")

	storagePrefix = []byte("S")
	storageSuffix = []byte("s")

	assetCodePrefix = []byte("C")
	assetCodeSuffix = []byte("c")

	assetIdPrefix = []byte("I")
	assetIdSuffix = []byte("i")

	lastScanPosPrefix = []byte("P")
	lastScanPosSuffix = []byte("p")

	currentBlockKey = []byte("LEMO-CURRENT-BLOCK")
)

func encodeNumber(height uint32) []byte {
	enc := make([]byte, 4)
	binary.BigEndian.PutUint32(enc, height)
	return enc
}

func GetCanonicalKey(height uint32) []byte {
	newKey := make([]byte, 0, 6)
	newKey = append(append(append(newKey, heightPrefix...), encodeNumber(height)...), heightSuffix...)
	return newKey
}

func GetBlockHashKey(hash common.Hash) []byte {
	newKey := make([]byte, 0, 34)
	newKey = append(append(append(newKey, hashPrefix...), hash.Bytes()...), hashSuffix...)
	return newKey
}

func GetAddressKey(addr common.Address) []byte {
	newKey := make([]byte, 0, 22)
	newKey = append(append(append(newKey, accountPrefix...), addr.Bytes()...), accountSuffix...)
	return newKey
}

func GetStorageKey(hash common.Hash) []byte {
	newKey := make([]byte, 0, 34)
	newKey = append(append(append(newKey, storagePrefix...), hash.Bytes()...), storageSuffix...)
	return newKey
}

/**
 * （1） hash => block
 * （2） hash => tx
 * （3） address => account
 */
type KvDao struct {
	engine *sql.DB
}

func NewKvDao(db DBEngine) *KvDao {
	return &KvDao{engine: db.GetDB()}
}

func (dao *KvDao) Get(key []byte) ([]byte, error) {
	if len(key) <= 0 {
		log.Errorf("get k/v. key is nil.")
		return nil, ErrArgInvalid
	}

	row := dao.engine.QueryRow("SELECT lm_val FROM t_kv WHERE lm_key = ?", common.ToHex(key))
	var val []byte
	err := row.Scan(&val)
	if ErrIsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	} else {
		return val, nil
	}
}

func (dao *KvDao) Set(key []byte, val []byte) error {
	if len(key) <= 0 {
		log.Errorf("set k/v. key is nil.")
		return ErrArgInvalid
	}

	result, err := dao.engine.Exec("REPLACE INTO t_kv(lm_key, lm_val) VALUES (?,?)", common.ToHex(key), val)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected <= 0 {
		log.Errorf("set k/v affected = 0")
		return ErrUnKnown
	}

	return nil
}
