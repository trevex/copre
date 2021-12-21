package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

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
		prefix: "",
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return b.Loader(LoaderFunc(func(dst interface{}) error {
		return visitStruct(dst, func(path []string, field reflect.StructField) (interface{}, error) {
			tag, ok := field.Tag.Lookup("env")
			if !ok {
				return nil, nil
			}
			params := strings.Split(tag, ",")
			key := params[0]
			if key == "" {
				return nil, nil
			}
			if o.prefix != "" {
				key = fmt.Sprintf("%s_%s", o.prefix, key)
			}
			// TODO: Support base64 and hex!
			if val, ok := os.LookupEnv(key); ok {
				return convertString(field.Type, val)
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
		flagMap := getFlagMap(flags, &o)
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
				// TODO: Handle type conversion from underlying type to field.Type()?
				//       Or at least inform user if they are mismatched?
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
	// Loop twice over loaders, one preprocess pass, then the real one
	// for _, l := range b.loaders {
	// 	if err := l.Preprocess(b.dst); err != nil {
	// 		return err
	// 	}
	// }
	for _, l := range b.loaders {
		if err := l.Process(b.dst); err != nil {
			return err
		}
	}
	return nil
}

func getFlagMap(flags *pflag.FlagSet, o *flagSetOptions) map[string]interface{} {
	flagMap := map[string]interface{}{}
	flags.VisitAll(func(flag *pflag.Flag) {
		// If neither includeUnchanged is set nor useFlagDefaults, we have
		// to return and ignore the unchanged flag
		if !flag.Changed && !(o.includeUnchanged || o.useFlagDefaults) {
			return
		}
		// TODO: What about the pflag advanced types, e.g. IP
		var v interface{}
		switch flag.Value.Type() {
		case "bool":
			v, _ = flags.GetBool(flag.Name)
		case "int":
			v, _ = flags.GetInt(flag.Name)
		case "int8":
			v, _ = flags.GetInt8(flag.Name)
		case "int16":
			v, _ = flags.GetInt16(flag.Name)
		case "int32":
			v, _ = flags.GetInt32(flag.Name)
		case "int64":
			v, _ = flags.GetInt64(flag.Name)
		case "uint":
			v, _ = flags.GetUint(flag.Name)
		case "uint8":
			v, _ = flags.GetUint8(flag.Name)
		case "uint16":
			v, _ = flags.GetUint16(flag.Name)
		case "uint32":
			v, _ = flags.GetUint32(flag.Name)
		case "uint64":
			v, _ = flags.GetUint64(flag.Name)
		case "float32":
			v, _ = flags.GetFloat32(flag.Name)
		case "float":
			v, _ = flags.GetFloat64(flag.Name)
		case "stringSlice":
			v, _ = flags.GetStringSlice(flag.Name)
		case "intSlice":
			v, _ = flags.GetIntSlice(flag.Name)
		default:
			v = flag.Value.String()
		}
		t := reflect.TypeOf(v)
		// If the flag has the corresponding zero-type set, do not set it
		if t.Comparable() && v == reflect.Zero(t).Interface() {
			return
		}
		// If the flag is an empty slice, do not set it
		if t.Kind() == reflect.Slice && reflect.ValueOf(v).Len() == 0 {
			return
		}
		flagMap[flag.Name] = v
	})
	return flagMap
}
