package orms

import (
	"database/sql"

	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
)

// Create inserts value, returning the inserted data's primary key in value's id
func (q *Query[T]) Create(value any) (tx *gorm.DB) {
	tx = q.tx.Create(value)
	return
}

// CreateInBatches inserts value in batches of batchSize
func (q *Query[T]) CreateInBatches(value any, batchSize int) (tx *gorm.DB) {
	tx = q.tx.CreateInBatches(value, batchSize)
	return
}

// Save updates value in database. If value doesn't contain a matching primary key, value is inserted.
func (q *Query[T]) Save(value any) (tx *gorm.DB) {
	tx = q.tx.Save(value)
	return
}

// First finds the first record ordered by primary key, matching given conditions conds
func (q *Query[T]) First(dest any, conds ...any) (tx *gorm.DB) {
	tx = q.tx.First(dest, conds...)
	return
}

// Take finds the first record returned by the database in no specified order, matching given conditions conds
func (q *Query[T]) Take(dest any, conds ...any) (tx *gorm.DB) {
	tx = q.tx.Take(dest, conds...)
	return
}

// Last finds the last record ordered by primary key, matching given conditions conds
func (q *Query[T]) Last(dest any, conds ...any) (tx *gorm.DB) {
	tx = q.tx.Last(dest, conds...)
	return
}

// Find finds all records matching given conditions conds
func (q *Query[T]) Find(dest any, conds ...any) *Query[T] {
	q.tx = q.tx.Find(dest, conds...)
	return q
}

// FindInBatches finds all records in batches of batchSize
func (q *Query[T]) FindInBatches(dest any, batchSize int, fc func(tx *gorm.DB, batch int) error) *gorm.DB {
	q.tx = q.tx.FindInBatches(dest, batchSize, fc)
	return q.tx
}

// FirstOrInit finds the first matching record, otherwise if not found initializes a new instance with given conds.
// Each conds must be a struct or map.
//
// FirstOrInit never modifies the database. It is often used with Assign and Attrs.
//
//	// assign an email if the record is not found
//	db.Where(User{Name: "non_existing"}).Attrs(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "non_existing", Email: "fake@fake.org"}
//
//	// assign email regardless of if record is found
//	db.Where(User{Name: "jinzhu"}).Assign(User{Email: "fake@fake.org"}).FirstOrInit(&user)
//	// user -> User{Name: "jinzhu", Age: 20, Email: "fake@fake.org"}
func (q *Query[T]) FirstOrInit(dest any, conds ...any) (tx *gorm.DB) {
	tx = q.tx.FirstOrInit(dest, conds...)
	return
}

// FirstOrCreate finds the first matching record, otherwise if not found creates a new instance with given conds.
// Each conds must be a struct or map.
//
// Using FirstOrCreate in conjunction with Assign will result in an update to the database even if the record exists.
//
//	// assign an email if the record is not found
//	result := db.Where(User{Name: "non_existing"}).Attrs(User{Email: "fake@fake.org"}).FirstOrCreate(&user)
//	// user -> User{Name: "non_existing", Email: "fake@fake.org"}
//	// result.RowsAffected -> 1
//
//	// assign email regardless of if record is found
//	result := db.Where(User{Name: "jinzhu"}).Assign(User{Email: "fake@fake.org"}).FirstOrCreate(&user)
//	// user -> User{Name: "jinzhu", Age: 20, Email: "fake@fake.org"}
//	// result.RowsAffected -> 1
func (q *Query[T]) FirstOrCreate(dest any, conds ...any) (tx *gorm.DB) {
	tx = q.tx.FirstOrCreate(dest, conds...)
	return
}

// Update updates column with value using callbacks. Reference: https://gorm.io/docs/update.html#Update-Changed-Fields
func (q *Query[T]) Update(column string, value any) (tx *gorm.DB) {
	tx = q.tx.Update(column, value)
	return
}

// Updates updates attributes using callbacks. values must be a struct or map. Reference: https://gorm.io/docs/update.html#Update-Changed-Fields
func (q *Query[T]) Updates(values any) *Query[T] {
	q.tx = q.tx.Updates(values)
	return q
}

func (q *Query[T]) UpdateColumn(column string, value any) (tx *gorm.DB) {
	tx = q.tx.UpdateColumn(column, value)
	return
}

func (q *Query[T]) UpdateColumns(values any) (tx *gorm.DB) {
	tx = q.tx.UpdateColumns(values)
	return
}

// Delete deletes value matching given conditions. If value contains primary key it is included in the conditions. If
// value includes a deleted_at field, then Delete performs a soft delete instead by setting deleted_at with the current
// time if null.
func (q *Query[T]) Delete(value any, conds ...any) (tx *gorm.DB) {
	tx = q.tx.Delete(value, conds...)
	return
}

func (q *Query[T]) Count(count *int64) *Query[T] {
	q.tx = q.tx.Count(count)
	return q
}

func (q *Query[T]) Row() *sql.Row {
	return q.tx.Row()
}

func (q *Query[T]) Rows() (*sql.Rows, error) {
	return q.tx.Rows()
}

// Scan scans selected value to the struct dest
func (q *Query[T]) Scan(dest any) *Query[T] {
	q.tx = q.tx.Scan(dest)
	return q
}

// Pluck queries a single column from a model, returning in the slice dest. E.g.:
//
//	var ages []int64
//	db.Model(&users).Pluck("age", &ages)
func (q *Query[T]) Pluck(column string, dest any) (tx *gorm.DB) {
	tx = q.tx.Pluck(column, dest)
	return
}

func (q *Query[T]) ScanRows(rows *sql.Rows, dest any) error {
	return q.tx.ScanRows(rows, dest)
}

// Connection uses a db connection to execute an arbitrary number of commands in fc. When finished, the connection is
// returned to the connection pool.
func (q *Query[T]) Connection(fc func(tx *gorm.DB) error) error {
	return q.tx.Connection(fc)
}

// Transaction start a transaction as a block, return error will rollback, otherwise to commit. Transaction executes an
// arbitrary number of commands in fc within a transaction. On success the changes are committed; if an error occurs
// they are rolled back.
func (q *Query[T]) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return q.tx.Transaction(fc, opts...)
}

// Begin begins a transaction with any transaction options opts
func (q *Query[T]) Begin(opts ...*sql.TxOptions) *Query[T] {
	q.tx = q.tx.Begin(opts...)
	return q
}

// Commit commits the changes in a transaction
func (q *Query[T]) Commit() *Query[T] {
	q.tx = q.tx.Commit()
	return q
}

// Rollback rollbacks the changes in a transaction
func (q *Query[T]) Rollback() *Query[T] {
	q.tx = q.tx.Rollback()
	return q
}

func (q *Query[T]) SavePoint(name string) *Query[T] {
	q.tx = q.tx.SavePoint(name)
	return q
}

func (q *Query[T]) RollbackTo(name string) *Query[T] {
	q.tx = q.tx.RollbackTo(name)
	return q
}

// Exec executes raw sql
func (q *Query[T]) Exec(sql string, values ...any) (tx *gorm.DB) {
	tx = q.tx.Exec(sql, values...)
	return
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
