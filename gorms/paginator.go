package gorms

import (
	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
)

type Pagination struct {
	Page int `json:"page" query:"page"`
	Size int `json:"size" query:"size"`
}

func (s *Pagination) Scope(db *gorm.DB) *gorm.DB {
	if s.Page < 1 {
		s.Page = 1
	}
	if s.Size < 1 && s.Size > 100 {
		s.Size = 20
	}
	offset := (s.Page - 1) * s.Size
	return db.Offset(offset).Limit(s.Size)
}

type PageResult[T any] struct {
	Total   int64 `json:"total"`
	Page    int64 `json:"page"`
	Content []T   `json:"content"`
}

// Page paging the data
func Page[T any](tx *gorm.DB, pagination Pagination) (*gorm.DB, PageResult[T]) {
	content := []T{}
	var count int64
	tx = tx.
		Count(&count).
		Scopes(pagination.Scope).
		Find(&content)
	page := PageResult[T]{
		Total:   count,
		Page:    maths.CeilDivide(count, int64(pagination.Size)),
		Content: content,
	}
	return tx, page
}

// PageConv paging & convert data
func PageConv[T, E any](tx *gorm.DB, pagination Pagination, conv func(v T) E) (PageResult[E], error) {
	var total int64
	var content []T
	tx = tx.
		Count(&total).
		Scopes(pagination.Scope).
		Find(&content)
	converts := []E{}
	for _, i := range content {
		converts = append(converts, conv(i))
	}
	page := PageResult[E]{
		Total:   total,
		Page:    maths.CeilDivide(total, int64(pagination.Size)),
		Content: converts,
	}
	return page, tx.Error
}
