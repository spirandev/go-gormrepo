package pkhelper

import (
	"fmt"
	"reflect"
	"strings"
)

type primaryKeyProvider interface {
	GetID() any
}

func GetPrimaryKey(entity any) (string, any, error) {
	if pk, ok := entity.(primaryKeyProvider); ok {
		return "id", pk.GetID(), nil
	}

	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	pkName, pkValue := findPrimaryKey(val, val.Type())
	if pkName != "" {
		return pkName, pkValue, nil
	}

	return "", nil, fmt.Errorf("primary key not found in %T", entity)
}

func findPrimaryKey(val reflect.Value, typ reflect.Type) (string, any) {
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if strings.EqualFold(field.Name, "id") ||
			strings.Contains(field.Tag.Get("gorm"), "primaryKey") {
			return field.Name, fieldVal.Interface()
		}

		if field.Anonymous && fieldVal.Kind() == reflect.Struct {
			if pkName, pkValue := findPrimaryKey(fieldVal, field.Type); pkName != "" {
				return pkName, pkValue
			}
		}
	}

	return "", nil
}
