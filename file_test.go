package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfigFileOptions struct {
	A string `json:"a"`
	B string `json:"b"`
	C string `json:"c"`
}

func TestFileOptions(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	dir, err := ioutil.TempDir("", "fileoptions")
	require.NoError(err)
	defer os.RemoveAll(dir)

	pa := filepath.Join(dir, "a")
	pb := filepath.Join(dir, "b")
	pc := filepath.Join(dir, "c")
	for _, p := range []string{pa, pb, pc} {
		err = os.Mkdir(p, 0700)
		require.NoError(err)
	}
	fa := filepath.Join(pa, "d.json")
	fb := filepath.Join(pb, "d.json")
	fc := filepath.Join(pc, "d.json")

	err = os.WriteFile(fa, []byte(`{ "a": "a"}`), 0600)
	require.NoError(err)
	err = os.WriteFile(fb, []byte(`{ "b": "b"}`), 0600)
	require.NoError(err)
	err = os.WriteFile(fc, []byte(`{ "c": "c"}`), 0600)
	require.NoError(err)

	t.Run("MergeAll", func(t *testing.T) {
		result := TestConfigFileOptions{}
		err = File(fa, json.Unmarshal,
			UseSearchPaths(pb, pc),
			MergeFiles(),
		).Process(&result)
		require.NoError(err)
		assert.Equal("a", result.A)
		assert.Equal("b", result.B)
		assert.Equal("c", result.C)
	})
	t.Run("FirstFound", func(t *testing.T) {
		result := TestConfigFileOptions{}
		err = File(fa, json.Unmarshal,
			UseSearchPaths(pb, pc),
		).Process(&result)
		require.NoError(err)
		assert.Equal("a", result.A)
		assert.Equal("", result.B)
		assert.Equal("", result.C)
	})
}

func TestFileIgnoreNotFound(t *testing.T) {
	fp := path.Join("does", "not", "exist")
	result := TestConfigFileOptions{}
	t.Run("WithOptionNoError", func(t *testing.T) {
		err := File(fp, json.Unmarshal, IgnoreNotFound()).Process(&result)
		require.NoError(t, err)
	})
	t.Run("WithoutOptionError", func(t *testing.T) {
		err := File(fp, json.Unmarshal).Process(&result)
		require.Error(t, err)
	})
}
