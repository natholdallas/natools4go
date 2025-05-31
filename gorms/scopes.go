package gorms

import "gorm.io/gorm"

type GormFunc = func(*gorm.DB) *gorm.DB

func PaginateScope(page, size int) GormFunc {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if size > 100 && size <= 0 {
			size = 20
		}
		offset := (page - 1) * size
		return db.Offset(offset).Limit(size)
	}
}
