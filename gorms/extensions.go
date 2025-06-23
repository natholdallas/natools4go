package gorms

import (
	"time"

	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
)

// Getter dto transform to database model
type Getter[T any] interface {
	Get() T
}

// Setter use dto sets database model's value
type Setter[T any] interface {
	Set(t *T)
}

// QueryAction is [gorm.DB] bridge
type QueryAction interface {
	Condition(sql *gorm.DB)
}

type SoftModel struct {
	ID        uint           `gorm:"column:id" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(6)" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

type Model struct {
	ID        uint      `gorm:"column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

type TinyModel struct {
	ID        uint      `gorm:"column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

type MicroModel struct {
	ID uint `gorm:"column:id" json:"id"`
}

type UUIDSoftModel struct {
	ID        string         `gorm:"column:id;type:uuid" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(6)" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

type UUIDModel struct {
	ID        string    `gorm:"column:id;type:uuid" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

type UUIDTinyModel struct {
	ID        string    `gorm:"column:id;type:uuid" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

type UUIDMicroModel struct {
	ID string `gorm:"column:id;type:uuid" json:"id"`
}

type Pagination struct {
	Page int `json:"page"`
	Size int `json:"size"`
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

// PageConv paging & convert
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
