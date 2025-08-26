package gorms

import (
	"time"

	"gorm.io/gorm"
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

// QueryAction is [gorm.DB] condition bridge
type QueryAction interface {
	Condition(tx *gorm.DB)
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

type TinyModel[T any] struct {
	ID        T         `gorm:"column:id;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

type IDModel[T any] struct {
	ID T `gorm:"column:id;primaryKey" json:"id"`
}
