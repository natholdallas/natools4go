package orms

import (
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

// FindByID retrieves a single record by its primary key.
func FindByID[T any](tx *gorm.DB, id any) (T, error) {
	var v T
	err := tx.First(&v, id).Error
	return v, err
}

// FindAll list all record
func FindAll[T any](tx *gorm.DB) ([]T, error) {
	v := []T{}
	err := tx.Find(&v).Error
	return v, err
}

// DeleteByID performs a batch delete for the given primary keys.
func DeleteByID[T any](tx *gorm.DB, ids ...any) error {
	return tx.Delete(new(T), ids).Error
}

// Create inserts a new record into the database.
func Create[T any](tx *gorm.DB, v *T) error {
	return tx.Create(v).Error
}

// UpdatesByID updates a specific record by its primary key using a map or struct.
// It is recommended to use a map for updates to include zero-value fields (like 0, false, "").
func UpdatesByID[T any](tx *gorm.DB, id, values any) error {
	return tx.Model(new(T)).Where("id = ?", id).Updates(values).Error
}

// Save performs an Upsert (Update or Insert) based on the primary key's presence.
func Save[T any](tx *gorm.DB, v *T) error {
	return tx.Save(v).Error
}
