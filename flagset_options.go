package config

type flagSetOptions struct {
	includeUnchanged bool
	useFlagDefaults  bool
}

type FlagSetOption interface {
	apply(*flagSetOptions)
}

type flagSetOptionAdapter func(*flagSetOptions)

func (c flagSetOptionAdapter) apply(l *flagSetOptions) {
	c(l)
}

// IncludeUnchanged will also process the values of unchanged flags. Effectively
// this means the flag defaults, if non zero, will be set as well.
func IncludeUnchanged(f ...bool) FlagSetOption {
	return flagSetOptionAdapter(func(l *flagSetOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		l.includeUnchanged = v
	})
}

// UseFlagDefaults will run and check the FlagSet once before any other method
// is called to set the FlagSet defaults to the destination config. The changed
// values of the FlagSet still respect the defined precendence.
func UseFlagDefaults(f ...bool) FlagSetOption {
	return flagSetOptionAdapter(func(l *flagSetOptions) {
		v := true
		if len(f) > 0 {
			v = f[0]
		}
		l.useFlagDefaults = v
	})
}
