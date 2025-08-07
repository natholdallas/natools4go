// Package gorms is tiny packaging support gorm
// and in this file we provide some function, but honestly they can used in [sql.DB], so that requires design
package gorms

import (
	"database/sql"
	"log"
	"time"

	"gorm.io/gorm"
)

// StrictOpen is preset function to open gorm datasource, if err not nil it will be fatal
// but actually i want use panic to replace log.Fatal
// the template code shouldn't use such [log.Fatal] this function, it is business code
func StrictOpen(dialector gorm.Dialector, opts ...gorm.Option) *gorm.DB {
	tx, err := gorm.Open(dialector, opts...)
	if err != nil {
		log.Fatal(err)
		// TODO: replaced in next version
		// panic(err)
	}
	return tx
}

// CreateDropDB can create database and drop the database, it will faster than turncate and
// most important it is affinity with dev mode while you are first design your database
func CreateDropDB(dbName string, driverName, dataSourceName string) error {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}
	if _, err := db.Exec("DROP DATABASE IF EXISTS " + dbName); err != nil {
		return err
	}
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName); err != nil {
		return err
	}
	return db.Close()
}

// CreateDB can create database while you use sql
func CreateDB(dbName string, driverName, dataSourceName string) error {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName); err != nil {
		return err
	}
	return db.Close()
}

// GormPreset just to eliminate the reuse of code
func GormPreset(tx *gorm.DB) error {
	db, err := tx.DB()
	if err != nil {
		return err
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
	return nil
}
