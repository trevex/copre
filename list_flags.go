package config

import (
	"reflect"

	"github.com/spf13/pflag"
)

func listFlags(flags *pflag.FlagSet, includeUnchanged, useFlagDefaults bool) map[string]interface{} {
	flagMap := map[string]interface{}{}
	flags.VisitAll(func(flag *pflag.Flag) {
		// If neither includeUnchanged is set nor useFlagDefaults, we have
		// to return and ignore the unchanged flag
		if !flag.Changed && !(includeUnchanged || useFlagDefaults) {
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
