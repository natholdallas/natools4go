package gorms

import (
	"context"

	"gorm.io/gorm"
)

// Count to get an table data count
func Count(tx *gorm.DB, model any) (count int64) {
	tx.Model(model).Count(&count)
	return
}

// Emptied used to check table's data is empty
func Emptied(tx *gorm.DB, model any) bool {
	var count int64
	tx.Model(model).Count(&count)
	return count <= 0
}

// Exists search table has any data
func Exists(tx *gorm.DB, model any) bool {
	var count int64
	tx.Model(model).Limit(1).Count(&count)
	return count > 0
}

func SelectByID[T any](tx *gorm.DB, id any) (T, error) {
	return gorm.G[T](tx).Where("id", id).First(context.TODO())
}

func DeleteByID[T any](tx *gorm.DB, id any) (int, error) {
	return gorm.G[T](tx).Where("id", id).Delete(context.TODO())
}

func UpdateByID[T any](tx *gorm.DB, id any, name, value string) (int, error) {
	return gorm.G[T](tx).Where("id", id).Update(context.TODO(), name, value)
}

func UpdatesByID[T any](tx *gorm.DB, id any, t T) (int, error) {
	return gorm.G[T](tx).Where("id", id).Updates(context.TODO(), t)
}
