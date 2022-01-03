package copre

import (
	"fmt"
	"reflect"

	"github.com/imdario/mergo"
)

// Loader is the interface that needs to be implemented to be able to load
// configuration from a configuration source. See Env, File or FlagSet for
// the implementations provided by this library.
// They only have to implement a single method Process, which populates the
// passed-in configuration-struct dst and returns an error if problems occur.
type Loader interface {
	Process(dst interface{}) error
}

// LoaderFunc implements the Loader interface for a individual functions.
type LoaderFunc func(dst interface{}) error

// Process calls the LoaderFunc underneath.
func (fn LoaderFunc) Process(dst interface{}) error {
	return fn(dst)
}

// Load is the central function tying all the building blocks together.
// It allows the composition of the loader precedence.
// For a given pointer to struct dst, an empty instantiation of the same type
// is create for each loader. Each loader populates its copy and all copies are
// merged into dst in the specified order.
//
// See the README for several examples.
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
