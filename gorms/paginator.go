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

type Page[T any] struct {
	Total   int64 `json:"total"`
	Page    int64 `json:"page"`
	Content []T   `json:"content"`
}

func PageMap[T, E any](page Page[T], conv func(v T) E) Page[E] {
	converts := []E{}
	for i := range page.Content {
		converts = append(converts, conv(page.Content[i]))
	}
	return Page[E]{
		Total:   page.Total,
		Page:    page.Page,
		Content: converts,
	}
}

func Paginate[T any](db *gorm.DB, pagination Pagination) (Page[T], error) {
	var total int64
	var content []T
	err := db.Count(&total).Scopes(pagination.Scope).Find(&content).Error
	page := Page[T]{
		Total:   total,
		Page:    maths.CeilDivide(total, int64(pagination.Size)),
		Content: content,
	}
	return page, err
}
