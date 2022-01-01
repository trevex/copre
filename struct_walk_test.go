package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructWalker(t *testing.T) {
	assert := assert.New(t)
	dst := struct {
		A string
		B *struct {
			C string
		}
	}{}
	expectedPaths := [][]string{{"A"}, {"B", "C"}}
	paths := [][]string{}
	err := StructWalk(&dst, func(path []string, field reflect.StructField) (interface{}, error) {
		paths = append(paths, path)
		return "value", nil
	})
	assert.NoError(err)
	assert.Equal(expectedPaths, paths)
	assert.Equal("value", dst.A)
	assert.Equal("value", dst.B.C)
}

func TestStructWalkerIncompatibleTypes(t *testing.T) {
	assert := assert.New(t)
	dst := struct {
		A string
	}{}
	err := StructWalk(&dst, func(path []string, field reflect.StructField) (interface{}, error) {
		return 1, nil
	})
	assert.Error(err)
}
