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

type ConfigAllSources struct {
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
		Input    ConfigAllSources
		Expected ConfigAllSources
	}{
		"file<flag<env": {
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
			Input: ConfigAllSources{
				D: "struct",
			},
			Expected: ConfigAllSources{
				A: "file",
				B: "flag",
				C: "env",
				D: "struct",
			},
		},
		"env<file<flag": {
			EnvVars: map[string]string{
				"A": "env",
				"B": "env",
				"C": "env",
			},
			Yaml: `
b: "file"
c: "file"
`,
			Flags: map[string]string{
				"c": "flag",
			},
			Input: ConfigAllSources{
				D: "struct",
			},
			Expected: ConfigAllSources{
				A: "env",
				B: "file",
				C: "flag",
				D: "struct",
			},
		},
		"flag<env<file": {
			Flags: map[string]string{
				"a": "flag",
				"b": "flag",
				"c": "flag",
			},
			EnvVars: map[string]string{
				"B": "env",
				"C": "env",
			},
			Yaml: `
c: "file"
`,
			Input: ConfigAllSources{
				D: "struct",
			},
			Expected: ConfigAllSources{
				A: "flag",
				B: "env",
				C: "file",
				D: "struct",
			},
		},
	}

	for precendence, test := range tests {
		buildOrder := strings.Split(precendence, "<")
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
			prefix := strings.ToUpper(strings.ReplaceAll(precendence, "<", "_"))
			for key, value := range test.EnvVars {
				os.Setenv(prefix+"_"+key, value)
			}
			result := test.Input
			loaders := []Loader{}
			for _, phase := range buildOrder {
				if phase == "file" {
					loaders = append(loaders, File(tf.Name(), yaml.Unmarshal))
				} else if phase == "flag" {
					loaders = append(loaders, FlagSet(f))
				} else {
					loaders = append(loaders, Env(WithPrefix(prefix)))
				}
			}
			err = Load(&result, loaders...)
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
	result := ConfigAllSources{D: "struct"}
	err := Load(&result, FlagSet(f))
	require.Error(err)
}

type ConfigEnvOptions struct {
	A []byte `env:"NOPREFIX_A,noprefix"`
	B []byte `env:"B,hex"`
	C []byte `env:",base64"`
}

func TestBuilderEnvOptions(t *testing.T) {
	require := require.New(t)
	expected := ConfigEnvOptions{
		A: []byte{1, 1},
		B: []byte{255},
		C: []byte("1"),
	}
	prefix := "ENV_OPTIONS"
	os.Setenv("NOPREFIX_A", "1,1")
	os.Setenv(prefix+"_B", "FF")
	os.Setenv(prefix+"_C", "MQ==")
	result := ConfigEnvOptions{}
	err := Load(&result, Env(WithPrefix(prefix), ComputeEnvKey(UpperSnakeCase)))
	require.NoError(err)
	require.Equal(expected, result)
}

type ConfigEnvOptionsUnsupported struct {
	A string `env:"A,hex"`
}

func TestBuilderEnvOptionUnsupported(t *testing.T) {
	result := ConfigEnvOptionsUnsupported{}
	err := Load(&result, Env())
	require.Error(t, err)
}
