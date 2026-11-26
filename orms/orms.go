// Package orms provides advanced utilities for GORM, including generic models,
// automated pagination, dynamic sorting, and a fluent query builder.
package orms

import (
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/natholdallas/natools4go/narder"
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
	// A file-backed SQLite database allows one writer at a time; a multi-
	// connection pool turns concurrent writes into SQLITE_BUSY errors. Pin the
	// pool to a single connection. Consequence: do not start a goroutine that
	// queries the same *gorm.DB while a transaction on it is open — it will wait
	// for the only connection.
	if dialector != nil && dialector.Name() == "sqlite" {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	}
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

// AutoMigrate creates tables for the given models.
//
// MySQL/MariaDB get the InnoDB engine plus an explicit charset/collation; GORM
// appends the option string directly after the closing parenthesis of the
// CREATE TABLE, so the leading space is intentional.
//
// SQL Server has no table-level COLLATE clause (case sensitivity is a
// database- or column-level setting), so no table options are set there — the
// previous value emitted syntactically invalid DDL.
func AutoMigrate(tx *gorm.DB, dst ...any) error {
	switch tx.Name() {
	case "mysql", "mariadb":
		tx = tx.Set("gorm:table_options", " ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin")
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

		actual, err := columnOrder(tx, model)
		if err != nil {
			return err
		}
		desired := desiredColumns(s)

		var extras []string
		if mode == modeHard {
			// MySQL/MariaDB can drop in place before the in-place reorder. Other
			// drivers skip DropColumn (GORM's SQLite implementation drops the
			// table's indexes with the column) and simply exclude the column
			// from the rebuilt target instead.
			if tx.Name() == "mysql" || tx.Name() == "mariadb" {
				for _, name := range actual {
					if slices.Contains(desired, name) {
						continue
					}
					narder.Infof("sync %s: drop column %s", table, name)
					if err := tx.Migrator().DropColumn(model, name); err != nil {
						return err
					}
				}
			}
		} else {
			// soft mode: keep database-only columns, sorted by name, at the end
			for _, name := range actual {
				if !slices.Contains(desired, name) {
					extras = append(extras, name)
				}
			}
			slices.Sort(extras)
		}

		// Add columns missing from the database. They must exist in the old
		// table before the rebuild copies the data across.
		have := make(map[string]bool, len(actual))
		for _, name := range actual {
			have[name] = true
		}
		for _, name := range desired {
			if have[name] {
				continue
			}
			narder.Infof("sync %s: add column %s", table, name)
			if err := tx.Migrator().AddColumn(model, name); err != nil {
				return err
			}
		}

		// Reorder to the declaration order, keeping soft-mode extras last.
		// reorderColumns is a no-op when the physical order already matches.
		if err := reorderColumns(tx, model, s, slices.Concat(desired, extras)); err != nil {
			return err
		}
	}
	return nil
}

// ResetTables drops and recreates every registered table from its struct so the
// column order exactly matches the declaration order. All data is lost.
//
// Tables are dropped with the migrator, never with CASCADE: a table that is
// still referenced by another table outside models makes PostgreSQL and SQL
// Server return an error instead of silently deleting dependent rows.
//
// Everything runs on a single dedicated connection: MySQL's
// SET FOREIGN_KEY_CHECKS is session-scoped, and issuing it on a different pooled
// connection than the DDL silently leaves the checks enabled.
func ResetTables(tx *gorm.DB, models ...any) error {
	return tx.Connection(func(db *gorm.DB) error {
		switch db.Name() {
		case "mysql", "mariadb":
			if err := db.Exec("SET FOREIGN_KEY_CHECKS=0").Error; err != nil {
				return err
			}
			defer db.Exec("SET FOREIGN_KEY_CHECKS=1")
		}
		for _, model := range models {
			s, err := schemaOf(db, model)
			if err != nil {
				return err
			}
			// The migrator quotes the table name per driver; a raw DROP TABLE
			// built from the schema name would break on reserved words.
			if err := db.Migrator().DropTable(model); err != nil {
				return err
			}
			if err := AutoMigrate(db, model); err != nil {
				return err
			}
			narder.Infof("reset table %s", s.Table)
		}
		return nil
	})
}

func MustResetTables(tx *gorm.DB, models ...any) {
	if err := ResetTables(tx, models...); err != nil {
		panic(fmt.Errorf("failed to reset tables: %w", err))
	}
}
