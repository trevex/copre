package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/pflag"
)

// Builder

type Builder struct {
	dst interface{}
	// All loaders that will be processed in order to build the configuration
	loaders []Loader
}

func NewBuilder(dst interface{}) *Builder {
	return &Builder{dst, []Loader{}}
}

func (b *Builder) Loader(loader Loader) *Builder {
	b.loaders = append(b.loaders, loader)
	return b
}

func (b *Builder) File(filepath string, unmarshal UnmarshalFunc, opts ...FileOption) *Builder {
	return b.Loader(NewFileLoader(filepath, unmarshal, opts...))
}

func (b *Builder) Env(prefix string, opts ...EnvOption) *Builder {
	return b.Loader(NewEnvLoader(prefix, opts...))
}

func (b *Builder) FlagSet(flags *pflag.FlagSet, opts ...FlagSetOption) *Builder {
	return b.Loader(NewFlagSetLoader(flags, opts...))
}

func (b *Builder) Build() error {
	// Make sure dst is a pointer to struct
	v := reflect.ValueOf(b.dst)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected destination to be pointer not %s", v.Kind())
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("Expected destination to point to struct not %s", v.Kind())
	}
	// Loop twice over loaders, one preprocess pass, then the real one
	for _, l := range b.loaders {
		if err := l.Preprocess(b.dst); err != nil {
			return err
		}
	}
	for _, l := range b.loaders {
		if err := l.Process(b.dst); err != nil {
			return err
		}
	}
	return nil
}
