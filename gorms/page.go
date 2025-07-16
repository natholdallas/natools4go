package gorms

import (
	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
)

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
func PageConv[T, E any](tx *gorm.DB, s Pagination, convert func(v T) E) (*gorm.DB, PageResult[E]) {
	content := []T{}
	var total int64
	tx = tx.
		Count(&total).
		Scopes(PaginateScope(s.Page, s.Size)).
		Find(&content)
	converts := []E{}
	for _, i := range content {
		converts = append(converts, convert(i))
	}
	page := PageResult[E]{
		Total:   total,
		Page:    maths.CeilDivide(total, int64(s.Size)),
		Content: converts,
	}
	return tx, page
}
