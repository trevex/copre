package config

import (
	"io/ioutil"
	"os"
)

type UnmarshalFunc func(data []byte, dst interface{}) error

type fileLoader struct {
	filepath       string
	unmarshal      UnmarshalFunc
	ignoreNotFound bool
	expandEnv      bool
}

type FileOption interface {
	apply(*fileLoader)
}

type fileOptionAdapter func(*fileLoader)

func (c fileOptionAdapter) apply(l *fileLoader) {
	c(l)
}

// TODO: UseSearchPaths(paths []string) Option to look for a file in
//       several directories.

func IgnoreNotFound(f ...bool) FileOption {
	return fileOptionAdapter(func(l *fileLoader) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		l.ignoreNotFound = v
	})
}

func ExpandEnv(f ...bool) FileOption {
	return fileOptionAdapter(func(l *fileLoader) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		l.expandEnv = v
	})
}

func NewFileLoader(path string, unmarshal UnmarshalFunc, opts ...FileOption) Loader {
	l := fileLoader{
		filepath:  path,
		unmarshal: unmarshal,
	}
	for _, opt := range opts {
		opt.apply(&l)
	}
	return &l
}

func (l *fileLoader) Preprocess(dst interface{}) error {
	return nil
}

func (l *fileLoader) Process(dst interface{}) error {
	d, err := ioutil.ReadFile(l.filepath)
	if l.ignoreNotFound && os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if l.expandEnv {
		d = []byte(os.ExpandEnv(string(d)))
	}
	if err := l.unmarshal(d, dst); err != nil {
		return err
	}
	return nil
}
