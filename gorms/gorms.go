// Package gorms is tiny packaging support gorm
package gorms

import (
	"database/sql"
)

func CreateDrop(databaseName string, driverName, dataSourceName string) error {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	} else {
		if _, err := db.Exec("DROP DATABASE IF EXISTS " + databaseName); err != nil {
			return err
		}
	}
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + databaseName); err != nil {
		return err
	}
	return db.Close()
}
