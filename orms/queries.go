package orms

import (
	"database/sql"

	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Query is a fluent, generic wrapper around [gorm.DB] for cleaner query construction.
type Query[T any] struct {
	tx *gorm.DB
}

// Q initializes a Query wrapper without setting a default model.
func Q[T any](tx *gorm.DB) *Query[T] {
	return &Query[T]{tx}
}

// QE initializes a Query wrapper and sets the generic type T as the GORM model.
func QE[T any](tx *gorm.DB) *Query[T] {
	return &Query[T]{tx: tx.Model(new(T))}
}

// QM initializes a Query wrapper where T is the result type and M is the database model.
func QM[T, M any](tx *gorm.DB) *Query[T] {
	return &Query[T]{tx: tx.Model(new(M))}
}

// QT initializes a Query wrapper targeting a specific table name.
func QT[T any](tx *gorm.DB, name string, args ...any) *Query[T] {
	return &Query[T]{tx: tx.Table(name, args...)}
}

// All subsequent methods are fluent proxies for standard GORM operations.

func (q *Query[T]) Model(value any) *Query[T] {
	q.tx = q.tx.Model(value)
	return q
}

func (q *Query[T]) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *Query[T] {
	q.tx = q.tx.Scopes(funcs...)
	return q
}

func (q *Query[T]) Preload(query string, args ...any) *Query[T] {
	q.tx = q.tx.Preload(query, args...)
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

func (q *Query[T]) Join(query string, args ...any) *Query[T] {
	q.tx = q.tx.Joins(query, args...)
	return q
}

func (q *Query[T]) Assign(attrs ...any) *Query[T] {
	q.tx = q.tx.Assign(attrs...)
	return q
}

func (q *Query[T]) Attrs(attrs ...any) *Query[T] {
	q.tx = q.tx.Attrs(attrs...)
	return q
}

func (q *Query[T]) Order(value any) *Query[T] {
	q.tx = q.tx.Order(value)
	return q
}

func (q *Query[T]) Unscoped() *Query[T] {
	q.tx = q.tx.Unscoped()
	return q
}

func (q *Query[T]) Clauses(conds ...clause.Expression) *Query[T] {
	q.tx = q.tx.Clauses(conds...)
	return q
}

// Transaction performs operations within a database transaction.
func (q *Query[T]) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return q.tx.Transaction(fc, opts...)
}

func (q *Query[T]) Begin(opts ...*sql.TxOptions) *Query[T] {
	q.tx = q.tx.Begin(opts...)
	return q
}

func (q *Query[T]) Commit() *Query[T] {
	q.tx = q.tx.Commit()
	return q
}

// Paginate executes the query with pagination and returns a Page[T].
// It automatically sets the model to T if not already defined.
func (q *Query[T]) Paginate(pagination Pagination) (Page[T], *gorm.DB) {
	if q.tx.Statement.Model == nil {
		model := new(T)
		q.tx = q.tx.Model(model)
	}
	total := int64(0)
	content := []T{}
	q.tx = q.tx.Count(&total).Scopes(pagination.Scope).Find(&content)
	page := maths.DivCeil(total, int64(pagination.Size))
	v := Page[T]{total, page, content}
	return v, q.tx
}
