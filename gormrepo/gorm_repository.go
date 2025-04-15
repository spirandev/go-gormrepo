package gormrepo

import (
	"context"

	"gorm.io/gorm"
)

func (r *GenericRepository[T]) Begin() (*gorm.DB, error) {
	return r.db.Begin(), nil
}

func (r *GenericRepository[T]) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

func (r *GenericRepository[T]) Rollback(tx *gorm.DB) error {
	return tx.Rollback().Error
}

func (r *GenericRepository[T]) SingleResult() (*T, error) {
	var entity T
	err := r.db.First(&entity).Error
	return &entity, err
}

func (r *GenericRepository[T]) ListResult() (*[]T, error) {
	var entities []T
	err := r.db.Find(&entities).Error
	return &entities, err
}

func (r *GenericRepository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

func (r *GenericRepository[T]) CreateBatch(entities *[]T) error {
	return r.db.Create(entities).Error
}

func (r *GenericRepository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

func (r *GenericRepository[T]) Delete(id int64) error {
	return r.db.Delete(new(T), id).Error
}

func (r *GenericRepository[T]) DeleteEntity(entity *T) error {
	return r.db.Delete(entity).Error
}

func (r *GenericRepository[T]) DeleteBatch(entities *[]T) error {
	return r.db.Delete(entities).Error
}

func (r *GenericRepository[T]) FindByID(id int64) (*T, error) {
	return r.Where("id = ?", id).First()
}

func (r *GenericRepository[T]) FindFirst() (*T, error) {
	return r.Limit(1).First()
}

func (r *GenericRepository[T]) FindAll() (*[]T, error) {
	return r.ListResult()
}

func (r *GenericRepository[T]) Preload(associations ...string) *GenericRepository[T] {
	for _, association := range associations {
		r.db = r.db.Preload(association)
	}
	return r
}

func (r *GenericRepository[T]) WithJoins(joins ...string) *GenericRepository[T] {
	for _, join := range joins {
		r.db = r.db.Joins(join)
	}
	return r
}

func (r *GenericRepository[T]) Where(query interface{}, args ...interface{}) *GenericRepository[T] {
	r.db = r.db.Where(query, args...)
	return r
}

func (r *GenericRepository[T]) Order(value interface{}) *GenericRepository[T] {
	r.db = r.db.Order(value)
	return r
}

func (r *GenericRepository[T]) Count(filters map[string]interface{}) (int64, error) {
	filterRepo := &GenericRepository[T]{db: r.db.Model(new(T))}
	for k, v := range filters {
		filterRepo = filterRepo.Where(k+" = ?", v)
	}
	var count int64
	err := filterRepo.db.Count(&count).Error
	return count, err
}

func (r *GenericRepository[T]) Exists(filters map[string]interface{}) (bool, error) {
	count, err := r.Count(filters)
	return count > 0, err
}

func (r *GenericRepository[T]) CreateWithContext(ctx context.Context, entity *T) error {
	contextRepo := &GenericRepository[T]{db: r.db.WithContext(ctx)}
	return contextRepo.db.Create(entity).Error
}

func (r *GenericRepository[T]) FindByIDWithContext(ctx context.Context, id int64) (*T, error) {
	contextRepo := &GenericRepository[T]{db: r.db.WithContext(ctx)}
	return contextRepo.Where("id = ?", id).First()
}

func (r *GenericRepository[T]) FindOne(filters map[string]interface{}) (*T, error) {
	filterRepo := &GenericRepository[T]{db: r.db}
	for k, v := range filters {
		filterRepo = filterRepo.Where(k+" = ?", v)
	}
	return filterRepo.First()
}

func (r *GenericRepository[T]) Limit(limit int) *GenericRepository[T] {
	r.db = r.db.Limit(limit)
	return r
}

func (r *GenericRepository[T]) Offset(offset int) *GenericRepository[T] {
	r.db = r.db.Offset(offset)
	return r
}

func (r *GenericRepository[T]) Paginate(page, pageSize int) *GenericRepository[T] {
	offset := (page - 1) * pageSize
	return r.Offset(offset).Limit(pageSize)
}

func (r *GenericRepository[T]) Transaction(fn func(tx *GenericRepository[T]) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &GenericRepository[T]{db: tx}
		return fn(txRepo)
	})
}

func (r *GenericRepository[T]) WithDB(db *gorm.DB) *GenericRepository[T] {
	return &GenericRepository[T]{db: db}
}

func (r *GenericRepository[T]) Select(query interface{}, args ...interface{}) *GenericRepository[T] {
	r.db = r.db.Select(query, args...)
	return r
}

func (r *GenericRepository[T]) Group(name string) *GenericRepository[T] {
	r.db = r.db.Group(name)
	return r
}

func (r *GenericRepository[T]) Having(query interface{}, args ...interface{}) *GenericRepository[T] {
	r.db = r.db.Having(query, args...)
	return r
}

func (r *GenericRepository[T]) Or(query interface{}, args ...interface{}) *GenericRepository[T] {
	r.db = r.db.Or(query, args...)
	return r
}

func (r *GenericRepository[T]) Not(query interface{}, args ...interface{}) *GenericRepository[T] {
	r.db = r.db.Not(query, args...)
	return r
}

func (r *GenericRepository[T]) First() (*T, error) {
	return r.SingleResult()
}

func (r *GenericRepository[T]) Get() (*[]T, error) {
	return r.ListResult()
}

func (r *GenericRepository[T]) One() (*T, error) {
	r.db = r.db.Limit(1)
	return r.SingleResult()
}
