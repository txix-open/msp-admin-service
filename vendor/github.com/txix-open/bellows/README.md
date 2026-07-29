# bellows
Flatten and expand golang maps and structs

## Features
There are some existing golang flatten/expand implementations, but they are targeted to specific use-cases, and none of them flatten Structs.

This `Flatten` will skip parsing a Map that does not have a String as the index.

This `Flatten` recursively passes values as their original Interface, so it is simpler than implementations that rely on passing reflect.Value.

## Usage

```go
// Expand a dot-separated flat map into a nested maps and slices
func Expand(flat map[string]interface{}) interface{}

// Flatten a nested map into a dot-separated flat map
func Flatten(value interface{}) map[string]interface{} {}

// Flatten a nested map into a dot-separated flat map, with a prefix
func FlattenPrefixed(value interface{}, prefix string) map[string]interface{} {}

// Flatten a nested map into an existing dot-separated flat map, with a prefix
func FlattenPrefixedToResult(value interface{}, prefix string, m map[string]interface{}) {}
```

## Other golang flatten/expand implementations

  * [hashicorp/terraform/flatmap](https://github.com/hashicorp/terraform/tree/master/flatmap)
  * [turtlemonvh/mapmap](https://github.com/turtlemonvh/mapmap)
  * [peking2/func-go](https://github.com/peking2/func-go)
  * [wolfeidau/unflatten](https://github.com/wolfeidau/unflatten)
  * [jeremywohl/flatten](https://github.com/jeremywohl/flatten)

(C) Copyright 2015 doublerebel. MIT Licensed.
