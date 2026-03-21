package orms

import (
	"database/sql"

	"github.com/natholdallas/natools4go/maths"
	"gorm.io/gorm"
)

// Create inserts value, returning the inserted data's primary key in value's id
func (q *Query[T]) Create(value any) error {
	return q.db.Create(value).Error
}

// CreateInBatches inserts value in batches of batchSize
func (q *Query[T]) CreateInBatches(value any, batchSize int) error {
	return q.db.CreateInBatches(value, batchSize).Error
}

// Save updates value in database. If value doesn't contain a matching primary key, value is inserted.
func (q *Query[T]) Save(value any) error {
	return q.db.Save(value).Error
}

// First finds the first record ordered by primary key, matching given conditions conds
func (q *Query[T]) First(conds ...any) (T, error) {
	var v T
	err := q.db.First(&v, conds...).Error
	return v, err
}

// IFirst finds the first record ordered by primary key
// matching given conditions conds, if not found it will return a zero value
func (q *Query[T]) IFirst(conds ...any) T {
	var v T
	q.db.First(&v, conds...)
	return v
}

// Take finds the first record returned by the database in no specified order, matching given conditions conds
func (q *Query[T]) Take(conds ...any) (T, error) {
	var v T
	err := q.db.Take(&v, conds...).Error
	return v, err
}

// ITake finds the first record returned by the database in no specified order
// matching given conditions conds, if not found it will return a zero value
func (q *Query[T]) ITake(conds ...any) T {
	var v T
	q.db.Take(&v, conds...)
	return v
}

// Last finds the last record ordered by primary key, matching given conditions conds
func (q *Query[T]) Last(conds ...any) (T, error) {
	var v T
	err := q.db.Last(&v, conds...).Error
	return v, err
}

// ILast finds the last record ordered by primary key
// matching given conditions conds, if not found it will return a zero value
func (q *Query[T]) ILast(conds ...any) T {
	var v T
	q.db.Last(&v, conds...)
	return v
}

// Find finds all records matching given conditions conds
func (q *Query[T]) Find(conds ...any) ([]T, error) {
	v := []T{}
	err := q.db.Find(&v, conds...).Error
	return v, err
}

// IFind finds all records
func (q *Query[T]) IFind(conds ...any) []T {
	var v []T
	q.db.Find(&v, conds...)
	return v
}

// FindInBatches finds all records in batches of batchSize
func (q *Query[T]) FindInBatches(batchSize int, fc func(tx *gorm.DB, batch int) error) ([]T, error) {
	var v []T
	err := q.db.FindInBatches(&v, batchSize, fc).Error
	return v, err
}

// IFindInBatches finds all records in batches of batchSize
func (q *Query[T]) IFindInBatches(batchSize int, fc func(tx *gorm.DB, batch int) error) []T {
	var v []T
	q.db.FindInBatches(&v, batchSize, fc)
	return v
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
	tx = q.db.FirstOrInit(dest, conds...)
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
	tx = q.db.FirstOrCreate(dest, conds...)
	return
}

// Update updates column with value using callbacks. Reference: https://gorm.io/docs/update.html#Update-Changed-Fields
func (q *Query[T]) Update(column string, value any) error {
	return q.db.Update(column, value).Error
}

// Updates updates attributes using callbacks. values must be a struct or map. Reference: https://gorm.io/docs/update.html#Update-Changed-Fields
func (q *Query[T]) Updates(values any) error {
	return q.db.Updates(values).Error
}

func (q *Query[T]) UpdateColumn(column string, value any) (tx *gorm.DB) {
	tx = q.db.UpdateColumn(column, value)
	return
}

func (q *Query[T]) UpdateColumns(values any) (tx *gorm.DB) {
	tx = q.db.UpdateColumns(values)
	return
}

// Delete deletes value matching given conditions. If value contains primary key it is included in the conditions. If
// value includes a deleted_at field, then Delete performs a soft delete instead by setting deleted_at with the current
// time if null.
func (q *Query[T]) Delete(conds ...any) error {
	return q.db.Delete(new(T), conds...).Error
}

func (q *Query[T]) Count() int64 {
	var count int64
	q.db.Count(&count)
	return count
}

func (q *Query[T]) Row() *sql.Row {
	return q.db.Row()
}

func (q *Query[T]) Rows() (*sql.Rows, error) {
	return q.db.Rows()
}

// Scan scans selected value to the struct dest
func (q *Query[T]) Scan(dest any) *Query[T] {
	q.db = q.db.Scan(dest)
	return q
}

// Pluck queries a single column from a model, returning in the slice dest. E.g.:
//
//	var ages []int64
//	db.Model(&users).Pluck("age", &ages)
func (q *Query[T]) Pluck(column string, dest any) (tx *gorm.DB) {
	tx = q.db.Pluck(column, dest)
	return
}

func (q *Query[T]) ScanRows(rows *sql.Rows, dest any) error {
	return q.db.ScanRows(rows, dest)
}

// Connection uses a db connection to execute an arbitrary number of commands in fc. When finished, the connection is
// returned to the connection pool.
func (q *Query[T]) Connection(fc func(tx *gorm.DB) error) error {
	return q.db.Connection(fc)
}

// Transaction start a transaction as a block, return error will rollback, otherwise to commit. Transaction executes an
// arbitrary number of commands in fc within a transaction. On success the changes are committed; if an error occurs
// they are rolled back.
func (q *Query[T]) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(fc, opts...)
}

// Begin begins a transaction with any transaction options opts
func (q *Query[T]) Begin(opts ...*sql.TxOptions) *Query[T] {
	q.db = q.db.Begin(opts...)
	return q
}

// Commit commits the changes in a transaction
func (q *Query[T]) Commit() *Query[T] {
	q.db = q.db.Commit()
	return q
}

// Rollback rollbacks the changes in a transaction
func (q *Query[T]) Rollback() *Query[T] {
	q.db = q.db.Rollback()
	return q
}

func (q *Query[T]) SavePoint(name string) *Query[T] {
	q.db = q.db.SavePoint(name)
	return q
}

func (q *Query[T]) RollbackTo(name string) *Query[T] {
	q.db = q.db.RollbackTo(name)
	return q
}

// Exec executes raw sql
func (q *Query[T]) Exec(sql string, values ...any) (tx *gorm.DB) {
	tx = q.db.Exec(sql, values...)
	return
}

// Paginate executes the query with pagination and returns a Page[T].
// It automatically sets the model to T if not already defined.
func (q *Query[T]) Paginate(pagination Pagination) (Page[T], *gorm.DB) {
	if q.db.Statement.Model == nil {
		model := new(T)
		q.db = q.db.Model(model)
	}
	total := int64(0)
	content := []T{}
	q.db = q.db.Count(&total).Scopes(pagination.Scope).Find(&content)
	page := maths.DivCeil(total, int64(pagination.Size))
	return Page[T]{total, page, content}, q.db
}

// IPaginate executes the query with pagination and returns a Page[T].
// It automatically sets the model to T if not already defined.
func (q *Query[T]) IPaginate(pagination Pagination) Page[T] {
	if q.db.Statement.Model == nil {
		model := new(T)
		q.db = q.db.Model(model)
	}
	total := int64(0)
	content := []T{}
	q.db = q.db.Count(&total).Scopes(pagination.Scope).Find(&content)
	page := maths.DivCeil(total, int64(pagination.Size))
	return Page[T]{total, page, content}
}
