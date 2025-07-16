package gorms

import "gorm.io/gorm"

// Count to get an table data count
func Count(tx *gorm.DB, model any) int64 {
	var count int64
	tx.Model(model).Count(&count)
	return count
}

// Exists search table has any data
func Exists(tx *gorm.DB, model any) bool {
	var count int64
	tx.Model(model).Limit(1).Count(&count)
	return count > 0
}
