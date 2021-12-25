package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// TODO: including struct preset we have 4! = 24 combinations to test
type T1 struct {
	A string `yaml:"a" env:"A" flag:"a"`
	B string `yaml:"b" env:"B" flag:"b"`
	C string `yaml:"c" env:"C" flag:"c"`
	D string `yaml:"d" env:"D" flag:"d"`
}

func TestBuilderPrecendence(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tests := map[string]struct {
		Yaml     string
		Flags    map[string]string
		EnvVars  map[string]string
		Input    T1
		Expected T1
	}{
		"file>flag>env": {
			Yaml: `
a: "file"
b: "file"
c: "file"
`,
			Flags: map[string]string{
				"b": "flag",
				"c": "flag",
			},
			EnvVars: map[string]string{
				"C": "env",
			},
			Input: T1{
				D: "struct",
			},
			Expected: T1{
				A: "file",
				B: "flag",
				C: "env",
				D: "struct",
			},
		},
	}

	for precendence, test := range tests {
		buildOrder := strings.Split(precendence, ">")
		t.Run(precendence, func(t *testing.T) {
			// Create config file
			tf, err := ioutil.TempFile("", "test")
			require.NoError(err)
			defer os.Remove(tf.Name())
			_, err = tf.WriteString(test.Yaml)
			require.NoError(err)
			// Prepare FlagSet
			f := pflag.NewFlagSet("test", pflag.ContinueOnError)
			args := []string{}
			for key, value := range test.Flags {
				f.String(key, "", "")
				args = append(args, fmt.Sprintf("--%s=%s", key, value))
			}
			_ = f.Parse(args)
			// Setup environment
			for key, value := range test.EnvVars {
				os.Setenv(key, value)
			}
			result := test.Input
			b := NewBuilder(&result)
			for _, phase := range buildOrder {
				if phase == "file" {
					b.File(tf.Name(), yaml.Unmarshal)
				} else if phase == "flag" {
					b.FlagSet(f)
				} else {
					b.Env()
				}
			}
			err = b.Build()
			require.NoError(err)
			assert.Equal(test.Expected, result)
		})
	}
}

func TestBuilderFlagsetMismatch(t *testing.T) {
	require := require.New(t)
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.Int32("a", 0, "")
	_ = f.Parse([]string{"--a=1"})
	t1 := T1{D: "struct"}
	err := NewBuilder(&t1).
		FlagSet(f).
		Build()
	require.Error(err)
}
