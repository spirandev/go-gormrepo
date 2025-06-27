package gormrepo

import (
	"fmt"
	"reflect"
	"strings"
)

// Projection interface to define projection contracts
type Projection interface {
	GetSelectFields() []string
	GetTableName() string
}

// BaseProjection base struct for projections
type BaseProjection struct {
	selectFields []string
	tableName    string
}

func (p *BaseProjection) GetSelectFields() []string {
	return p.selectFields
}

func (p *BaseProjection) GetTableName() string {
	return p.tableName
}

// ProjectionBuilder to build projections dynamically
type ProjectionBuilder struct {
	fields    []string
	tableName string
}

func NewProjectionBuilder(tableName string) *ProjectionBuilder {
	return &ProjectionBuilder{
		fields:    make([]string, 0),
		tableName: tableName,
	}
}

func (pb *ProjectionBuilder) AddField(field string) *ProjectionBuilder {
	pb.fields = append(pb.fields, field)
	return pb
}

func (pb *ProjectionBuilder) AddFields(fields ...string) *ProjectionBuilder {
	pb.fields = append(pb.fields, fields...)
	return pb
}

func (pb *ProjectionBuilder) AddFieldWithAlias(field, alias string) *ProjectionBuilder {
	pb.fields = append(pb.fields, fmt.Sprintf("%s AS %s", field, alias))
	return pb
}

func (pb *ProjectionBuilder) Build() Projection {
	return &BaseProjection{
		selectFields: pb.fields,
		tableName:    pb.tableName,
	}
}

// AutoProjection creates projection automatically based on struct
func AutoProjection[T any]() Projection {
	var entity T
	entityType := reflect.TypeOf(entity)

	// If it's a pointer, get the base type
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	var fields []string
	tableName := getTableName(entityType)

	// Go through all struct fields
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)

		// Skip private fields
		if !field.IsExported() {
			continue
		}

		// Check if it has gorm tag
		gormTag := field.Tag.Get("gorm")
		if gormTag == "-" {
			continue
		}

		// Extract column name from gorm tag or use field name
		columnName := getColumnName(field)
		fields = append(fields, columnName)
	}

	return &BaseProjection{
		selectFields: fields,
		tableName:    tableName,
	}
}

// CreateProjectionFromStruct creates a projection automatically from a DTO struct
// Similar to Spring Boot, where you define an interface/DTO and it maps automatically
func CreateProjectionFromStruct[S any, D any](sourceEntity S) Projection {
	sourceType := reflect.TypeOf(sourceEntity)
	var destEntity D
	destType := reflect.TypeOf(destEntity)

	// Remove pointers if they exist
	if sourceType.Kind() == reflect.Ptr {
		sourceType = sourceType.Elem()
	}
	if destType.Kind() == reflect.Ptr {
		destType = destType.Elem()
	}

	var fields []string
	tableName := getTableName(sourceType)

	// Go through all DTO fields
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)

		if !field.IsExported() {
			continue
		}

		// Extract column name
		columnName := getColumnName(field)
		fields = append(fields, columnName)
	}

	return &BaseProjection{
		selectFields: fields,
		tableName:    tableName,
	}
}

// ProjectWith cria uma projeção baseada em uma struct DTO
// Uso: repo.ProjectWith(UserBasicInfo{}).Find(&results)
func ProjectWith[T any](dtoStruct T) Projection {
	dtoType := reflect.TypeOf(dtoStruct)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	var fields []string

	// Extract fields from DTO struct
	for i := 0; i < dtoType.NumField(); i++ {
		field := dtoType.Field(i)

		if !field.IsExported() {
			continue
		}

		// Check tags for mapping
		columnName := getColumnNameFromDTO(field)
		fields = append(fields, columnName)
	}

	return &BaseProjection{
		selectFields: fields,
		tableName:    "", // Will be determined by context
	}
}

// Helper functions for reflection and mapping

// getTableName extracts table name from struct
func getTableName(entityType reflect.Type) string {
	// Try to get from TableName() interface if it exists
	entityValue := reflect.New(entityType).Interface()
	if tableNamer, ok := entityValue.(interface{ TableName() string }); ok {
		return tableNamer.TableName()
	}

	// Otherwise, use GORM convention (snake_case of struct name + s)
	return toSnakeCase(entityType.Name()) + "s"
}

// getColumnName extracts column name from struct field
func getColumnName(field reflect.StructField) string {
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		// Parse gorm tag to extract column name
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
		}
	}

	// If no column tag, use field name in snake_case
	return toSnakeCase(field.Name)
}

// getColumnNameFromDTO extracts column name from a DTO struct field
func getColumnNameFromDTO(field reflect.StructField) string {
	// 1. Check projection tag first
	if projection := field.Tag.Get("projection"); projection != "" {
		return projection
	}

	// 2. Check column tag
	if gormTag := field.Tag.Get("gorm"); gormTag != "" {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
		}
	}

	// 3. Check json tag
	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		return strings.Split(jsonTag, ",")[0]
	}

	// 4. Use field name in snake_case
	return toSnakeCase(field.Name)
}

// toSnakeCase converts string to snake_case
func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// createProjectionFromDTO creates a projection based on a DTO struct
func createProjectionFromDTO(dtoInterface interface{}) []string {
	dtoType := reflect.TypeOf(dtoInterface)

	// If it's a pointer, get the base type
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	var fields []string

	// Go through all DTO struct fields
	for i := 0; i < dtoType.NumField(); i++ {
		field := dtoType.Field(i)

		// Skip private fields
		if !field.IsExported() {
			continue
		}

		// Skip struct fields (they will be handled via joins/preloads)
		if isBasicType(field.Type) {
			columnName := getColumnNameFromDTO(field)
			fields = append(fields, columnName)
		}
	}

	return fields
}

// mapToDTO maps data from an entity to a DTO using reflection
func mapToDTO(source interface{}, dest interface{}) error {
	sourceValue := reflect.ValueOf(source)
	destValue := reflect.ValueOf(dest)

	// dest must be a pointer
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()

	// If source is a pointer, dereference it
	if sourceValue.Kind() == reflect.Ptr {
		sourceValue = sourceValue.Elem()
	}

	sourceType := sourceValue.Type()
	destType := destValue.Type()

	// Map fields by name (case-insensitive)
	for i := 0; i < destType.NumField(); i++ {
		destField := destType.Field(i)
		destFieldValue := destValue.Field(i)

		if !destFieldValue.CanSet() {
			continue
		}

		// Find corresponding field in source
		sourceFieldValue := findFieldByName(sourceValue, sourceType, destField.Name)
		if sourceFieldValue.IsValid() && sourceFieldValue.Type().ConvertibleTo(destFieldValue.Type()) {
			destFieldValue.Set(sourceFieldValue.Convert(destFieldValue.Type()))
		}
	}

	return nil
}

// findFieldByName finds a field by name (case-insensitive)
func findFieldByName(value reflect.Value, valueType reflect.Type, fieldName string) reflect.Value {
	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		if strings.EqualFold(field.Name, fieldName) {
			return value.Field(i)
		}
	}
	return reflect.Value{}
}

// createPartialEntity creates a new instance of the entity with only projection fields filled
func createPartialEntity[T any](fullEntity *T, dtoInterface interface{}) (*T, error) {
	if fullEntity == nil {
		return nil, fmt.Errorf("entity cannot be nil")
	}

	// Create a new empty instance of the entity
	var partialEntity T

	// Get projection fields
	projectionFields := createProjectionFromDTO(dtoInterface)

	// Entity values
	fullValue := reflect.ValueOf(fullEntity).Elem()
	partialValue := reflect.ValueOf(&partialEntity).Elem()

	entityType := reflect.TypeOf(partialEntity)

	// Map only fields that are in the projection
	for _, projField := range projectionFields {
		// Find corresponding field in entity
		for i := 0; i < entityType.NumField(); i++ {
			entityField := entityType.Field(i)

			// Check if field matches (by column name or field name)
			columnName := getColumnName(entityField)
			if columnName == projField || strings.EqualFold(entityField.Name, projField) {

				fullFieldValue := fullValue.Field(i)
				partialFieldValue := partialValue.Field(i)

				if partialFieldValue.CanSet() && fullFieldValue.IsValid() {
					partialFieldValue.Set(fullFieldValue)
				}
				break
			}
		}
	}

	return &partialEntity, nil
}

// mapEntityToDTO maps a complete entity to a DTO with support for nested structs
func mapEntityToDTO[T any](entity *T, dtoInterface interface{}) (interface{}, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity cannot be nil")
	}

	// Create a new instance of the DTO
	dtoType := reflect.TypeOf(dtoInterface)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	dtoValue := reflect.New(dtoType).Elem()
	entityValue := reflect.ValueOf(entity).Elem()
	entityType := reflect.TypeOf(*entity)

	// Map fields from entity to DTO
	for i := 0; i < dtoType.NumField(); i++ {
		dtoField := dtoType.Field(i)
		dtoFieldValue := dtoValue.Field(i)

		if !dtoFieldValue.CanSet() {
			continue
		}

		// Find corresponding field in entity
		var entityFieldValue reflect.Value

		// First try by exact name
		if entityValue.FieldByName(dtoField.Name).IsValid() {
			entityFieldValue = entityValue.FieldByName(dtoField.Name)
		} else {
			// Try by column name
			columnName := getColumnName(dtoField)
			for j := 0; j < entityType.NumField(); j++ {
				entityField := entityType.Field(j)
				if getColumnName(entityField) == columnName {
					entityFieldValue = entityValue.Field(j)
					break
				}
			}
		}

		// If field found, do the mapping
		if entityFieldValue.IsValid() {
			if err := mapFieldValue(entityFieldValue, dtoFieldValue, dtoField); err != nil {
				return nil, fmt.Errorf("error mapping field %s: %w", dtoField.Name, err)
			}
		}
	}

	return dtoValue.Addr().Interface(), nil
}

// mapFieldValue maps a single field value from entity to DTO, handling nested structs
func mapFieldValue(entityFieldValue, dtoFieldValue reflect.Value, dtoField reflect.StructField) error {
	// Handle basic types and direct conversions
	if entityFieldValue.Type().ConvertibleTo(dtoFieldValue.Type()) {
		dtoFieldValue.Set(entityFieldValue.Convert(dtoFieldValue.Type()))
		return nil
	}

	// Handle nested structs (from preloads)
	if entityFieldValue.Kind() == reflect.Struct && dtoFieldValue.Kind() == reflect.Struct {
		return mapStructToStruct(entityFieldValue, dtoFieldValue)
	}

	// Handle pointer to struct (common in GORM associations)
	if entityFieldValue.Kind() == reflect.Ptr && !entityFieldValue.IsNil() &&
		dtoFieldValue.Kind() == reflect.Struct && entityFieldValue.Elem().Kind() == reflect.Struct {
		return mapStructToStruct(entityFieldValue.Elem(), dtoFieldValue)
	}

	// Handle slice of structs (has many relationships)
	if entityFieldValue.Kind() == reflect.Slice && dtoFieldValue.Kind() == reflect.Slice {
		return mapSliceToSlice(entityFieldValue, dtoFieldValue)
	}

	// If types are not compatible and no special handling, skip the field
	return nil
}

// mapStructToStruct maps fields from source struct to destination struct
func mapStructToStruct(sourceValue, destValue reflect.Value) error {
	sourceType := sourceValue.Type()
	destType := destValue.Type()

	// Map each field from destination struct
	for i := 0; i < destType.NumField(); i++ {
		destField := destType.Field(i)
		destFieldValue := destValue.Field(i)

		if !destFieldValue.CanSet() {
			continue
		}

		// Find corresponding field in source
		var sourceFieldValue reflect.Value

		// Try by exact name first
		if sourceValue.FieldByName(destField.Name).IsValid() {
			sourceFieldValue = sourceValue.FieldByName(destField.Name)
		} else {
			// Try by column name mapping
			destColumnName := getColumnName(destField)
			for j := 0; j < sourceType.NumField(); j++ {
				sourceField := sourceType.Field(j)
				if getColumnName(sourceField) == destColumnName {
					sourceFieldValue = sourceValue.Field(j)
					break
				}
			}
		}

		// Map the field if found
		if sourceFieldValue.IsValid() {
			if err := mapFieldValue(sourceFieldValue, destFieldValue, destField); err != nil {
				return fmt.Errorf("error mapping nested field %s: %w", destField.Name, err)
			}
		}
	}

	return nil
}

// mapSliceToSlice maps slice elements from source to destination
func mapSliceToSlice(sourceValue, destValue reflect.Value) error {
	if sourceValue.Len() == 0 {
		return nil
	}

	destElemType := destValue.Type().Elem()
	sourceElemType := sourceValue.Type().Elem()

	// Create new slice for destination
	newSlice := reflect.MakeSlice(destValue.Type(), sourceValue.Len(), sourceValue.Len())

	for i := 0; i < sourceValue.Len(); i++ {
		sourceElem := sourceValue.Index(i)
		destElem := newSlice.Index(i)

		// Handle struct to struct mapping in slices
		if sourceElemType.Kind() == reflect.Struct && destElemType.Kind() == reflect.Struct {
			if err := mapStructToStruct(sourceElem, destElem); err != nil {
				return fmt.Errorf("error mapping slice element %d: %w", i, err)
			}
		} else if sourceElem.Type().ConvertibleTo(destElem.Type()) {
			destElem.Set(sourceElem.Convert(destElem.Type()))
		}
	}

	destValue.Set(newSlice)
	return nil
}
