package bellows

import (
	"fmt"
	"reflect"
)

func Flatten(value interface{}, opts ...option) map[string]interface{} {
	options := &bellowsOptions{
		prefix: "",
		sep:    ".",
	}
	for _, opt := range opts {
		opt(options)
	}
	m := make(map[string]interface{}, 5)
	FlattenPrefixedToResult(value, options, m)
	return m
}

func FlattenPrefixedToResult(value interface{}, opts *bellowsOptions, m map[string]interface{}) {
	original := reflect.ValueOf(value)
	kind := original.Kind()
	if kind == reflect.Ptr || kind == reflect.Interface {
		original = reflect.Indirect(original)
		kind = original.Kind()
	}

	if !original.IsValid() {
		if opts.prefix != "" {
			m[opts.prefix] = nil
		}
		return
	}

	t := original.Type()

	switch kind {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			break
		}
		keys := original.MapKeys()
		base := ""
		if opts.prefix != "" {
			base = opts.prefix + opts.sep
		}
		for _, childKey := range keys {
			childValue := original.MapIndex(childKey)
			FlattenPrefixedToResult(childValue.Interface(), &bellowsOptions{
				prefix: base + childKey.String(),
				sep:    opts.sep,
			}, m)
		}
	case reflect.Struct:
		numField := original.NumField()
		base := ""
		for i := 0; i < numField; i++ {
			childValue := original.Field(i)
			f := t.Field(i)
			childKey := f.Name
			if f.Anonymous {
				childKey = ""
				base = opts.prefix
			} else if opts.prefix != "" {
				base = opts.prefix + opts.sep
			}
			FlattenPrefixedToResult(childValue.Interface(), &bellowsOptions{
				prefix: base + childKey,
				sep:    opts.sep,
			}, m)
		}
	case reflect.Array, reflect.Slice:
		l := original.Len()
		base := opts.prefix
		for i := 0; i < l; i++ {
			childValue := original.Index(i)
			FlattenPrefixedToResult(childValue.Interface(), &bellowsOptions{
				prefix: fmt.Sprintf("%s%s[%d]", base, opts.sep, i),
				sep:    opts.sep,
			}, m)
		}
	default:
		if opts.prefix != "" {
			m[opts.prefix] = value
		}
	}
}
