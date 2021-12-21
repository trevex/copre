package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertScalars(t *testing.T) {
	assert := assert.New(t)
	input := "1"
	conversions := []interface{}{
		"1",
		uint8(1),
		uint16(1),
		uint32(1),
		uint64(1),
		uint(1),
		int8(1),
		int16(1),
		int32(1),
		int64(1),
		int(1),
		float32(1.0),
		float64(1.0),
		true,
	}
	for _, expected := range conversions {
		t.Logf("converting to %T", expected)
		converted, err := convertString(reflect.TypeOf(expected), input)
		assert.NoError(err)
		assert.Equal(expected, converted)
	}
}

func TestConvertSlices(t *testing.T) {
	assert := assert.New(t)
	input := "1,0"
	conversions := []interface{}{
		[]string{"1", "0"},
		[]uint8{1, 0},
		[]uint16{1, 0},
		[]uint32{1, 0},
		[]uint64{1, 0},
		[]uint{1, 0},
		[]int8{1, 0},
		[]int16{1, 0},
		[]int32{1, 0},
		[]int64{1, 0},
		[]int{1, 0},
		[]float32{1.0, 0.0},
		[]float64{1.0, 0.0},
		[]bool{true, false},
	}
	for _, expected := range conversions {
		t.Logf("converting to %T", expected)
		converted, err := convertString(reflect.TypeOf(expected), input)
		assert.NoError(err)
		assert.Equal(expected, converted)
	}
}

func TestConvertMaps(t *testing.T) {
	assert := assert.New(t)
	input := "0=1,1=0"
	conversions := []interface{}{
		map[string]uint8{"0": 1, "1": 0},
		map[string]int{"0": 1, "1": 0},
		map[string]float32{"0": 1.0, "1": 0.0},
		map[bool]bool{false: true, true: false},
		map[uint64]string{0: "1", 1: "0"},
	}
	for _, expected := range conversions {
		t.Logf("converting to %T", expected)
		converted, err := convertString(reflect.TypeOf(expected), input)
		assert.NoError(err)
		assert.Equal(expected, converted)
	}
}

func TestConvertErrors(t *testing.T) {
	assert := assert.New(t)
	input := "1"
	conversions := []interface{}{
		struct{}{},
	}
	for _, expected := range conversions {
		_, err := convertString(reflect.TypeOf(expected), input)
		assert.Error(err)
	}
}
