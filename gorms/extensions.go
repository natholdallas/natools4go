package gorms

import (
	"time"

	"gorm.io/gorm"
)

type GormScope = func(*gorm.DB) *gorm.DB

// Getter dto transform to database model
type Getter[T any] interface {
	Get() T
}

// Setter use dto sets database model's value
type Setter[T any] interface {
	Set(t *T)
}

// QueryAction is [gorm.DB] condition bridge
type QueryAction interface {
	Condition(tx *gorm.DB)
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
