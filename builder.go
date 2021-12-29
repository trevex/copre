package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/imdario/mergo"
	"github.com/spf13/pflag"
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
	o := fileOptions{
		ignoreNotFound: false,
		expandEnv:      false,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return b.Loader(LoaderFunc(func(dst interface{}) error {
		d, err := ioutil.ReadFile(filepath)
		if o.ignoreNotFound && os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if o.expandEnv {
			d = []byte(os.ExpandEnv(string(d)))
		}
		if err := unmarshal(d, dst); err != nil {
			return err
		}
		return nil
	}))
}

func (b *Builder) Env(opts ...EnvOption) *Builder {
	o := envOptions{
		prefix:    "",
		keyGetter: func(s []string) string { return "" },
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return b.Loader(LoaderFunc(func(dst interface{}) error {
		return visitStruct(dst, func(path []string, field reflect.StructField) (interface{}, error) {
			noPrefix := false
			key := o.keyGetter(path)
			targetType := field.Type
			if tag, ok := field.Tag.Lookup("env"); ok {
				params := strings.Split(tag, ",")
				// Only set key if provided
				if params[0] != "" {
					key = params[0]
				}

				if len(params) > 1 { // If options are set, let's handle them
					for _, param := range params[1:] {
						// Check options for byte arrays
						if param == "hex" || param == "base64" {
							if targetType.Kind() != reflect.Slice || targetType.Elem().Kind() != reflect.Uint8 {
								return nil, fmt.Errorf("unsupported option '%s' for type '%s'", params[1], targetType.String())
							}
							if param == "hex" {
								targetType = reflect.TypeOf(convertBytesHexMarker{})
							} else if param == "base64" {
								targetType = reflect.TypeOf(convertBytesBase64Marker{})
							}
						}
						if param == "noprefix" {
							noPrefix = true
						}
					}
				}
			}

			if o.prefix != "" && !noPrefix {
				key = fmt.Sprintf("%s_%s", o.prefix, key)
			}

			if val, ok := os.LookupEnv(key); ok {
				return convertString(targetType, val)
			}
			return nil, nil
		})
	}))
}

func (b *Builder) FlagSet(flags *pflag.FlagSet, opts ...FlagSetOption) *Builder {
	o := flagSetOptions{
		useFlagDefaults:  false,
		includeUnchanged: false,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	loaderFunc := LoaderFunc(func(dst interface{}) error {
		flagMap := listFlags(flags, o.includeUnchanged, o.useFlagDefaults)
		// Ok, so if useFlagDefaults is set this is the first run (as loader
		// will be run twice), for the second run, we need to ignore them
		if o.useFlagDefaults {
			o.useFlagDefaults = false
		}
		return visitStruct(dst, func(path []string, field reflect.StructField) (interface{}, error) {
			tag, ok := field.Tag.Lookup("flag")
			if !ok {
				return nil, nil
			}
			params := strings.Split(tag, ",")
			key := params[0]
			if key == "" {
				return nil, nil
			}
			if val, ok := flagMap[key]; ok {
				// Mismatch is handled by visitStruct
				return val, nil
			}
			return nil, nil
		})

	})
	// If use default is set we will run the loaderFunc twice:
	// Once before everything else to set defaults, second time in the correct order.
	if o.useFlagDefaults {
		b.loaders = append([]Loader{loaderFunc}, b.loaders...)
	}
	return b.Loader(loaderFunc)
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
	for _, l := range b.loaders {
		dst := reflect.New(v.Type()).Interface()
		if err := l.Process(dst); err != nil {
			return err
		}
		if err := mergo.Merge(b.dst, dst, mergo.WithOverride); err != nil {
			return err
		}
	}
	return nil
}
