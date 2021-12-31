package config

import (
	"io/ioutil"
	"os"
)

type fileOptions struct {
	ignoreNotFound bool
	expandEnv      bool
}

type FileOption interface {
	apply(*fileOptions)
}

type fileOptionAdapter func(*fileOptions)

func (c fileOptionAdapter) apply(l *fileOptions) {
	c(l)
}

// TODO: UseSearchPaths(paths []string) Option to look for a file in
//       several directories.

// TODO: Filepath overrides from env and flags!

func IgnoreNotFound(f ...bool) FileOption {
	return fileOptionAdapter(func(o *fileOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		o.ignoreNotFound = v
	})
}

func ExpandEnv(f ...bool) FileOption {
	return fileOptionAdapter(func(o *fileOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		o.expandEnv = v
	})
}

func File(filepath string, unmarshal UnmarshalFunc, opts ...FileOption) Loader {
	o := fileOptions{
		ignoreNotFound: false,
		expandEnv:      false,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return LoaderFunc(func(dst interface{}) error {
		d, err := ioutil.ReadFile(filepath)
		if o.ignoreNotFound && os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if o.expandEnv {
			d = []byte(os.ExpandEnv(string(d)))
		}
		if err := unmarshal(d, dst); err != nil {
			return err
		}
		return nil
	})
}
