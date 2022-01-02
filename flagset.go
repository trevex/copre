package config

import (
	"reflect"
	"strings"

	"github.com/spf13/pflag"
)

type flagSetOptions struct {
	tag              string
	includeUnchanged bool
	nameGetter       func([]string) string
}

// FlagSetOption configures how a FlagSet is used to populate a given structure.
type FlagSetOption interface {
	apply(*flagSetOptions)
}

type flagSetOptionAdapter func(*flagSetOptions)

func (c flagSetOptionAdapter) apply(o *flagSetOptions) {
	c(o)
}

// IncludeUnchanged will also process the values of unchanged flags. Effectively
// this means the flag defaults, if non zero, will be set as well.
func IncludeUnchanged(f ...bool) FlagSetOption {
	return flagSetOptionAdapter(func(o *flagSetOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		o.includeUnchanged = v
	})
}

// ComputeFlagName will remove the requirement to explicitly specify the
// flag-name with a tag. For all fields not explicitly tagged, the name
// will be computed based on the path by the provided nameGetter function.
// For example:
//  ComputeFlagName(KebabCase)
func ComputeFlagName(nameGetter func([]string) string) FlagSetOption {
	return flagSetOptionAdapter(func(o *flagSetOptions) {
		o.nameGetter = nameGetter
	})
}

func OverrideFlagTag(tag string) FlagSetOption {
	return flagSetOptionAdapter(func(o *flagSetOptions) {
		o.tag = tag
	})
}

func FlagSet(flags *pflag.FlagSet, opts ...FlagSetOption) Loader {
	o := flagSetOptions{
		tag:              "flag",
		includeUnchanged: false,
		nameGetter:       func(s []string) string { return "" },
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return LoaderFunc(func(dst interface{}) error {
		flagMap := listFlags(flags, o.includeUnchanged)
		return StructWalk(dst, func(path []string, field reflect.StructField) (interface{}, error) {
			name := o.nameGetter(path)
			if tag, ok := field.Tag.Lookup(o.tag); ok {
				params := strings.Split(tag, ",")
				name = params[0]
			}
			if name == "" {
				return nil, nil
			}
			if val, ok := flagMap[name]; ok {
				// Mismatch is handled by StructWalk
				return val, nil
			}
			return nil, nil
		})

	})
}

func listFlags(flags *pflag.FlagSet, includeUnchanged bool) map[string]interface{} {
	flagMap := map[string]interface{}{}
	flags.VisitAll(func(flag *pflag.Flag) {
		// If includeUnchanged is not set, we have to return and ignore the unchanged flag
		if !flag.Changed && !includeUnchanged {
			return
		}
		var v interface{}
		switch flag.Value.Type() {
		case "bool":
			v, _ = flags.GetBool(flag.Name)
		case "int":
			v, _ = flags.GetInt(flag.Name)
		case "intSlice":
			v, _ = flags.GetIntSlice(flag.Name)
		case "int8":
			v, _ = flags.GetInt8(flag.Name)
		case "int16":
			v, _ = flags.GetInt16(flag.Name)
		case "int32":
			v, _ = flags.GetInt32(flag.Name)
		case "int32Slice":
			v, _ = flags.GetInt32Slice(flag.Name)
		case "int64":
			v, _ = flags.GetInt64(flag.Name)
		case "int64Slice":
			v, _ = flags.GetInt64Slice(flag.Name)
		case "uint":
			v, _ = flags.GetUint(flag.Name)
		case "uintSlice":
			v, _ = flags.GetUintSlice(flag.Name)
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
		case "float32Slice":
			v, _ = flags.GetFloat32Slice(flag.Name)
		case "float64":
			v, _ = flags.GetFloat64(flag.Name)
		case "float64Slice":
			v, _ = flags.GetFloat64Slice(flag.Name)
		case "stringSlice":
			v, _ = flags.GetStringSlice(flag.Name)
		case "stringArray":
			v, _ = flags.GetStringArray(flag.Name)
		case "stringToInt":
			v, _ = flags.GetStringToInt(flag.Name)
		case "stringToInt64":
			v, _ = flags.GetStringToInt64(flag.Name)
		case "stringToString":
			v, _ = flags.GetStringToString(flag.Name)
		case "bytesBase64":
			v, _ = flags.GetBytesBase64(flag.Name)
		case "bytesHex":
			v, _ = flags.GetBytesHex(flag.Name)
		case "count":
			v, _ = flags.GetCount(flag.Name)
		case "duration":
			v, _ = flags.GetDuration(flag.Name)
		case "durationSlice":
			v, _ = flags.GetDurationSlice(flag.Name)
		case "ip":
			v, _ = flags.GetIP(flag.Name)
		case "ipMask":
			v, _ = flags.GetIPv4Mask(flag.Name)
		case "ipNet":
			v, _ = flags.GetIPNet(flag.Name)
		case "ipSlice":
			v, _ = flags.GetIPSlice(flag.Name)
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
