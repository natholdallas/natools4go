// Package orms provides advanced utilities for GORM, including generic models,
// automated pagination, dynamic sorting, and a fluent query builder.
package orms

import (
	"database/sql"
	"fmt"
	"slices"
	"time"

	"github.com/natholdallas/natools4go/slice"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// New returns a gorm.DB instance.
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

func MustNew(dialector gorm.Dialector, opts ...gorm.Option) *gorm.DB {
	tx, err := New(dialector, opts...)
	if err != nil {
		panic(err)
	}
	return tx
}

// DSN returns the dsn for the given driver name, dsn, and options.
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
		panic(fmt.Errorf("failed to generate DSN: %w", err))
	}
	return dsn
}

// Dialector returns a gorm.Dialector for the given driver name, dsn, and options.
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
		panic(fmt.Errorf("failed to open database: %w", err))
	}
	return d
}

// AutoMigrate creates tables for the given models.
func AutoMigrate(tx *gorm.DB, dst ...any) error {
	switch tx.Name() {
	case "mysql", "mariadb":
		tx = tx.Set("gorm:table_options", "COLLATE=utf8mb4_bin")
	case "sqlserver", "mssql":
		tx = tx.Set("gorm:table_options", "COLLATE=SQL_Latin1_General_CP1_CS_AS")
	}
	err := tx.AutoMigrate(dst...)
	return err
}

func MustAutoMigrate(tx *gorm.DB, dst ...any) {
	if err := AutoMigrate(tx, dst...); err != nil {
		panic(fmt.Errorf("failed to auto migrate: %w", err))
	}
}

// syncMode selects how SyncTables reconciles a table with its struct definition.
type syncMode uint8

const (
	modeSoft syncMode = iota
	modeHard
)

// SyncTablesSoft reconciles every registered table with its struct definition
// without dropping any data. Columns that exist in the database but not in the
// struct are kept, moved to the end of the table, and ordered by column name.
// Newly added struct columns are placed according to the declaration order, and
// shared columns are preserved as-is. Extra columns that must be moved are
// rebuilt with a best-effort type reconstructed from the database, so they may
// lose DEFAULT, COMMENT, or fractional-second precision.
func SyncTablesSoft(tx *gorm.DB, models ...any) error {
	return syncTables(tx, modeSoft, models...)
}

func MustSyncTablesSoft(tx *gorm.DB, models ...any) {
	if err := SyncTablesSoft(tx, models...); err != nil {
		panic(fmt.Errorf("failed to sync tables: %w", err))
	}
}

// SyncTablesHard reconciles every registered table with its struct definition.
// Data in kept columns is preserved; columns removed from the struct are
// dropped and the remaining columns are reordered to match the struct
// declaration order.
func SyncTablesHard(tx *gorm.DB, models ...any) error {
	return syncTables(tx, modeHard, models...)
}

func MustSyncTablesHard(tx *gorm.DB, models ...any) {
	if err := SyncTablesHard(tx, models...); err != nil {
		panic(fmt.Errorf("failed to sync tables: %w", err))
	}
}

// syncTables reconciles every registered table with its struct definition
// according to the given mode.
func syncTables(tx *gorm.DB, mode syncMode, models ...any) error {
	for _, model := range models {
		s, err := schemaOf(tx, model)
		if err != nil {
			return err
		}
		table := s.Table

		if !tx.Migrator().HasTable(model) {
			if autoMigrateErr := AutoMigrate(tx, model); autoMigrateErr != nil {
				return autoMigrateErr
			}
			continue
		}

		actual, err := columnPositions(tx, model)
		if err != nil {
			return err
		}
		desired := desiredColumns(s)

		// hard mode: drop columns that no longer exist in the struct (skip primary key)
		extras := []string{}
		if mode == modeHard {
			for name := range actual {
				if slices.Contains(desired, name) {
					continue
				}
				field := s.LookUpField(name)
				if field != nil && field.PrimaryKey {
					continue
				}
				fmt.Printf("sync %s: drop column %s\n", table, name)
				if err := tx.Migrator().DropColumn(model, name); err != nil {
					return err
				}
			}
		} else {
			// soft mode: keep database-only columns, sorted by name, at the end
			for name := range actual {
				if !slices.Contains(desired, name) {
					extras = append(extras, name)
				}
			}
			slices.Sort(extras)
		}

		// add columns that are missing from the database
		for _, name := range desired {
			if _, ok := actual[name]; ok {
				continue
			}
			fmt.Printf("sync %s: add column %s\n", table, name)
			if err := tx.Migrator().AddColumn(model, name); err != nil {
				return err
			}
		}

		// reorder columns to match the declaration order, keeping extras last
		return reorderColumns(tx, model, s, slices.Concat(desired, extras))
	}
	return nil
}

// ResetTables drops and recreates every registered table from its struct so the
// column order exactly matches the declaration order. All data is lost.
func ResetTables(tx *gorm.DB, models ...any) error {
	switch tx.Name() {
	case "mysql", "mariadb":
		tx.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer tx.Exec("SET FOREIGN_KEY_CHECKS=1")
	}
	for _, model := range models {
		s, err := schemaOf(tx, model)
		if err != nil {
			return err
		}
		if err := tx.Exec("DROP TABLE IF EXISTS ?", clause.Table{Name: s.Table}).Error; err != nil {
			return err
		}
		if err := AutoMigrate(tx, model); err != nil {
			return err
		}
		fmt.Printf("reset table %s\n", s.Table)
	}
	return nil
}

func MustResetTables(tx *gorm.DB, models ...any) {
	if err := ResetTables(tx, models...); err != nil {
		panic(fmt.Errorf("failed to reset tables: %w", err))
	}
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
