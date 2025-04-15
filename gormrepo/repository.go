package gormrepo

import (
	"context"

	"gorm.io/gorm"
)

type BaseRepository[T any] interface {
	Begin() (*gorm.DB, error)
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error

	Create(entity *T) error
	CreateBatch(entities *[]T) error
	Update(entity *T) error
	Delete(id int64) error
	DeleteEntity(entity *T) error
	DeleteBatch(entities *[]T) error

	FindByID(id int64) (*T, error)
	FindFirst() (*T, error)
	FindAll() (*[]T, error)

	Preload(associations ...string) *GenericRepository[T]
	WithJoins(joins ...string) *GenericRepository[T]

	Where(query interface{}, args ...interface{}) *GenericRepository[T]
	Order(value interface{}) *GenericRepository[T]
	Count(filters map[string]interface{}) (int64, error)
	Exists(filters map[string]interface{}) (bool, error)

	CreateWithContext(ctx context.Context, entity *T) error
	FindByIDWithContext(ctx context.Context, id int64) (*T, error)
	FindOne(filters map[string]interface{}) (*T, error)

	Limit(limit int) *GenericRepository[T]
	Offset(offset int) *GenericRepository[T]
	Paginate(page, pageSize int) *GenericRepository[T]

	Transaction(fn func(tx *GenericRepository[T]) error) error
	WithDB(db *gorm.DB) *GenericRepository[T]

	Select(query interface{}, args ...interface{}) *GenericRepository[T]
	Group(name string) *GenericRepository[T]
	Having(query interface{}, args ...interface{}) *GenericRepository[T]
	Or(query interface{}, args ...interface{}) *GenericRepository[T]
	Not(query interface{}, args ...interface{}) *GenericRepository[T]

	First() (*T, error)
	Get() (*[]T, error)
	One() (*T, error)
}

type GenericRepository[T any] struct {
	db *gorm.DB
}

func New[T any](db *gorm.DB) *GenericRepository[T] {
	if db == nil {
		panic("banco de dados não inicializado")
	}
	return &GenericRepository[T]{db: db}
}
