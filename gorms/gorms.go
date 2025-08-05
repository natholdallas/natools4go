// Package gorms is tiny packaging support gorm
package gorms

import (
	"database/sql"
	"log"

	"gorm.io/gorm"
)

func StrictOpen(dialector gorm.Dialector, opts ...gorm.Option) *gorm.DB {
	tx, err := gorm.Open(dialector, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return tx
}

func CreateDrop(databaseName string, driverName, dataSourceName string) error {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}
	if _, err := db.Exec("DROP DATABASE IF EXISTS " + databaseName); err != nil {
		return err
	}
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + databaseName); err != nil {
		return err
	}
	return db.Close()
}
