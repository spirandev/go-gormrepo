package gormrepo

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/spirandev/go-gormrepo/gormrepo/internal/pkhelper"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *GenericRepository[T]) singleResult() (*T, error) {
	var entity T
	err := r.db.First(&entity).Error
	return &entity, err
}

func (r *GenericRepository[T]) listResult() (*[]T, error) {
	var entities []T
	err := r.db.Find(&entities).Error
	return &entities, err
}
func (r *GenericRepository[T]) Create(entity *T) *GenericRepository[T] {
	err := r.db.Create(entity).Error
	if err != nil {
		r.lastError = err
		return r
	}

	r.currentResult = entity
	return r
}

func (r *GenericRepository[T]) CreateWithPreload(entity *T, associations ...string) *GenericRepository[T] {
	err := r.db.Create(entity).Error
	if err != nil {
		r.lastError = err
		return r
	}

	pkName, pkValue, err := pkhelper.GetPrimaryKey(entity)
	if err != nil {
		r.currentResult = entity
		return r
	}

	var createdEntity T
	query := r.db

	for _, association := range associations {
		query = query.Preload(association)
	}

	if err := query.First(&createdEntity, fmt.Sprintf("%s = ?", pkName), pkValue).Error; err != nil {
		r.currentResult = entity
		return r
	}

	r.currentResult = &createdEntity
	return r
}

func (r *GenericRepository[T]) CreateWithAllAssociations(entity *T) *GenericRepository[T] {
	err := r.db.Create(entity).Error
	if err != nil {
		r.lastError = err
		return r
	}

	pkName, pkValue, err := pkhelper.GetPrimaryKey(entity)
	if err != nil {
		r.currentResult = entity
		return r
	}

	var createdEntity T
	if err := r.db.Preload(clause.Associations).First(&createdEntity, fmt.Sprintf("%s = ?", pkName), pkValue).Error; err != nil {
		r.currentResult = entity
		return r
	}

	r.currentResult = &createdEntity
	return r
}

func (r *GenericRepository[T]) CreateBatch(entities *[]T) *GenericRepository[T] {
	err := r.db.Create(entities).Error
	if err != nil {
		r.lastError = err
		return r
	}

	r.currentSlice = entities
	return r
}

func (r *GenericRepository[T]) Update(entity *T) *GenericRepository[T] {
	err := r.db.Save(entity).Error
	if err != nil {
		r.lastError = err
		return r
	}

	r.currentResult = entity
	return r
}

func (r *GenericRepository[T]) UpdateWithPreload(entity *T, associations ...string) *GenericRepository[T] {
	err := r.db.Save(entity).Error
	if err != nil {
		r.lastError = err
		return r
	}
	pkName, pkValue, err := pkhelper.GetPrimaryKey(entity)
	if err != nil {
		r.currentResult = entity
		return r
	}
	var updatedEntity T
	query := r.db
	for _, association := range associations {
		query = query.Preload(association)
	}
	if err := query.First(&updatedEntity, fmt.Sprintf("%s = ?", pkName), pkValue).Error; err != nil {
		r.currentResult = entity
		return r
	}
	r.currentResult = &updatedEntity
	return r
}

func (r *GenericRepository[T]) UpdateFields(entity *T, fields map[string]interface{}) *GenericRepository[T] {
	pkName, pkValue, err := pkhelper.GetPrimaryKey(entity)
	if err != nil {
		r.lastError = err
		return r
	}
	err = r.db.Model(entity).Where(fmt.Sprintf("%s = ?", pkName), pkValue).Updates(fields).Error
	if err != nil {
		r.lastError = err
		return r
	}
	r.currentResult = entity
	return r
}

func getID(entity interface{}) interface{} {
	val := reflect.ValueOf(entity).Elem()
	return val.FieldByName("ID").Interface()
}

func (r *GenericRepository[T]) Delete(id int64) *GenericRepository[T] {
	err := r.db.Delete(new(T), id).Error
	if err != nil {
		r.lastError = err
	}
	return r
}

func (r *GenericRepository[T]) DeleteEntity(entity *T) *GenericRepository[T] {
	err := r.db.Delete(entity).Error
	if err != nil {
		r.lastError = err
	}
	return r
}

func (r *GenericRepository[T]) DeleteBatch(entities *[]T) *GenericRepository[T] {
	err := r.db.Delete(entities).Error
	if err != nil {
		r.lastError = err
	}
	return r
}

func (r *GenericRepository[T]) FindByID(id int64) *GenericRepository[T] {
	return r.Where("id = ?", id)
}

func (r *GenericRepository[T]) FindFirst() *GenericRepository[T] {
	return r.Limit(1)
}

func (r *GenericRepository[T]) FindAll() *GenericRepository[T] {
	// No need to do anything here, just return the repository
	// The query will be executed when First(), Get() or One() are called
	return r
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

func (r *GenericRepository[T]) CreateWithContext(ctx context.Context, entity *T) *GenericRepository[T] {
	contextRepo := &GenericRepository[T]{
		db:             r.db.WithContext(ctx),
		projection:     r.projection,
		projectionMode: r.projectionMode,
	}
	return contextRepo.Create(entity)
}

func (r *GenericRepository[T]) FindByIDWithContext(ctx context.Context, id int64) *GenericRepository[T] {
	contextRepo := &GenericRepository[T]{
		db:             r.db.WithContext(ctx),
		projection:     r.projection,
		projectionMode: r.projectionMode,
	}
	return contextRepo.Where("id = ?", id)
}

func (r *GenericRepository[T]) FindOne(filters map[string]interface{}) *GenericRepository[T] {
	// Apply filters to the existing db (which may already have preloads/joins configured)
	query := r.db
	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}

	// Update the db to preserve the configuration for the next operations
	r.db = query

	// Execute query and store result for chaining
	var entity T
	err := r.db.First(&entity).Error
	if err != nil {
		r.lastError = err
		return r
	}

	r.currentResult = &entity
	return r
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
	return r.singleResult()
}

func (r *GenericRepository[T]) Get() (*[]T, error) {
	return r.listResult()
}

func (r *GenericRepository[T]) One() (*T, error) {
	r.db = r.db.Limit(1)
	return r.singleResult()
}

// ==================== PROJECTION METHODS ====================

// ProjectToDTO configures the repository to return only the DTO
// RETURNS: The specified DTO, not the complete entity
// Usage: repo.ProjectToDTO(UserBasicInfo{}).FindOne(filters).Project()
// AUTOMATICALLY DETECTS STRUCTS: If the DTO has struct fields, preloads will be applied automatically
func (r *GenericRepository[T]) ProjectToDTO(dtoInterface interface{}) *GenericRepository[T] {
	// Create a new repository instance to not affect the original
	newRepo := &GenericRepository[T]{
		db:             r.db,
		projection:     dtoInterface,
		projectionMode: "dto",
		currentResult:  r.currentResult,
		currentSlice:   r.currentSlice,
		lastError:      r.lastError,
	}

	// Check if the DTO has struct fields and apply preloads automatically
	if hasStructFields(dtoInterface) {
		// Apply preloads for nested structs
		preloads := extractPreloadsFromDTO(dtoInterface)

		for _, preload := range preloads {
			newRepo.db = newRepo.db.Preload(preload)
		}

		// When there are preloads, let GORM manage field selection automatically
		// Using Select() would break the preloads functionality
	} else {
		// Standard projection without preloads
		fields := createProjectionFromDTO(dtoInterface)
		if len(fields) > 0 {
			selectFields := strings.Join(fields, ", ")
			newRepo.db = newRepo.db.Select(selectFields)
		}
	}

	return newRepo
}

// Project converts current result to real DTO using configured projection
func (r *GenericRepository[T]) Project() (interface{}, error) {
	if r.projection == nil {
		return nil, fmt.Errorf("no projection configured - use ProjectToDTO() first")
	}

	// Use the stored current result instead of making a new query
	if r.currentResult == nil {
		return nil, fmt.Errorf("no current result available - execute a query first (FindOne, FindByID, etc.)")
	}

	return mapEntityToDTO(r.currentResult, r.projection)
}

// ==================== HELPER FUNCTIONS ====================

// hasStructFields checks if the DTO has any struct fields (for auto-preload)
func hasStructFields(dtoInterface interface{}) bool {
	dtoType := reflect.TypeOf(dtoInterface)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	for i := 0; i < dtoType.NumField(); i++ {
		field := dtoType.Field(i)

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// If it's a struct and not a basic type, it should be preloaded
		if fieldType.Kind() == reflect.Struct && !isBasicType(fieldType) {
			return true
		}
	}

	return false
}

// extractPreloadsFromDTO extracts preload associations for nested structs
func extractPreloadsFromDTO(dtoInterface interface{}) []string {
	var preloads []string

	dtoType := reflect.TypeOf(dtoInterface)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	for i := 0; i < dtoType.NumField(); i++ {
		field := dtoType.Field(i)

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// If it's a struct and not a basic type, add it to preloads
		if fieldType.Kind() == reflect.Struct && !isBasicType(fieldType) {
			// Use the field name as the association name (can be customized with preload tag)
			preloadName := field.Tag.Get("preload")
			if preloadName == "" {
				preloadName = field.Name
			}
			preloads = append(preloads, preloadName)
		}
	}

	return preloads
}

// extractMainTableFields extracts only fields that belong to the main table (no struct fields)
func extractMainTableFields(dtoInterface interface{}) []string {
	var fields []string

	dtoType := reflect.TypeOf(dtoInterface)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	for i := 0; i < dtoType.NumField(); i++ {
		field := dtoType.Field(i)

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Only include basic type fields (not structs)
		if isBasicType(fieldType) {
			columnName := getColumnName(field)
			fields = append(fields, columnName)
		}
	}

	return fields
}

// isBasicType checks if a type is a basic type (not a struct)
func isBasicType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	}
	return false
}

// ==================== HELPER METHODS ====================

// HasError checks if there is any stored error
func (r *GenericRepository[T]) HasError() bool {
	return r.lastError != nil
}

// Error returns the last error that occurred
func (r *GenericRepository[T]) Error() error {
	return r.lastError
}

// Result returns the current result and last error
func (r *GenericRepository[T]) Result() (*T, error) {
	return r.currentResult, r.lastError
}

// Results returns the current slice result and last error
func (r *GenericRepository[T]) Results() (*[]T, error) {
	return r.currentSlice, r.lastError
}

// Execute finalizes the operation and returns only the error (if any)
// Useful when you only care about success/failure, not the result
func (r *GenericRepository[T]) Execute() error {
	return r.lastError
}

// ==================== PROJECTION COMPATIBILITY METHODS ====================

// ProjectEntity converts an entity to real DTO (static method)
// Usage: dto, err := repo.ProjectEntity(user, UserBasicInfo{})
func (r *GenericRepository[T]) ProjectEntity(entity *T, dtoInterface interface{}) (interface{}, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity cannot be nil")
	}

	return mapEntityToDTO(entity, dtoInterface)
}

// ProjectEntitySlice converts a list of entities to slice of DTOs (static method)
// Usage: dtos, err := repo.ProjectEntitySlice(users, UserBasicInfo{})
func (r *GenericRepository[T]) ProjectEntitySlice(entities *[]T, dtoInterface interface{}) (interface{}, error) {
	if entities == nil {
		return nil, fmt.Errorf("entity slice cannot be nil")
	}

	if len(*entities) == 0 {
		// Return empty slice of DTO type
		dtoType := reflect.TypeOf(dtoInterface)
		if dtoType.Kind() == reflect.Ptr {
			dtoType = dtoType.Elem()
		}
		sliceType := reflect.SliceOf(dtoType)
		return reflect.MakeSlice(sliceType, 0, 0).Interface(), nil
	}

	// Get DTO type
	dtoType := reflect.TypeOf(dtoInterface)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	// Create DTO slice
	sliceType := reflect.SliceOf(dtoType)
	resultSlice := reflect.MakeSlice(sliceType, 0, len(*entities))

	// Convert each entity
	for _, entity := range *entities {
		dto, err := mapEntityToDTO(&entity, dtoInterface)
		if err != nil {
			return nil, fmt.Errorf("error converting entity: %w", err)
		}

		// Add to slice
		dtoValue := reflect.ValueOf(dto)
		if dtoValue.Kind() == reflect.Ptr {
			dtoValue = dtoValue.Elem()
		}
		resultSlice = reflect.Append(resultSlice, dtoValue)
	}

	return resultSlice.Interface(), nil
}

// ProjectSlice converts current result (currentSlice) to slice of DTOs using configured projection
// Usage: dtos, err := repo.ProjectToDTO(UserBasicInfo{}).FindAll().ProjectSlice()
func (r *GenericRepository[T]) ProjectSlice() (interface{}, error) {
	if r.lastError != nil {
		return nil, r.lastError
	}

	if r.currentSlice == nil {
		return nil, fmt.Errorf("no slice result available for conversion")
	}

	if r.projection == nil {
		return nil, fmt.Errorf("no projection configured - use ProjectToDTO() first")
	}

	if len(*r.currentSlice) == 0 {
		// Return empty slice of DTO type
		dtoType := reflect.TypeOf(r.projection)
		if dtoType.Kind() == reflect.Ptr {
			dtoType = dtoType.Elem()
		}
		sliceType := reflect.SliceOf(dtoType)
		return reflect.MakeSlice(sliceType, 0, 0).Interface(), nil
	}

	// Get DTO type
	dtoType := reflect.TypeOf(r.projection)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	// Create DTO slice
	sliceType := reflect.SliceOf(dtoType)
	resultSlice := reflect.MakeSlice(sliceType, 0, len(*r.currentSlice))

	// Convert each entity
	for _, entity := range *r.currentSlice {
		dto, err := mapEntityToDTO(&entity, r.projection)
		if err != nil {
			return nil, fmt.Errorf("error converting entity: %w", err)
		}

		// Add to slice
		dtoValue := reflect.ValueOf(dto)
		if dtoValue.Kind() == reflect.Ptr {
			dtoValue = dtoValue.Elem()
		}
		resultSlice = reflect.Append(resultSlice, dtoValue)
	}

	return resultSlice.Interface(), nil
}
