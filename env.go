package config

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

type envOptions struct {
	prefix    string
	keyGetter func([]string) string
}

type EnvOption interface {
	apply(*envOptions)
}

type envOptionAdapter func(*envOptions)

func (c envOptionAdapter) apply(l *envOptions) {
	c(l)
}

func WithPrefix(prefix string) EnvOption {
	return envOptionAdapter(func(o *envOptions) {
		o.prefix = prefix
	})
}

func ComputeEnvKey(keyGetter func([]string) string) EnvOption {
	return envOptionAdapter(func(o *envOptions) {
		o.keyGetter = keyGetter
	})
}

func Env(opts ...EnvOption) Loader {
	o := envOptions{
		prefix:    "",
		keyGetter: func(s []string) string { return "" },
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return LoaderFunc(func(dst interface{}) error {
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
	})
}

const (
	arrayDelimiter = ","
	mapDelimiter   = ","
	mapKVDelimiter = "="
)

type convertBytesHexMarker []byte

type convertBytesBase64Marker []byte

// Converts `input` string to type `t` or returns error if operation is not
// possible. Type `t` needs to be NOT a pointer kind!
func convertString(t reflect.Type, input string) (interface{}, error) {
	if t.Kind() == reflect.Ptr {
		return nil, fmt.Errorf("no pointer kinds allowed")
	}

	// Let's handle the marker structs
	if strings.Contains(t.PkgPath(), "copre") {
		switch t.Name() {
		case "convertBytesHexMarker":
			return hex.DecodeString(input)
		case "convertBytesBase64Marker":
			return base64.StdEncoding.DecodeString(input)
		}
	}

	// Handle supported net-types
	if t.PkgPath() == "net" {
		switch t.Name() {
		case "IP":
			ip := net.ParseIP(input)
			if ip == nil {
				return nil, fmt.Errorf("unable to parse '%s' as net.IP", input)
			}
			return ip, nil
		case "IPMask":
			ipMask := pflag.ParseIPv4Mask(input)
			if ipMask == nil {
				return nil, fmt.Errorf("unable to parse '%s' as net.IPMast", input)
			}
			return ipMask, nil
		case "IPNet":
			ip, ipNet, err := net.ParseCIDR(input)
			if err != nil {
				return nil, fmt.Errorf("unable to parse '%s' as net.IPNet, failed with: %w", input, err)
			}
			ipNet.IP = ip
			return *ipNet, nil
		}
	}

	// Parse time.Duration
	if t.PkgPath() == "time" && t.Name() == "Duration" {
		d, err := time.ParseDuration(input)
		if err != nil {
			return nil, fmt.Errorf("unable to parse '%s' as time.Duration, failed with: %w", input, err)
		}
		return d, nil
	}

	// Convert to in-built types
	var (
		v   interface{}
		err error
	)
	switch t.Kind() {
	case reflect.String:
		v = input
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err = strconv.ParseInt(input, 0, t.Bits())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err = strconv.ParseUint(input, 0, t.Bits())
	case reflect.Bool:
		v, err = strconv.ParseBool(input)
	case reflect.Float32, reflect.Float64:
		v, err = strconv.ParseFloat(input, t.Bits())
	case reflect.Slice:
		elems := strings.Split(input, arrayDelimiter)
		values := reflect.MakeSlice(t, len(elems), len(elems))
		for i, elem := range elems {
			convertedValue, err := convertString(t.Elem(), elem)
			if err != nil {
				return nil, err
			}
			values.Index(i).Set(reflect.ValueOf(convertedValue))
		}
		return values.Interface(), nil
	case reflect.Map:
		values := reflect.MakeMap(t)
		keyValues := strings.Split(input, mapDelimiter)
		for _, keyValueUnsplit := range keyValues {
			keyValue := strings.Split(keyValueUnsplit, mapKVDelimiter)
			if len(keyValue) != 2 {
				return nil, fmt.Errorf("invalid key value item provided: %s", keyValueUnsplit)
			}
			key := reflect.New(t.Key()).Elem()
			keyData, err := convertString(key.Type(), keyValue[0])
			if err != nil {
				return nil, err
			}
			key.Set(reflect.ValueOf(keyData))
			value := reflect.New(t.Elem()).Elem()
			valueData, err := convertString(value.Type(), keyValue[1])
			if err != nil {
				return nil, err
			}
			value.Set(reflect.ValueOf(valueData))
			values.SetMapIndex(key, value)
		}
		return values.Interface(), nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind().String())
	}

	if err != nil {
		return nil, err
	}

	// `strconv` functions always return largest type, so the primary purpose of
	// this function is to convert them using reflection, e.g. int64 -> int32
	vv := reflect.ValueOf(v)
	vv = vv.Convert(t)
	return vv.Interface(), nil
}