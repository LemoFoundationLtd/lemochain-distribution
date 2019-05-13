package database

import (
	"database/sql"
)

func ErrIsNotExist(err error) bool {
	if err == sql.ErrNoRows {
		return true
	} else {
		return false
	}
}

type DBEngine interface {
	GetDB() *sql.DB
}

type MySqlDB struct {
	engine *sql.DB
	driver string
	dns    string
}

func NewMySqlDB(driver string, dns string) *MySqlDB {
	db, err := Open(driver, dns)
	if err != nil {
		panic("open mysql err: " + err.Error())
	}

	return &MySqlDB{
		engine: db,
		driver: driver,
		dns:    dns,
	}
}

func (db *MySqlDB) GetDB() *sql.DB {
	return db.engine
}

func (db *MySqlDB) Clear() error {
	_, err := db.engine.Exec("DELETE FROM t_kv")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_tx")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_context")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_asset")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_equity")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_meta_data")
	if err != nil {
		return err
	}

	_, err = db.engine.Exec("DELETE FROM t_candidates")
	if err != nil {
		return err
	}
	return nil
}

func (db *MySqlDB) Close() {
	db.engine.Close()
}
