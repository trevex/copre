package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type T1 struct {
	A string `yaml:"a" env:"A" flag:"a"`
	B string `yaml:"b" env:"B" flag:"b"`
	C string `yaml:"c" env:"C" flag:"c"`
	D string `yaml:"d" env:"D" flag:"d"`
}

func TestBuilderPrecendence(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Create config file
	t1yaml := `
a: "file"
b: "file"
c: "file"
`
	tf, err := ioutil.TempFile("", "test")
	require.NoError(err)
	defer os.Remove(tf.Name())
	_, err = tf.WriteString(t1yaml)
	require.NoError(err)
	// Prepare FlagSet
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String("b", "", "")
	f.String("c", "", "")
	_ = f.Parse([]string{"--b=flag", "--c=flag"})
	// Setup environment
	os.Setenv("C", "env")
	// Build config and check expected
	t1 := T1{D: "struct"}
	err = NewBuilder(&t1).
		File(tf.Name(), yaml.Unmarshal).
		FlagSet(f).
		Env().Build()
	require.NoError(err)
	assert.Equal("file", t1.A)
	assert.Equal("flag", t1.B)
	assert.Equal("env", t1.C)
	assert.Equal("struct", t1.D)
}
