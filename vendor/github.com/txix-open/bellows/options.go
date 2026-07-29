package bellows

type bellowsOptions struct {
	prefix string
	sep    string
}

type option func(o *bellowsOptions)

func WithPrefix(prefix string) option {
	return func(o *bellowsOptions) {
		o.prefix = prefix
	}
}

func WithSep(sep string) option {
	return func(o *bellowsOptions) {
		o.sep = sep
	}
}
