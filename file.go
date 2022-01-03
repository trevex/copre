package copre

import (
	"fmt"
	"io/ioutil"
	"os"
)

type fileOptions struct {
	ignoreNotFound bool
	expandEnv      bool
	mergeFiles     bool
	filePaths      []string
}

// FileOption configures how given configuration files are used to populate a given structure.
type FileOption interface {
	apply(*fileOptions)
}

type fileOptionAdapter func(*fileOptions)

func (c fileOptionAdapter) apply(o *fileOptions) {
	c(o)
}

// AppendFilePaths appends paths to the list of paths used to locate
// configuration files.
//
// See File for details on how configuration files are located.
func AppendFilePaths(paths ...string) FileOption {
	return fileOptionAdapter(func(o *fileOptions) {
		o.filePaths = append(o.filePaths, paths...)
	})
}

// MergeFiles changes the default behavior of using the first file found to
// load configuration. Instead all files that are available will be loaded and
// unmarshalled into the configuration struct.
func MergeFiles(f ...bool) FileOption {
	return fileOptionAdapter(func(o *fileOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		o.mergeFiles = v
	})
}

// IgnoreNotFound surpresses File from returning fs.ErrNotExist errors
// effectively making the configuration file optional.
func IgnoreNotFound(f ...bool) FileOption {
	return fileOptionAdapter(func(o *fileOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		o.ignoreNotFound = v
	})
}

// ExpandEnv expands environment variables in loaded configuration files.
func ExpandEnv(f ...bool) FileOption {
	return fileOptionAdapter(func(o *fileOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		o.expandEnv = v
	})
}

// UnmarshalFunc is a function that File can use to unmarshal data into a struct.
// Compatible with the common signature provided by json.Unmarshal, yaml.Unmarshal and similar.
type UnmarshalFunc func(data []byte, dst interface{}) error

// File implements a Loader, that uses a file or files to retrieve configuration values.
//
// By default the provided filePath is used. However this behaviour can be changed
// using options.
// If File should look in multiple locations, additional paths can be appended
// using AppendFilePaths. File will check the existence of those files one by one
// and load the first found.
// If MergeFiles is specified, all files will be loaded and unmarshalled in the
// order specified by the search paths.
//
// Simple standalone example:
//  err := File("/etc/myapp/config.json", json.Unmarshal, IgnoreNotFound()).Process(&cfg)
// Advanced standalone example:
// 	err := File("./config.json", json.Unmarshal,
//    AppendFilePaths("/etc/myapp/myapp.json", path.Join(userHomeDir, ".config/myapp/myapp.json")),
//    MergeFiles(),
//  ).Process(&cfg)
func File(filePath string, unmarshal UnmarshalFunc, opts ...FileOption) Loader {
	o := fileOptions{
		ignoreNotFound: false,
		expandEnv:      false,
		mergeFiles:     false,
		filePaths:      []string{filePath},
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return LoaderFunc(func(dst interface{}) error {
		// Okay, let's load the files
		var (
			fileData = map[string][]byte{}
			err      error
		)

		for _, fp := range o.filePaths {
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
