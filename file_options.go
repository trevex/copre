package config

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
