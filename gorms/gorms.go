// Package gorms provides advanced utilities for GORM, including generic models,
// automated pagination, dynamic sorting, and a fluent query builder.
package gorms

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// New is preset function to open gorm datasource, if err not nil it will be fatal
// but actually i want use panic to replace log.Fatal
// the template code shouldn't use such [log.Fatal] this function, it is business code
func New(dialector gorm.Dialector, opts ...gorm.Option) *gorm.DB {
	tx, err := gorm.Open(dialector, opts...)
	if err != nil {
		// Panic is preferred over log.Fatal in library code to allow for recovery
		// and to ensure stack traces are available.
		panic(fmt.Errorf("failed to open database: %w", err))
	}
	return tx
}

// ResetDB is a strategy to create database and drop the database, it will faster than turncate and
// most important it is affinity with dev mode while you are first design your database
func ResetDB(dbName string, driverName, dataSourceName string) error {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer db.Close() // Ensure connection is closed even if Exec fails

	// Warning: Ensure dbName is sanitized to prevent SQL injection
	dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	if _, err := db.Exec(dropQuery); err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	createQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	if _, err := db.Exec(createQuery); err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	return nil
}

// EnsureDB creates a database if it does not already exist using the provided driver and data source.
func EnsureDB(dbName string, driverName, dataSourceName string) error {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	createQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	if _, err := db.Exec(createQuery); err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	return nil
}
