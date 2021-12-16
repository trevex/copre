package config

import (
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/pflag"
)

type FlagNameGetter func(path []string) string

var (
	DefaultFlagNameGetter FlagNameGetter = func(path []string) string {
		return strings.ToLower(strings.Join(path, "-"))
	}
	KebapFlagNameGetter FlagNameGetter = func(path []string) string {
		return strcase.ToKebab(strings.Join(path, "-"))
	}
	DefaultFlagTagHandler = func(key string, field reflect.StructField) string {
		if tag, ok := field.Tag.Lookup("flag"); ok {
			return strings.Split(tag, ",")[0]
		}
		return key
	}
)

type flagSetLoader struct {
	flags        *pflag.FlagSet
	visitedFlags []string
	// Options
	includeUnchanged bool
	useFlagDefaults  bool
	flagNameGetter   FlagNameGetter
	tagHandler       TagHandler
}

type FlagSetOption interface {
	apply(*flagSetLoader)
}

type flagSetOptionAdapter func(*flagSetLoader)

func (c flagSetOptionAdapter) apply(l *flagSetLoader) {
	c(l)
}

// IncludeUnchanged will also process the values of unchanged flags. Effectively
// this means the flag defaults, if non zero, will be set as well.
func IncludeUnchanged(f ...bool) FlagSetOption {
	return flagSetOptionAdapter(func(l *flagSetLoader) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		l.includeUnchanged = v
	})
}

// UseFlagDefaults will run and check the FlagSet once before any other method
// is called to set the FlagSet defaults to the destination config. The changed
// values of the FlagSet still respect the defined precendence.
func UseFlagDefaults(f ...bool) FlagSetOption {
	return flagSetOptionAdapter(func(l *flagSetLoader) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		l.useFlagDefaults = v
	})
}

func WithFlagNameGetter(flagNameGetter FlagNameGetter) FlagSetOption {
	return flagSetOptionAdapter(func(l *flagSetLoader) {
		l.flagNameGetter = flagNameGetter
	})
}

func NewFlagSetLoader(flags *pflag.FlagSet, opts ...FlagSetOption) Loader {
	l := flagSetLoader{
		flags:          flags,
		visitedFlags:   []string{},
		flagNameGetter: DefaultFlagNameGetter,
		tagHandler:     DefaultFlagTagHandler,
	}
	for _, opt := range opts {
		opt.apply(&l)
	}
	return &l
}

func (l *flagSetLoader) Preprocess(dst interface{}) error {
	// If useFlagDefaults is specified the operation runs twice, once including
	// unchanged flags before anything else to make sure they are set first
	// and then the normal pass.
	if l.useFlagDefaults {
		return l.Process(dst)
	}
	return nil
}

func (l *flagSetLoader) Process(dst interface{}) error {

	// First get all flag values
	flagMap := map[string]interface{}{}
	l.flags.VisitAll(func(flag *pflag.Flag) {
		l.visitedFlags = append(l.visitedFlags, flag.Name)
		// If neither includeUnchanged is set nor useFlagDefaults, we have
		// to return and ignore the unchanged flag
		if !flag.Changed && !(l.includeUnchanged || l.useFlagDefaults) {
			return
		}
		// TODO: What about the pflag advanced types, e.g. IP
		var v interface{}
		switch flag.Value.Type() {
		case "bool":
			v, _ = l.flags.GetBool(flag.Name)
		case "int":
			v, _ = l.flags.GetInt(flag.Name)
		case "int8":
			v, _ = l.flags.GetInt8(flag.Name)
		case "int16":
			v, _ = l.flags.GetInt16(flag.Name)
		case "int32":
			v, _ = l.flags.GetInt32(flag.Name)
		case "int64":
			v, _ = l.flags.GetInt64(flag.Name)
		case "uint":
			v, _ = l.flags.GetUint(flag.Name)
		case "uint8":
			v, _ = l.flags.GetUint8(flag.Name)
		case "uint16":
			v, _ = l.flags.GetUint16(flag.Name)
		case "uint32":
			v, _ = l.flags.GetUint32(flag.Name)
		case "uint64":
			v, _ = l.flags.GetUint64(flag.Name)
		case "float32":
			v, _ = l.flags.GetFloat32(flag.Name)
		case "float":
			v, _ = l.flags.GetFloat64(flag.Name)
		case "stringSlice":
			v, _ = l.flags.GetStringSlice(flag.Name)
		case "intSlice":
			v, _ = l.flags.GetIntSlice(flag.Name)
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
	// Ok so if useFlagDefaults is set this is the first run and we included
	// unchanged flags, for the second run, we need to ignore them
	if l.useFlagDefaults {
		l.useFlagDefaults = false
	}
	return populateStruct(dst, func(path []string, field reflect.StructField) (interface{}, error) {
		key := l.tagHandler(l.flagNameGetter(path), field)
		if val, ok := flagMap[key]; ok {
			// TODO: Handle type conversion from underlying type to field.Type()?
			//       Or at least inform user if they are mismatched?
			return val, nil
		}
		return nil, nil
	})
}
