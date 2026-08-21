// Package orms provides advanced utilities for GORM, including generic models,
// automated pagination, dynamic sorting, and a fluent query builder.
package orms

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/natholdallas/natools4go/slice"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db, err := tx.DB()
	if err != nil {
		panic(fmt.Errorf("failed to get underlying sql.DB instance: %w", err))
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return tx
}

func LogPreset(out io.Writer, level logger.LogLevel) logger.Interface {
	return logger.New(log.New(out, "[DB] ", log.Ldate|log.Ltime), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
}

// Reset is a strategy to create database and drop the database, it will faster than turncate and
// most important it is affinity with dev mode while you are first design your database
func Reset(dbName, driverName, dataSourceName string) error {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer db.Close()
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

// AutoCreate creates a database if it does not already exist using the provided driver and data source.
func AutoCreate(dbName, driverName, dataSourceName string) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	defer db.Close()
	createQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	if _, err := db.Exec(createQuery); err != nil {
		panic(fmt.Errorf("failed to create database %s: %w", dbName, err))
	}
}

func Dsn(driverName, name, username, password, host, port string) (string, error) {
	switch driverName {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port), nil
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s user=%s password=%s port=%s",
			host, username, password, port), nil
	case "sqlite", "sqlite3":
		return name, nil
	case "sqlserver", "mssql":
		return fmt.Sprintf("sqlserver://%s:%s@%s:%s", username, password, host, port), nil
	case "clickhouse":
		return fmt.Sprintf("tcp://%s:%s?username=%s&password=%s", host, port, username, password), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s (supported: mysql, postgres, sqlite, sqlserver, clickhouse)", driverName)
	}
}

func Dialector(driverName string, dsn, name, query string, prepare ...bool) (gorm.Dialector, error) {
	needPrepare := slice.Defu(false, prepare)
	switch driverName {
	case "mysql":
		if needPrepare {
			AutoCreate(name, "mysql", dsn)
		}
		d := dsn + name
		if query != "" {
			d += "?" + query
		}
		return mysql.Open(d), nil
	case "postgres", "postgresql":
		if needPrepare {
			AutoCreate(name, driverName, dsn)
		}
		d := dsn + " dbname=" + name
		if query != "" {
			d += " " + query
		}
		return postgres.Open(d), nil
	case "sqlite", "sqlite3":
		d := name
		if query != "" {
			d += "?" + query
		}
		return sqlite.Open(d), nil
	case "sqlserver", "mssql":
		if needPrepare {
			AutoCreate(name, driverName, dsn)
		}
		d := dsn + "?database=" + name
		if query != "" {
			d += "&" + query
		}
		return sqlserver.Open(d), nil
	case "clickhouse":
		if needPrepare {
			AutoCreate(name, driverName, dsn)
		}
		d := dsn + "&database=" + name
		if query != "" {
			d += "&" + query
		}
		return clickhouse.Open(d), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s (supported: mysql, postgres, sqlite, sqlserver, clickhouse)", driverName)
	}
}
