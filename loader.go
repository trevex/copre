package config

import (
	"reflect"
)

// Loader are chained by Builder to populate a given config.
type Loader interface {
	Preprocess(dst interface{}) error
	Process(dst interface{}) error
}

// A function that will will try to extract a value from struct tags.
// If it is not able to do so, fallback will be returned instead.
type TagHandler func(fallback string, field reflect.StructField) string
