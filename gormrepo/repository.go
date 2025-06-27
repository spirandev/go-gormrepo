package gormrepo

import (
	"context"

	"gorm.io/gorm"
)

type BaseRepository[T any] interface {
	Begin() (*gorm.DB, error)
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error

	// Fluent methods - return *GenericRepository[T] for chaining
	Create(entity *T) *GenericRepository[T]
	CreateWithPreload(entity *T, associations ...string) *GenericRepository[T]
	CreateWithAllAssociations(entity *T) *GenericRepository[T]
	CreateBatch(entities *[]T) *GenericRepository[T]

	Update(entity *T) *GenericRepository[T]
	UpdateWithPreload(entity *T, fields ...string) *GenericRepository[T]
	UpdateFields(entity *T, fields map[string]interface{}) *GenericRepository[T]

	Delete(id int64) *GenericRepository[T]
	DeleteEntity(entity *T) *GenericRepository[T]
	DeleteBatch(entities *[]T) *GenericRepository[T]

	FindByID(id int64) *GenericRepository[T]
	FindAll() *GenericRepository[T]

	Preload(associations ...string) *GenericRepository[T]
	WithJoins(joins ...string) *GenericRepository[T]

	Where(query interface{}, args ...interface{}) *GenericRepository[T]
	Order(value interface{}) *GenericRepository[T]
	Count(filters map[string]interface{}) (int64, error)
	Exists(filters map[string]interface{}) (bool, error)

	CreateWithContext(ctx context.Context, entity *T) *GenericRepository[T]
	FindByIDWithContext(ctx context.Context, id int64) *GenericRepository[T]
	FindOne(filters map[string]interface{}) *GenericRepository[T]

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

	// Finalizer methods - execute the query and return the result
	First() (*T, error) // Returns first entity found
	Get() (*[]T, error) // Returns slice of entities
	One() (*T, error)   // Returns one entity or error if not exactly one found
	// FindFirst() (*T, error) // Alias for First() for compatibility

	// Projection methods - return repository configured to use projection
	// ProjectTo(dtoInterface interface{}) *GenericRepository[T]
	// ProjectToPartial(dtoInterface interface{}) *GenericRepository[T] // Returns entity with only projection fields filled
	ProjectToDTO(dtoInterface interface{}) *GenericRepository[T] // Returns only DTO, not complete entity

	// Conversion methods for real DTO - works with repository current result
	Project() (interface{}, error)      // Converts currentResult to real DTO using configured projection
	ProjectSlice() (interface{}, error) // Converts currentSlice to slice of real DTOs using configured projection

	// Static conversion methods (for compatibility)
	ProjectEntity(entity *T, dtoInterface interface{}) (interface{}, error)
	ProjectEntitySlice(entities *[]T, dtoInterface interface{}) (interface{}, error)

	// Helper methods to check state
	HasError() bool
	Error() error
	Result() (*T, error)    // Returns currentResult and lastError
	Results() (*[]T, error) // Returns currentSlice and lastError
	Execute() error         // Finalizes operation and returns only error
}
type GenericRepository[T any] struct {
	db             *gorm.DB
	projection     interface{} // Stores DTO type for projection
	projectionMode string      // "full", "partial", "dto"
	currentResult  *T          // Stores current result for chaining
	currentSlice   *[]T        // Stores slice of results for chaining
	lastError      error       // Stores last error that occurred
}

func New[T any](db *gorm.DB) *GenericRepository[T] {
	if db == nil {
		panic("database not initialized")
	}
	return &GenericRepository[T]{
		db:            db,
		currentResult: nil,
		currentSlice:  nil,
		lastError:     nil,
	}
}
