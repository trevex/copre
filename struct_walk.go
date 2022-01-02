package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/reflectwalk"
)

// FieldMapper is a function that takes the path of a field in a nested structure
// and field itself, to return a value or an error.
type FieldMapper func(path []string, field reflect.StructField) (interface{}, error)

// StructWalk walks/visits every field of a struct (including nested) and calls
// fieldMapper for every field. If an internal error is encountered or an error
// is returned by fieldMapper it is immediately returned and traversal stopped.
func StructWalk(dst interface{}, fieldMapper FieldMapper) error {
	w := &structWalker{
		Path:        []string{},
		FieldMapper: fieldMapper,
	}
	return reflectwalk.Walk(dst, w)
}

type structWalker struct {
	Path        []string
	FieldMapper FieldMapper
}

func (w *structWalker) Enter(l reflectwalk.Location) error {
	return nil
}

func (w *structWalker) Exit(l reflectwalk.Location) error {
	if l == reflectwalk.Struct && len(w.Path) > 0 {
		w.Path = w.Path[:len(w.Path)-1]
	}
	return nil
}

func (w *structWalker) Struct(v reflect.Value) error {
	return nil
}

func (w *structWalker) StructField(sf reflect.StructField, v reflect.Value) (err error) {
	if sf.Type.Kind() == reflect.Ptr {
		// For pointers that are nil we try to set the default value
		if v.IsNil() {
			if !v.CanSet() {
				return
			}
			ptr := reflect.New(sf.Type.Elem())
			v.Set(ptr)
		}
		// We do not mutate pointers, so let's sets retrieve element
		v = v.Elem()
	}

	// If type is struct or pointer to struct, append path
	if sf.Type.Kind() == reflect.Struct || (sf.Type.Kind() == reflect.Ptr && sf.Type.Elem().Kind() == reflect.Struct) {
		w.Path = append(w.Path, sf.Name)
		return
	}

	var result interface{}
	path := append(w.Path, sf.Name)
	result, err = w.FieldMapper(path, sf)
	if err != nil {
		return
	}

	// If result is not nil, set it
	// (playing it safe here using reflection)
	if result == nil || (reflect.ValueOf(result).Kind() == reflect.Ptr && reflect.ValueOf(result).IsNil()) {
		return nil
	}

	defer func() {
		if recover() != nil {
			err = fmt.Errorf("failed to set value at path '.%s': expected type '%s', got '%T'.", strings.Join(path, "."), v.Type().String(), result)
		}
	}()
	v.Set(reflect.ValueOf(result))
	return
}
