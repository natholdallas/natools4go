package orms

import (
	"io"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/natholdallas/natools4go/maths"
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
	ID        T              `gorm:"column:id;primaryKey;comment:ID" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;comment:Created Time" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;comment:Updated Time" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;comment:Deleted Time" json:"deleted_at"`
} // @name SoftModel

// Model is a generic base model including ID and standard timestamps without Soft Delete.
type Model[T any] struct {
	ID        T         `gorm:"column:id;primaryKey;comment:ID" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;comment:Created Time" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;comment:Updated Time" json:"updated_at"`
} // @name Model

// IDModel is a simple generic model containing only a primary key.
type IDModel[T any] struct {
	ID T `gorm:"column:id;primaryKey;comment:ID" json:"id"`
} // @name IDModel

// Sorter represents a single column sorting configuration.
type Sorter struct {
	Column string `query:"column" json:"column"`
	Desc   bool   `query:"desc" json:"desc"`
} // @name Sorter

// Scope applies the sorting condition to the GORM transaction.
func (s *Sorter) Scope(tx *gorm.DB) *gorm.DB {
	if s.Valid() {
		return tx.Order(s.ToOrderBy())
	}
	return tx
}

// Valid reports whether the column is a safe bare identifier. This prevents
// user-supplied sort columns (bound via `query:"column"`) from injecting
// arbitrary SQL or referencing unknown columns.
func (s *Sorter) Valid() bool {
	return validColumn(s.Column)
}

func validColumn(col string) bool {
	if col == "" {
		return false
	}
	for i, r := range col {
		if r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (i > 0 && r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

// ToOrderBy converts the Sorter to a GORM OrderByColumn expression.
func (s *Sorter) ToOrderBy() clause.OrderByColumn {
	return clause.OrderByColumn{Column: clause.Column{Name: s.Column}, Desc: s.Desc}
}

// Sorters represents multiple column sorting configurations.
type Sorters struct {
	Columns []Sorter `query:"columns" json:"columns"`
} // @name Sorters

// Scope applies multiple sorting conditions to the GORM transaction.
func (s *Sorters) Scope(tx *gorm.DB) *gorm.DB {
	if cols := s.ToOrderBy(); len(cols) > 0 {
		return tx.Order(clause.OrderBy{Columns: cols})
	}
	return tx
}

// ToOrderBy converts Sorters into a slice of GORM OrderByColumn expressions.
// Columns that fail validation are skipped.
func (s *Sorters) ToOrderBy() []clause.OrderByColumn {
	v := make([]clause.OrderByColumn, 0, len(s.Columns))
	for i := range s.Columns {
		if s.Columns[i].Valid() {
			v = append(v, s.Columns[i].ToOrderBy())
		}
	}
	return v
}

// Pagination defines the parameters for database results paging.
type Pagination struct {
	Page int `json:"page" query:"page"`
	Size int `json:"size" query:"size"`
} // @name Pagination

// Scope applies Offset and Limit to the GORM transaction based on pagination settings.
// It does not modify the receiver; invalid values are normalized on a copy.
func (s *Pagination) Scope(db *gorm.DB) *gorm.DB {
	page, size := s.normalize()
	offset := (page - 1) * size
	return db.Offset(offset).Limit(size)
}

// maxPage bounds the requested page to keep (page-1)*size from overflowing int
// on 32-bit platforms with the maximum size (100).
const maxPage = 1_000_000

// normalize returns sanitized page/size values without mutating the receiver.
func (s *Pagination) normalize() (page, size int) {
	page = min(max(s.Page, 1), maxPage)
	size = s.Size
	if size < 1 {
		size = 20
	} else if size > 100 {
		size = 100
	}
	return page, size
}

// Page represents a paginated result container.
type Page[T any] struct {
	Total   int64 `json:"total"`
	Page    int64 `json:"page"`
	Content []T   `json:"content"`
}

// PaginateMapping transforms the content of a Page from type T to E using a conversion function.
func PaginateMapping[T, E any](page Page[T], conv func(v T) E) Page[E] {
	converts := make([]E, 0, len(page.Content))
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
	// Compute the page count from the normalized size (the same value the
	// scope applies) so Size=0 or oversized values yield a consistent page.
	_, size := pagination.normalize()
	page := maths.DivCeil(total, int64(size))
	v := Page[T]{total, page, content}
	return v, tx
}

// LogPreset returns a logger with the given output and log level.
func LogPreset(out io.Writer, level logger.LogLevel) logger.Interface {
	return logger.New(log.New(out, "[DB] ", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
}
