package gorms

import (
	"gorm.io/gorm"
)

// Count to get an table data count
func Count[T any](tx *gorm.DB) int64 {
	model := new(T)
	count := int64(0)
	tx.Model(model).Count(&count)
	return count
}

// Exists search table has any data
func Exists[T any](tx *gorm.DB) bool {
	model := new(T)
	count := int64(0)
	tx.Model(model).Limit(1).Count(&count)
	return count > 0
}
