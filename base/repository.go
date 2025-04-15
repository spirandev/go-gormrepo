package gormrepo

import (
	"gorm.io/gorm"
)

type BaseRepository[T any] interface {
}

type GenericRepository[T any] struct {
	db     *gorm.DB
	entity *T
}
