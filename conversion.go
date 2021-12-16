package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TODO: Support time.Time and time.Duration?

type convertStringConfig struct {
	ArrayDelimiter string
	MapDelimiter   string
	MapKVDelimiter string
}

var defaultConvertStringConfig convertStringConfig = convertStringConfig{
	ArrayDelimiter: ",",
	MapDelimiter:   ",",
	MapKVDelimiter: "=",
}

// Converts `input` string to type `t` or returns error if operation is not
// possible. Supported target types are strings, bools, floats,
// integers (incl. unsigned). Slices and maps of those types are supported as well.
func convertString(t reflect.Type, input string, c *convertStringConfig) (interface{}, error) {
	if c == nil {
		return nil, fmt.Errorf("convertStringConfig missing")
	}
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
		elems := strings.Split(input, c.ArrayDelimiter)
		values := reflect.MakeSlice(t, len(elems), len(elems))
		for i, elem := range elems {
			convertedValue, err := convertString(t.Elem(), elem, c)
			if err != nil {
				return nil, err
			}
			values.Index(i).Set(reflect.ValueOf(convertedValue))
		}
		return values.Interface(), nil
	case reflect.Map:
		values := reflect.MakeMap(t)
		keyValues := strings.Split(input, c.MapDelimiter)
		for _, keyValueUnsplit := range keyValues {
			keyValue := strings.Split(keyValueUnsplit, c.MapKVDelimiter)
			if len(keyValue) != 2 {
				return nil, fmt.Errorf("invalid key value item provided: %s", keyValueUnsplit)
			}
			key := reflect.New(t.Key()).Elem()
			keyData, err := convertString(key.Type(), keyValue[0], c)
			if err != nil {
				return nil, err
			}
			key.Set(reflect.ValueOf(keyData))
			value := reflect.New(t.Elem()).Elem()
			valueData, err := convertString(value.Type(), keyValue[1], c)
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
