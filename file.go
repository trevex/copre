package config

import (
	"fmt"
	"io/ioutil"
	"os"
	pathfilepath "path/filepath"
)

type fileOptions struct {
	ignoreNotFound bool
	expandEnv      bool
	mergeFiles     bool
	searchPaths    []string
}

type FileOption interface {
	apply(*fileOptions)
}

type fileOptionAdapter func(*fileOptions)

func (c fileOptionAdapter) apply(o *fileOptions) {
	c(o)
}

func UseSearchPaths(paths ...string) FileOption {
	return fileOptionAdapter(func(o *fileOptions) {
		o.searchPaths = paths
	})
}

func MergeFiles(f ...bool) FileOption {
	return fileOptionAdapter(func(o *fileOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		o.mergeFiles = v
	})
}

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
		mergeFiles:     false,
		searchPaths:    []string{},
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return LoaderFunc(func(dst interface{}) error {
		// Let's compute the list of filepaths to check
		filepaths := []string{filepath}
		filename := pathfilepath.Base(filepath)
		for _, searchPath := range o.searchPaths {
			filepaths = append(filepaths, pathfilepath.Join(searchPath, filename))
		}

		// Okay, let's load the files
		var (
			fileData = map[string][]byte{}
			err      error
		)

		for _, fp := range filepaths {
			var d []byte
			d, err = ioutil.ReadFile(fp)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			if err != nil {
				continue
			}
			if o.expandEnv {
				d = []byte(os.ExpandEnv(string(d)))
			}
			fileData[fp] = d
			if !o.mergeFiles { // If we only want the first file we find, stop here
				break
			}
		}

		if o.ignoreNotFound && len(fileData) == 0 {
			return nil
		}
		if len(fileData) == 0 {
			return fmt.Errorf("no file loaded, last error was: %w", err)
		}

		for fp, d := range fileData {
			if err := unmarshal(d, dst); err != nil {
				return fmt.Errorf("failed to unmarshal '%s': %w", fp, err)
			}
		}

		return nil
	})
}
