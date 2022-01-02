package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfigAllSources struct {
	A string `json:"a" env:"A" flag:"a"`
	B string `json:"b" env:"B" flag:"b"`
	C string `json:"c" env:"C" flag:"c"`
	D string `json:"d" env:"D" flag:"d"`
}

func TestLoaderPrecendence(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tests := map[string]struct {
		JSON     string
		Flags    map[string]string
		EnvVars  map[string]string
		Input    TestConfigAllSources
		Expected TestConfigAllSources
	}{
		"file<flag<env": {
			JSON: `{
	"a": "file",
	"b": "file",
	"c": "file"
}`,
			Flags: map[string]string{
				"b": "flag",
				"c": "flag",
			},
			EnvVars: map[string]string{
				"C": "env",
			},
			Input: TestConfigAllSources{
				D: "struct",
			},
			Expected: TestConfigAllSources{
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
			JSON: `{
	"b": "file",
	"c": "file"
}`,
			Flags: map[string]string{
				"c": "flag",
			},
			Input: TestConfigAllSources{
				D: "struct",
			},
			Expected: TestConfigAllSources{
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
			JSON: `{ "c": "file" }`,
			Input: TestConfigAllSources{
				D: "struct",
			},
			Expected: TestConfigAllSources{
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
			_, err = tf.WriteString(test.JSON)
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
					loaders = append(loaders, File(tf.Name(), json.Unmarshal))
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
