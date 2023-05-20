package db

import (
	tmdb "github.com/tendermint/tm-db"
)

type DB struct {
	home  string
	inner *tmdb.GoLevelDB
	cdc   Codec
}

func NewDB(home string) (*DB, error) {
	db := DB{
		home: home,
		cdc:  JsonCodec{},
	}

	inner, err := tmdb.NewGoLevelDB("data", home)
	if err != nil {
		return nil, err
	}
	db.inner = inner
	return &db, nil
}

func (db *DB) Get(key []byte, value interface{}) (bool, error) {
	bz, err := db.inner.Get(key)
	if err != nil {
		return false, err
	}
	if len(bz) == 0 {
		return false, nil
	}
	return true, db.cdc.Decode(bz, value)
}

func (db *DB) Set(key []byte, value interface{}) error {
	bz, err := db.cdc.Encode(value)
	if err != nil {
		return err
	}
	return db.inner.Set(key, bz)
}

func (db *DB) Delete(key []byte) error {
	return db.inner.Delete(key)
}

func (db *DB) Close() error {
	return db.inner.Close()
}

func (db *DB) Has(key []byte) (bool, error) {
	return db.inner.Has(key)
}