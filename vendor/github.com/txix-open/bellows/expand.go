package bellows

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	arrayIndexRegexp = regexp.MustCompile("\\[\\d*]")
)

func Expand(flatMap map[string]interface{}, opts ...option) interface{} {
	options := &bellowsOptions{sep: "."}
	for _, opt := range opts {
		opt(options)
	}
	var dst interface{}
	for path, value := range flatMap {
		parts := strings.Split(path, options.sep)
		dst = put(dst, parts, value)
	}
	return dst
}

func put(dst interface{}, path []string, value interface{}) interface{} {
	if len(path) == 0 {
		return value
	}

	p := path[0]
	index, isArray := getArrayIndex(p)
	if isArray {
		if dst == nil {
			dst = make([]interface{}, 0, 3)
		}
		if arr, ok := dst.([]interface{}); ok {
			i := len(arr)
			if i == index {
				arr = append(arr, put(nil, path[1:], value))
			} else if i < index {
				toInsert := make([]interface{}, index-i)
				newItem := put(nil, path[1:], value)
				switch newItem.(type) {
				case []interface{}:
					for i := range toInsert {
						toInsert[i] = make([]interface{}, 0)
					}
				case map[string]interface{}:
					for i := range toInsert {
						toInsert[i] = make(map[string]interface{}, 0)
					}
				}
				arr = append(arr[:i], append(toInsert, newItem)...)
			} else {
				arr[index] = put(arr[index], path[1:], value)
			}

			dst = arr
		}
	} else {
		if dst == nil {
			dst = make(map[string]interface{}, 3)
		}
		if m, ok := dst.(map[string]interface{}); ok {
			if val, ok := m[p]; ok {
				m[p] = put(val, path[1:], value)
			} else {
				m[p] = put(nil, path[1:], value)
			}
		}
	}

	return dst
}

func getArrayIndex(part string) (int, bool) {
	index := arrayIndexRegexp.FindString(part)
	if index == "" {
		return 0, false
	}

	i, err := strconv.Atoi(index[1 : len(index)-1])
	if err != nil {
		return 0, false
	}

	return i, true
}
