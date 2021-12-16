package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type T1 struct {
	A string
	B struct {
		C string
		D string `env:"MY_FOO"`
		E string
		F string
		G string `flag:"my-flag"`
		H string `flag:"not-set"`
	}
}

func TestBuilderPrecendence(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	t1yaml := `
b:
  c: "asd"
  d: "asdasd"
  e: "e"
  h: "h"
`
	t1a := "fizzbuzz"
	t1 := T1{A: t1a}
	t1bc := "bar"
	// Setup environment
	os.Setenv("TEST1_B_C", t1bc)
	t1bd := "foo"
	os.Setenv("MY_FOO", t1bd)
	// Create config file
	tf, err := ioutil.TempFile("", "test")
	require.NoError(err)
	defer os.Remove(tf.Name())
	_, err = tf.WriteString(t1yaml)
	require.NoError(err)
	// Prepare FlagSet
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String("b-f", "", "")
	f.String("my-flag", "", "")
	f.String("not-set", "", "")
	t1bf := "flag"
	t1bg := "tag"
	_ = f.Parse([]string{fmt.Sprintf("--b-f=%s", t1bf), fmt.Sprintf("--my-flag=%s", t1bg)})
	// Build config and check expected
	err = NewBuilder(&t1).
		File(tf.Name(), yaml.Unmarshal).
		FlagSet(f).
		Env("TEST1").Build()
	require.NoError(err)
	assert.Equal(t1a, t1.A)
	assert.Equal(t1bc, t1.B.C)
	assert.Equal(t1bd, t1.B.D)
	assert.Equal("e", t1.B.E)
	assert.Equal(t1bf, t1.B.F)
	assert.Equal(t1bg, t1.B.G)
	assert.Equal("h", t1.B.H)
}
