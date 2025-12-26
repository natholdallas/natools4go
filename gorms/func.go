package gorms

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
	result := tx.Model(new(T)).Select("1").Limit(1).Scan(&n)
	return result.RowsAffected > 0
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
	var dest T
	err := tx.First(&dest, id).Error
	if err != nil {
		return dest, err
	}
	return dest, nil
}

// FindAll list all record
func FindAll[T any](tx *gorm.DB) ([]T, error) {
	dest := []T{}
	err := tx.Find(&dest).Error
	if err != nil {
		return dest, err
	}
	return dest, nil
}

// DeleteByIDs performs a batch delete for the given primary keys.
func DeleteByIDs[T any](tx *gorm.DB, ids []any) error {
	return tx.Delete(new(T), ids).Error
}

// Create inserts a new record into the database.
func Create[T any](tx *gorm.DB) (T, error) {
	v := new(T)
	err := tx.Create(v).Error
	return *v, err
}

// UpdateByID updates a specific record by its primary key using a map or struct.
// It is recommended to use a map for updates to include zero-value fields (like 0, false, "").
func UpdateByID[T any](tx *gorm.DB, id, values any) error {
	return tx.Model(new(T)).Where("id = ?", id).Updates(values).Error
}

// Save performs an Upsert (Update or Insert) based on the primary key's presence.
func Save[T any](tx *gorm.DB) (T, error) {
	v := new(T)
	err := tx.Save(v).Error
	return *v, err
}

// Updates applies batch updates to records matching the conditions in tx.
func Updates[T any](tx *gorm.DB, values any) error {
	return tx.Model(new(T)).Updates(values).Error
}
