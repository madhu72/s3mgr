package main

import (
	"github.com/dgraph-io/badger/v4"
	"s3mgr/config"
)

func InitDB(cfg *config.Config) (*badger.DB, error) {
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = "s3mgr.db"
	}
	
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable badger logging
	
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	
	return db, nil
}
