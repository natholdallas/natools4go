// Package orms provides advanced utilities for GORM, including generic models,
// automated pagination, dynamic sorting, and a fluent query builder.
package orms

import (
	"database/sql"
	"fmt"
	"slices"
	"time"

	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/natholdallas/natools4go/slice"
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
		if err := reorderColumns(tx, model, s, slices.Concat(desired, extras)); err != nil {
			return err
		}
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

// sqlDriverOf maps a GORM driver name to the database/sql driver that must be
// handed to sql.Open. GORM's dialectors blank-import their sql driver, so these
// are already registered by the time this package loads; the names simply do
// not match the GORM names ("sqlite" -> "sqlite3", "postgres" -> "pgx", ...).
//
// The second return value reports whether the driver is server-based, i.e.
// whether CREATE/DROP DATABASE is meaningful. File-backed sqlite databases
// create themselves on demand and have no "server" to connect to.
func sqlDriverOf(driverName string) (string, bool, error) {
	switch driverName {
	case "mysql", "mariadb":
		return "mysql", true, nil
	case "postgres", "postgresql":
		return "pgx", true, nil
	case "sqlserver", "mssql":
		return "sqlserver", true, nil
	case "clickhouse":
		return "clickhouse", true, nil
	case "sqlite", "sqlite3":
		return "sqlite3", false, nil
	default:
		return "", false, fmt.Errorf(
			"unsupported database driver: %s (supported: mysql, mariadb, postgres, sqlserver, clickhouse, sqlite)",
			driverName,
		)
	}
}

// createDatabase opens a connection to the server described by dataSourceName
// and creates dbName. It is shared by AutoCreate and Reset.
func createDatabase(dbName, driverName, dataSourceName string) error {
	driver, serverBased, err := sqlDriverOf(driverName)
	if err != nil {
		return err
	}
	if !serverBased {
		// A file-based database is created by the driver on first use; issuing
		// CREATE DATABASE against it is not supported and not needed.
		return nil
	}
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()
	createQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	if _, err := db.Exec(createQuery); err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}
	return nil
}

// Reset drops the configured database and recreates it empty. It is faster than
// truncating every table and is the intended workflow while designing a schema
// in development.
//
// Reset only supports server-based drivers (mysql, mariadb, postgres, sqlserver,
// clickhouse). For sqlite it is a no-op: delete the database file instead.
func Reset(dbName, driverName, dataSourceName string) error {
	if err := validateDBName(dbName); err != nil {
		return err
	}
	driver, serverBased, err := sqlDriverOf(driverName)
	if err != nil {
		return err
	}
	if !serverBased {
		return nil
	}
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer db.Close()
	dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	if _, err := db.Exec(dropQuery); err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}
	return createDatabase(dbName, driverName, dataSourceName)
}

// AutoCreate creates a database if it does not already exist using the provided
// driver and data source. It panics on failure; use AutoCreateE for error
// handling.
func AutoCreate(dbName, driverName, dataSourceName string) {
	if err := AutoCreateE(dbName, driverName, dataSourceName); err != nil {
		panic(fmt.Errorf("failed to create database: %w", err))
	}
}

// AutoCreateE is like AutoCreate but returns the error instead of panicking.
func AutoCreateE(dbName, driverName, dataSourceName string) error {
	if err := validateDBName(dbName); err != nil {
		return err
	}
	return createDatabase(dbName, driverName, dataSourceName)
}
