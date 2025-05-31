package gorms

import (
	"time"

	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
)

// high performance strategy: nano's level
// dto transform to database model
type Getter[T any] interface {
	Get() T
}

// high performance strategy: nano's level
// use dto sets database model's value
type Setter[T any] interface {
	Set(t *T)
}

type QueryAction interface {
	Condition(sql *gorm.DB)
}

// soft delete model design
type SoftModel struct {
	ID        uint           `gorm:"column:id" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(6)" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

// model design
type Model struct {
	ID        uint      `gorm:"column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// tiny model design
type TinyModel struct {
	ID        uint      `gorm:"column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

// micro model design
type MicroModel struct {
	ID uint `gorm:"column:id" json:"id"`
}

// uuid soft model design
type UUIDSoftModel struct {
	ID        string         `gorm:"column:id;type:uuid" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(6)" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

// uuid model design
type UUIDModel struct {
	ID        string    `gorm:"column:id;type:uuid" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// uuid tiny model design
type UUIDTinyModel struct {
	ID        string    `gorm:"column:id;type:uuid" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

// uuid micro model design
type UUIDMicroModel struct {
	ID string `gorm:"column:id;type:uuid" json:"id"`
}

// paginate configuration
type Pagination struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

// page result struct
type PageResult[T any] struct {
	Total   int64 `json:"total"`
	Page    int64 `json:"page"`
	Content []T   `json:"content"`
}

// paging
func Page[T any](tx *gorm.DB, s Pagination) (*gorm.DB, PageResult[T]) {
	content := []T{}
	var total int64
	sql := tx.
		Count(&total).
		Scopes(PaginateScope(s.Page, s.Size)).
		Find(&content)
	page := PageResult[T]{
		Total:   total,
		Page:    maths.CeilDivide(total, int64(s.Size)),
		Content: content,
	}
	return sql, page
}

// paging & convert
func PageConv[T, E any](tx *gorm.DB, s Pagination, convert func(v T) E) (*gorm.DB, PageResult[E]) {
	content := []T{}
	var total int64
	sql := tx.
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
	return sql, page
}

func Count(tx *gorm.DB, model any) int64 {
	var count int64
	tx.Model(model).Count(&count)
	return count
}

// TODO: use sql origin exists expression
func Exists(tx *gorm.DB, model any) bool {
	return Count(tx, model) > 0
}
