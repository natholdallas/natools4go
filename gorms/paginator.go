package gorms

import (
	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
)

type Pagination struct {
	Page int `json:"page" query:"page"`
	Size int `json:"size" query:"size"`
}

type PageResult[T any] struct {
	Total   int64 `json:"total"`
	Page    int64 `json:"page"`
	Content []T   `json:"content"`
}

// Page paging the data
func Page[T any](tx *gorm.DB, s Pagination) (*gorm.DB, PageResult[T]) {
	content := []T{}
	var total int64
	tx = tx.
		Count(&total).
		Scopes(PaginateScope(s.Page, s.Size)).
		Find(&content)
	page := PageResult[T]{
		Total:   total,
		Page:    maths.CeilDivide(total, int64(s.Size)),
		Content: content,
	}
	return tx, page
}

// PageConv paging & convert data
func PageConv[T, E any](tx *gorm.DB, s Pagination, conv func(v T) E) (*gorm.DB, PageResult[E]) {
	content := []T{}
	var total int64
	tx = tx.
		Count(&total).
		Scopes(PaginateScope(s.Page, s.Size)).
		Find(&content)
	converts := []E{}
	for _, i := range content {
		converts = append(converts, conv(i))
	}
	page := PageResult[E]{
		Total:   total,
		Page:    maths.CeilDivide(total, int64(s.Size)),
		Content: converts,
	}
	return tx, page
}

func PaginateScope(page, size int) GormScope {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if size < 1 && size > 100 {
			size = 20
		}
		offset := (page - 1) * size
		return db.Offset(offset).Limit(size)
	}
}
