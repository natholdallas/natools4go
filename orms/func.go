package orms

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Count returns the total number of records for the specified model type T.
func Count[T any](tx *gorm.DB) int64 {
	var count int64
	// Using a pointer to T ensures GORM can infer the table name correctly.
	tx.Model(new(T)).Count(&count)
	return count
}

// Exists efficiently checks if at least one record exists for model type T.
// Performance Note: It uses Select("1") and Limit(1) to avoid full table counts.
func Exists[T any](tx *gorm.DB) bool {
	var n int
	// We use RowsAffected which is faster than Count for existence checks.
	return tx.Model(new(T)).Select("1").Limit(1).Scan(&n).RowsAffected > 0
}

// PluckStrings extracts a single string column into a slice.
// Useful for getting lists of IDs or Names quickly.
func PluckStrings[T any](tx *gorm.DB, column string) ([]string, error) {
	var v []string
	err := tx.Model(new(T)).Pluck(column, &v).Error
	return v, err
}

// First retrieves a single record by its primary key.
func First[T any](tx *gorm.DB, conds ...any) (T, error) {
	var v T
	err := tx.First(&v, conds...).Error
	return v, err
}

// IFirst retrieves a single record by its primary key ignoring the error.
func IFirst[T any](tx *gorm.DB, conds ...any) T {
	v, _ := First[T](tx, conds...)
	return v
}

// Take retrieves a single record by its primary key, returning an error if not found.
func Take[T any](tx *gorm.DB, conds ...any) (T, error) {
	var v T
	err := tx.Take(&v, conds...).Error
	return v, err
}

// ITake retrieves a single record by its primary key, ignoring the error if not found.
func ITake[T any](tx *gorm.DB, conds ...any) T {
	v, _ := Take[T](tx, conds...)
	return v
}

// Last retrieves the last record of the model type T.
func Last[T any](tx *gorm.DB, conds ...any) (T, error) {
	var v T
	err := tx.Last(&v, conds...).Error
	return v, err
}

// ILast retrieves the last record of the model type T ignoring the error.
func ILast[T any](tx *gorm.DB, conds ...any) T {
	v, _ := Last[T](tx, conds...)
	return v
}

// Find list all record
func Find[T any](tx *gorm.DB, conds ...any) ([]T, error) {
	v := []T{}
	err := tx.Find(&v, conds...).Error
	return v, err
}

// IFind list all record, if not found it will return a zero value
func IFind[T any](tx *gorm.DB, conds ...any) []T {
	v, _ := Find[T](tx, conds...)
	return v
}

// Save performs an Upsert (Update or Insert) based on the primary key's presence.
func Save[T any](tx *gorm.DB, v *T) error {
	return tx.Save(v).Error
}

// Create inserts a new record into the database.
func Create[T any](tx *gorm.DB, v *T) error {
	return tx.Create(v).Error
}

// primaryKey resolves the primary key column name for model type T,
// defaulting to "id" when no primary key field is declared.
func primaryKey[T any](tx *gorm.DB) (string, error) {
	stmt := tx.Model(new(T))
	if stmt.Statement.Schema == nil {
		if err := stmt.Statement.Parse(new(T)); err != nil {
			return "", err
		}
	}
	pk := "id"
	if f := stmt.Statement.Schema.PrioritizedPrimaryField; f != nil {
		pk = f.DBName
	}
	return pk, nil
}

// UpdateByID update a specific record by its primary key using column and value
func UpdateByID[T any](tx *gorm.DB, id any, column string, value any) error {
	pk, err := primaryKey[T](tx)
	if err != nil {
		return err
	}
	return tx.Model(new(T)).Where(pk+" = ?", id).Update(column, value).Error
}

// UpdatesByID updates a specific record by its primary key using a map or struct.
// It is recommended to use a map for updates to include zero-value fields (like 0, false, "").
func UpdatesByID[T any](tx *gorm.DB, id, values any) error {
	pk, err := primaryKey[T](tx)
	if err != nil {
		return err
	}
	return tx.Model(new(T)).Where(pk+" = ?", id).Updates(values).Error
}

// Delete performs a batch delete for the given conditions.
//
// A bare string condition is passed straight to GORM, which treats a string
// containing a space or "?" as a raw WHERE clause. That turns an unvalidated
// route parameter into SQL injection:
//
//	orms.Delete[User](tx, c.Params("id"))   // "1 or 1=1" deletes every row
//
// To keep that footgun out of calling code, Delete rejects such conditions and
// points at the typed helpers. Use DeleteByID for primary-key deletes and
// DeleteWhere for anything else.
func Delete[T any](tx *gorm.DB, conds ...any) error {
	if err := checkDeleteConds(conds); err != nil {
		return err
	}
	return tx.Delete(new(T), conds...).Error
}

// checkDeleteConds rejects raw SQL-looking strings so an unvalidated request
// parameter cannot widen a delete into a table-wide one.
func checkDeleteConds(conds []any) error {
	for _, cond := range conds {
		s, ok := cond.(string)
		if !ok {
			continue
		}
		if s == "" {
			return fmt.Errorf("orms: refusing empty string condition (use DeleteByID or DeleteWhere)")
		}
		if strings.ContainsAny(s, " ?=<>'\"()") {
			return fmt.Errorf("orms: refusing raw string condition %q (use DeleteWhere for explicit WHERE clauses)", s)
		}
	}
	return nil
}

// DeleteByID deletes the single record whose primary key equals id and reports
// how many rows were affected. It is the safe replacement for passing a route
// parameter to Delete.
func DeleteByID[T any](tx *gorm.DB, id any) (int64, error) {
	pk, err := primaryKey[T](tx)
	if err != nil {
		return 0, err
	}
	res := tx.Where(pk+" = ?", id).Delete(new(T))
	return res.RowsAffected, res.Error
}

// DeleteWhere deletes every record matching an explicit SQL condition.
// The condition is a developer-authored clause, never user input: keep values in
// the args so they stay parameterised.
func DeleteWhere[T any](tx *gorm.DB, query any, args ...any) error {
	return tx.Where(query, args...).Delete(new(T)).Error
}

// Exec executes raw sql
func Exec(tx *gorm.DB, sql string, values ...any) *gorm.DB {
	return tx.Exec(sql, values...)
}

// Begin starts a transaction.
func Transaction(tx *gorm.DB, fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return tx.Transaction(fc, opts...)
}
