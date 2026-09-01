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
func New(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
	tx, err := gorm.Open(dialector, opts...)
	if err != nil {
		// Panic is preferred over log.Fatal in library code to allow for recovery
		// and to ensure stack traces are available.
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db, err := tx.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB instance: %w", err)
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return tx, nil
}

// MustNew returns the gorm.DB instance, panicking with context if an error occurs.
func MustNew(dialector gorm.Dialector, opts ...gorm.Option) *gorm.DB {
	tx, err := New(dialector, opts...)
	if err != nil {
		panic(err)
	}
	return tx
}

func DSN(driverName, name, username, password, host, port string) (string, error) {
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

func MustDSN(driverName, name, username, password, host, port string) string {
	dsn, err := DSN(driverName, name, username, password, host, port)
	if err != nil {
		panic(err)
	}
	return dsn
}

func Dialector(driverName string, dsn, name, query string, prepare ...bool) (gorm.Dialector, error) {
	needPrepare := slice.Defu(false, prepare)
	if driverName != "sqlite" && needPrepare {
		dn := driverName
		if dn == "mariadb" {
			dn = "mysql"
		}
		AutoCreate(name, dn, dsn)
	}
	switch driverName {
	case "mysql", "mariadb":
		d := dsn + name
		if query != "" {
			d += "?" + query
		}
		return mysql.Open(d), nil
	case "postgres", "postgresql":
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
		d := dsn + "?database=" + name
		if query != "" {
			d += "&" + query
		}
		return sqlserver.Open(d), nil
	case "clickhouse":
		d := dsn + "&database=" + name
		if query != "" {
			d += "&" + query
		}
		return clickhouse.Open(d), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s (supported: mysql, postgres, sqlite, sqlserver, clickhouse)", driverName)
	}
}

func MustDialector(driverName string, dsn, name, query string, prepare ...bool) gorm.Dialector {
	d, err := Dialector(driverName, dsn, name, query, prepare...)
	if err != nil {
		panic(err)
	}
	return d
}

func LogPreset(out io.Writer, level logger.LogLevel) logger.Interface {
	return logger.New(log.New(out, "[DB] ", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
}

// validateDBName ensures dbName is a safe, bare SQL identifier so it cannot
// be used for SQL injection in interpolated DDL.
func validateDBName(dbName string) error {
	for i, r := range dbName {
		valid := r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (i > 0 && r >= '0' && r <= '9')
		if !valid {
			return fmt.Errorf("invalid database name %q: must be a bare identifier (letters, digits, underscore)", dbName)
		}
	}
	if dbName == "" {
		return fmt.Errorf("invalid database name: must not be empty")
	}
	return nil
}

// Reset is a strategy to create database and drop the database, it will faster than turncate and
// most important it is affinity with dev mode while you are first design your database
func Reset(dbName, driverName, dataSourceName string) error {
	if err := validateDBName(dbName); err != nil {
		return err
	}
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to connect tt server: %w", err)
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
	if err := validateDBName(dbName); err != nil {
		panic(fmt.Errorf("failed to create database: %w", err))
	}
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
