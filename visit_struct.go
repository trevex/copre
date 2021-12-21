package config

import (
	"reflect"

	"github.com/imdario/mergo"
)

type fieldMapper func(path []string, field reflect.StructField) (interface{}, error)

// visitStruct will populate fields with the values returned from mapper.
// dst is expected to be pointer!
func visitStruct(dst interface{}, mapper fieldMapper) error {
	v := reflect.ValueOf(dst).Elem()
	src := map[string]interface{}{} // will be mapFields/visit dst, but merge src
	if err := visitFields([]string{}, v, src, mapper); err != nil {
		return err
	}
	if err := mergo.Map(dst, src, mergo.WithOverride); err != nil {
		return err
	}
	return nil
}

func visitFields(path []string, value reflect.Value, dst map[string]interface{}, mapper fieldMapper) error {
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		// If pointer, dereference
		if field.Kind() == reflect.Ptr {
			// TODO: Do we have to check IsNil and set zero value although
			//       we are just traversing? => Test this!
			field = field.Elem()
		}
		structField := valueType.Field(i)
		fieldName := structField.Name
		fieldPath := append(path, fieldName)
		// If struct, call mapFields again
		if field.Kind() == reflect.Struct {
			fieldDst := map[string]interface{}{}
			err := visitFields(fieldPath, field, fieldDst, mapper)
			if err != nil {
				return err
			}
			dst[fieldName] = fieldDst
			continue
		}
		result, err := mapper(fieldPath, structField)
		if err != nil {
			return err
		}
		// If result is not nil, set it
		// (playing it safe here using reflection)
		if result == nil || (reflect.ValueOf(result).Kind() == reflect.Ptr && reflect.ValueOf(result).IsNil()) {
			continue
		}
		dst[fieldName] = result
	}
	return nil
}
