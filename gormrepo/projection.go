package gormrepo

import (
	"fmt"
	"reflect"
	"strings"
)

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

		if fieldType.Kind() == reflect.Struct && !isBasicType(fieldType) {
			return true
		}
	}

	return false
}

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

		if fieldType.Kind() == reflect.Struct && !isBasicType(fieldType) {
			preloadName := field.Tag.Get("preload")
			if preloadName == "" {
				preloadName = field.Name
			}
			preloads = append(preloads, preloadName)
		}
	}

	return preloads
}

func isBasicType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	}
	return false
}

func getColumnName(field reflect.StructField) string {
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
		}
	}

	return toSnakeCase(field.Name)
}

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

func createProjectionFromDTO(dtoInterface interface{}) []string {
	dtoType := reflect.TypeOf(dtoInterface)

	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	var fields []string

	for i := 0; i < dtoType.NumField(); i++ {
		field := dtoType.Field(i)

		if !field.IsExported() {
			continue
		}

		if isBasicType(field.Type) {
			columnName := getColumnNameFromDTO(field)
			fields = append(fields, columnName)
		}
	}

	return fields
}

func getColumnNameFromDTO(field reflect.StructField) string {
	if projection := field.Tag.Get("projection"); projection != "" {
		return projection
	}

	if gormTag := field.Tag.Get("gorm"); gormTag != "" {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
		}
	}

	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		return strings.Split(jsonTag, ",")[0]
	}

	return toSnakeCase(field.Name)
}

func mapEntityToDTO[T any](entity *T, dtoInterface interface{}) (interface{}, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity cannot be nil")
	}

	dtoType := reflect.TypeOf(dtoInterface)
	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	dtoValue := reflect.New(dtoType).Elem()
	entityValue := reflect.ValueOf(entity).Elem()
	entityType := reflect.TypeOf(*entity)

	for i := 0; i < dtoType.NumField(); i++ {
		dtoField := dtoType.Field(i)
		dtoFieldValue := dtoValue.Field(i)

		if !dtoFieldValue.CanSet() {
			continue
		}

		var entityFieldValue reflect.Value

		if entityValue.FieldByName(dtoField.Name).IsValid() {
			entityFieldValue = entityValue.FieldByName(dtoField.Name)
		} else {
			columnName := getColumnName(dtoField)
			for j := 0; j < entityType.NumField(); j++ {
				entityField := entityType.Field(j)
				if getColumnName(entityField) == columnName {
					entityFieldValue = entityValue.Field(j)
					break
				}
			}
		}

		if entityFieldValue.IsValid() {
			if err := mapFieldValue(entityFieldValue, dtoFieldValue, dtoField); err != nil {
				return nil, fmt.Errorf("error mapping field %s: %w", dtoField.Name, err)
			}
		}
	}

	return dtoValue.Addr().Interface(), nil
}

func mapFieldValue(entityFieldValue, dtoFieldValue reflect.Value, dtoField reflect.StructField) error {
	if entityFieldValue.Type().ConvertibleTo(dtoFieldValue.Type()) {
		dtoFieldValue.Set(entityFieldValue.Convert(dtoFieldValue.Type()))
		return nil
	}

	if entityFieldValue.Kind() == reflect.Struct && dtoFieldValue.Kind() == reflect.Struct {
		return mapStructToStruct(entityFieldValue, dtoFieldValue)
	}

	if entityFieldValue.Kind() == reflect.Ptr && !entityFieldValue.IsNil() &&
		dtoFieldValue.Kind() == reflect.Struct && entityFieldValue.Elem().Kind() == reflect.Struct {
		return mapStructToStruct(entityFieldValue.Elem(), dtoFieldValue)
	}

	if entityFieldValue.Kind() == reflect.Slice && dtoFieldValue.Kind() == reflect.Slice {
		return mapSliceToSlice(entityFieldValue, dtoFieldValue)
	}

	return nil
}

func mapStructToStruct(sourceValue, destValue reflect.Value) error {
	sourceType := sourceValue.Type()
	destType := destValue.Type()

	for i := 0; i < destType.NumField(); i++ {
		destField := destType.Field(i)
		destFieldValue := destValue.Field(i)

		if !destFieldValue.CanSet() {
			continue
		}

		var sourceFieldValue reflect.Value

		if sourceValue.FieldByName(destField.Name).IsValid() {
			sourceFieldValue = sourceValue.FieldByName(destField.Name)
		} else {
			destColumnName := getColumnName(destField)
			for j := 0; j < sourceType.NumField(); j++ {
				sourceField := sourceType.Field(j)
				if getColumnName(sourceField) == destColumnName {
					sourceFieldValue = sourceValue.Field(j)
					break
				}
			}
		}

		if sourceFieldValue.IsValid() {
			if err := mapFieldValue(sourceFieldValue, destFieldValue, destField); err != nil {
				return fmt.Errorf("error mapping nested field %s: %w", destField.Name, err)
			}
		}
	}

	return nil
}

func mapSliceToSlice(sourceValue, destValue reflect.Value) error {
	if sourceValue.Len() == 0 {
		return nil
	}

	destElemType := destValue.Type().Elem()
	sourceElemType := sourceValue.Type().Elem()

	newSlice := reflect.MakeSlice(destValue.Type(), sourceValue.Len(), sourceValue.Len())

	for i := 0; i < sourceValue.Len(); i++ {
		sourceElem := sourceValue.Index(i)
		destElem := newSlice.Index(i)

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
