package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisitStructsInStructs(t *testing.T) {
	assert := assert.New(t)
	dst := struct {
		A string
		B *struct {
			C string
		}
	}{}
	expectedPaths := [][]string{{"A"}, {"B", "C"}}
	paths := [][]string{}
	visitStruct(&dst, func(path []string, field reflect.StructField) (interface{}, error) {
		paths = append(paths, path)
		return nil, nil
	})
	assert.Equal(expectedPaths, paths)
}
