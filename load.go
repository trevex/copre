package config

import (
	"fmt"
	"reflect"

	"github.com/imdario/mergo"
)

type UnmarshalFunc func(data []byte, dst interface{}) error

// Loaders are chained by Builder to populate the given struct.
type Loader interface {
	Process(dst interface{}) error
}

type LoaderFunc func(dst interface{}) error

func (fn LoaderFunc) Process(dst interface{}) error {
	return fn(dst)
}

func Load(dst interface{}, loaders ...Loader) error {
	// Make sure dst is a pointer to struct
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected destination to be pointer not %s", v.Kind())
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("Expected destination to point to struct not %s", v.Kind())
	}
	for _, l := range loaders {
		tmp := reflect.New(v.Type()).Interface()
		if err := l.Process(tmp); err != nil {
			return err
		}
		if err := mergo.Merge(dst, tmp, mergo.WithOverride); err != nil {
			return err
		}
	}
	return nil
}
