package orms

import (
	"gorm.io/gorm"
)

// Query is a fluent, generic wrapper around [gorm.DB] for cleaner query construction.
type Query[T any] struct {
	db *gorm.DB
}

// Q initializes a Query wrapper without setting a default model.
func Q[T any](db *gorm.DB) *Query[T] {
	return &Query[T]{db}
}

// QE initializes a Query wrapper and sets the generic type T as the GORM model.
func QE[T any](db *gorm.DB) *Query[T] {
	return &Query[T]{db: db.Model(new(T))}
}

// QM initializes a Query wrapper where T is the result type and M is the database model.
func QM[T, M any](db *gorm.DB) *Query[T] {
	return &Query[T]{db: db.Model(new(M))}
}

// QT initializes a Query wrapper targeting a specific table name.
func QT[T any](db *gorm.DB, name string, args ...any) *Query[T] {
	return &Query[T]{db: db.Table(name, args...)}
}
