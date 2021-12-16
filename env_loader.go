package config

import (
	"os"
	"reflect"
	"strings"
)

// EnvKeyGetter-functions translate prefix and path into an environment variable name.
type EnvKeyGetter func(prefix string, path []string) string

var (
	DefaultEnvKeyGetter EnvKeyGetter = func(prefix string, path []string) string {
		return strings.ToUpper(prefix + "_" + strings.Join(path, "_"))
	}
	DefaultEnvTagHandler = func(fallback string, field reflect.StructField) string {
		if tag, ok := field.Tag.Lookup("env"); ok {
			return strings.Split(tag, ",")[0]
		}
		return fallback
	}
)

type envLoader struct {
	prefix       string
	envKeyGetter EnvKeyGetter
	tagHandler   TagHandler
}

type EnvOption interface {
	apply(*envLoader)
}

type envOptionAdapter func(*envLoader)

func (c envOptionAdapter) apply(l *envLoader) {
	c(l)
}

func WithEnvKeyGetter(envKeyGetter EnvKeyGetter) EnvOption {
	return envOptionAdapter(func(l *envLoader) {
		l.envKeyGetter = envKeyGetter
	})
}

func WithEnvTagHandler(tagHandler TagHandler) EnvOption {
	return envOptionAdapter(func(l *envLoader) {
		l.tagHandler = tagHandler
	})
}

func NewEnvLoader(prefix string, opts ...EnvOption) Loader {
	l := envLoader{
		prefix:       prefix,
		envKeyGetter: DefaultEnvKeyGetter,
		tagHandler:   DefaultEnvTagHandler,
	}
	for _, opt := range opts {
		opt.apply(&l)
	}
	return &l
}

func (l *envLoader) Preprocess(dst interface{}) error {
	return nil
}

func (l *envLoader) Process(dst interface{}) error {
	return populateStruct(dst, func(path []string, field reflect.StructField) (interface{}, error) {
		key := l.tagHandler(l.envKeyGetter(l.prefix, path), field)
		if val, ok := os.LookupEnv(key); ok {
			return convertString(field.Type, val, &defaultConvertStringConfig) // TODO: make config configurable
		}
		return nil, nil
	})
}
