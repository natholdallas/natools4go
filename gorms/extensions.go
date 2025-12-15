package gorms

import (
	"time"

	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormScope = func(*gorm.DB) *gorm.DB

// Getter dto transform to database model
type Getter[T any] interface {
	Get() *T
}

// Setter use dto sets database model's value
type Setter[T any] interface {
	Set(t *T)
}

// Scoper is [gorm.DB] condition bridge
type Scoper interface {
	Scope(tx *gorm.DB) *gorm.DB
}

type SoftModel[T any] struct {
	ID        T              `gorm:"column:id;primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

type Model[T any] struct {
	ID        T         `gorm:"column:id;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type IDModel[T any] struct {
	ID T `gorm:"column:id;primaryKey" json:"id"`
}

type Sorter struct {
	Column string `query:"column" json:"column"`
	Desc   bool   `query:"desc" json:"desc"`
}

func (s *Sorter) Scope(tx *gorm.DB) *gorm.DB {
	if s.Column != "" {
		return tx.Order(s.Conv())
	}
	return tx
}

func (s *Sorter) Conv() clause.OrderByColumn {
	return clause.OrderByColumn{Column: clause.Column{Name: s.Column}, Desc: s.Desc}
}

type Sorters struct {
	Columns []Sorter `query:"columns" json:"columns"`
}

func (s *Sorters) Scope(tx *gorm.DB) *gorm.DB {
	if len(s.Columns) > 0 {
		return tx.Order(clause.OrderBy{Columns: s.Conv()})
	}
	return tx
}

func (s *Sorters) Conv() []clause.OrderByColumn {
	res := []clause.OrderByColumn{}
	for i := range s.Columns {
		res = append(res, s.Columns[i].Conv())
	}
	return res
}

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

func Paginate[T any](tx *gorm.DB, pagination Pagination) (Page[T], error) {
	total := int64(0)
	content := []T{}
	err := tx.
		Count(&total).
		Scopes(pagination.Scope).
		Find(&content).Error
	page := Page[T]{
		Total:   total,
		Page:    maths.CeilDivide(total, int64(pagination.Size)),
		Content: content,
	}
	return page, err
}

// pagination & sort v2 design

type Query[T any] struct {
	tx *gorm.DB
}

func Q[T any](tx *gorm.DB) *Query[T] {
	return &Query[T]{tx}
}

func (q *Query[T]) Model(value any) *Query[T] {
	q.tx = q.tx.Model(value)
	return q
}

func (q *Query[T]) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *Query[T] {
	q.tx = q.tx.Scopes(funcs...)
	return q
}

func (q *Query[T]) Find(dest any, conds ...any) *Query[T] {
	q.tx = q.tx.Find(dest, conds...)
	return q
}

func (q *Query[T]) Select(query any, args ...any) *Query[T] {
	q.tx = q.tx.Select(query, args...)
	return q
}

func (q *Query[T]) First(dest any, conds ...any) *Query[T] {
	q.tx = q.tx.First(dest, conds...)
	return q
}

func (q *Query[T]) Where(query any, args ...any) *Query[T] {
	q.tx = q.tx.Where(query, args...)
	return q
}

func (q *Query[T]) Count(count *int64) *Query[T] {
	q.tx = q.tx.Count(count)
	return q
}

func (q *Query[T]) Scan(dest any) *Query[T] {
	q.tx = q.tx.Scan(dest)
	return q
}

func (q *Query[T]) Sort(sorter Sorter) *Query[T] {
	q.tx = q.tx.Scopes(sorter.Scope)
	return q
}

func (q *Query[T]) Sorts(sorter Sorters) *Query[T] {
	q.tx = q.tx.Scopes(sorter.Scope)
	return q
}

func (q *Query[T]) Paginate(v *Page[T], pagination Pagination) *Query[T] {
	total := int64(0)
	content := []T{}
	q.tx = q.tx.Count(&total).Scopes(pagination.Scope).Find(&content)
	v.Total = total
	v.Content = content
	v.Page = maths.CeilDivide(total, int64(pagination.Size))
	return q
}
