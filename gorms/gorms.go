// Package gorms is tiny packaging support gorm
package gorms

import (
	"database/sql"
	"log"
)

func CreateDrop(database string, driverName, dataSourceName string) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Panic("failed to connect database, so we gonna be create it.")
	} else {
		log.Println("strategy create-drop enabled, so we reset the database to default.")
		db.Exec("DROP DATABASE IF EXISTS " + database)
	}
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + database); err != nil {
		log.Panic("failed to create database")
	}
	db.Close()
}
