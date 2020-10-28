package functions

import (
	"reflect"
)

func IsAvailable(data interface{}, fieldName string) bool {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	return v.FieldByName(fieldName).IsValid()
}
