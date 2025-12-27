package orms

import (
	"time"

	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormScope defines a function type compatible with gorm.Scopes.
type GormScope = func(*gorm.DB) *gorm.DB

// Getter defines a DTO that can transform itself into a database model.
type Getter[T any] interface {
	Get() *T
}

// Setter defines an interface to update a database model's values using a DTO.
type Setter[T any] interface {
	Set(t *T)
}

// Scoper defines an interface for components that can apply conditions to a gorm.DB instance.
type Scoper interface {
	Scope(tx *gorm.DB) *gorm.DB
}

// SoftModel is a generic base model including ID and standard timestamps with Soft Delete support.
type SoftModel[T any] struct {
	ID        T              `gorm:"column:id;primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// Model is a generic base model including ID and standard timestamps without Soft Delete.
type Model[T any] struct {
	ID        T         `gorm:"column:id;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// IDModel is a simple generic model containing only a primary key.
type IDModel[T any] struct {
	ID T `gorm:"column:id;primaryKey" json:"id"`
}

// Sorter represents a single column sorting configuration.
type Sorter struct {
	Column string `query:"column" json:"column"`
	Desc   bool   `query:"desc" json:"desc"`
}

// Scope applies the sorting condition to the GORM transaction.
func (s *Sorter) Scope(tx *gorm.DB) *gorm.DB {
	if s.Column != "" {
		return tx.Order(s.Conv())
	}
	return tx
}

// Conv converts the Sorter to a GORM OrderByColumn expression.
func (s *Sorter) Conv() clause.OrderByColumn {
	return clause.OrderByColumn{Column: clause.Column{Name: s.Column}, Desc: s.Desc}
}

// Sorters represents multiple column sorting configurations.
type Sorters struct {
	Columns []Sorter `query:"columns" json:"columns"`
}

// Scope applies multiple sorting conditions to the GORM transaction.
func (s *Sorters) Scope(tx *gorm.DB) *gorm.DB {
	if len(s.Columns) > 0 {
		return tx.Order(clause.OrderBy{Columns: s.Conv()})
	}
	return tx
}

// Conv converts Sorters into a slice of GORM OrderByColumn expressions.
func (s *Sorters) Conv() []clause.OrderByColumn {
	v := make([]clause.OrderByColumn, 0, len(s.Columns))
	for i := range s.Columns {
		v = append(v, s.Columns[i].Conv())
	}
	return v
}

// Pagination defines the parameters for database results paging.
type Pagination struct {
	Page int `json:"page" query:"page"`
	Size int `json:"size" query:"size"`
}

// Scope applies Offset and Limit to the GORM transaction based on pagination settings.
func (s *Pagination) Scope(db *gorm.DB) *gorm.DB {
	if s.Page < 1 {
		s.Page = 1
	}
	if s.Size < 1 || s.Size > 100 {
		s.Size = 20
	}
	offset := (s.Page - 1) * s.Size
	return db.Offset(offset).Limit(s.Size)
}

// Page represents a paginated result container.
type Page[T any] struct {
	Total   int64 `json:"total"`
	Page    int64 `json:"page"`
	Content []T   `json:"content"`
}

// PaginateMapping transforms the content of a Page from type T to E using a conversion function.
func PaginateMapping[T, E any](page Page[T], conv func(v T) E) Page[E] {
	converts := []E{}
	for i := range page.Content {
		converts = append(converts, conv(page.Content[i]))
	}
	return Page[E]{page.Total, page.Page, converts}
}

// Paginate executes a count query and a find query to return a populated Page container.
func Paginate[T any](tx *gorm.DB, pagination Pagination) (Page[T], *gorm.DB) {
	total := int64(0)
	content := []T{}
	tx = tx.Count(&total).Scopes(pagination.Scope).Find(&content)
	page := maths.DivCeil(total, int64(pagination.Size))
	v := Page[T]{total, page, content}
	return v, tx
}
