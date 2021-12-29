package config

type envOptions struct {
	prefix    string
	keyGetter func([]string) string
}

type EnvOption interface {
	apply(*envOptions)
}

type envOptionAdapter func(*envOptions)

func (c envOptionAdapter) apply(l *envOptions) {
	c(l)
}

func WithPrefix(prefix string) EnvOption {
	return envOptionAdapter(func(o *envOptions) {
		o.prefix = prefix
	})
}

func ComputeEnvKey(keyGetter func([]string) string) EnvOption {
	return envOptionAdapter(func(o *envOptions) {
		o.keyGetter = keyGetter
	})
}
