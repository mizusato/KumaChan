package object

import "unsafe"

const MapShrinkFactor = 2

type DictMap interface {
	Has(key string) bool
	Get(key string) Object
	Set(key string, value Object)
	Delete(key string)
}

type IntMap interface {
	Has(key int) bool
	Get(key int) Object
	Set(key int, value Object)
	Delete(key int)
}

type DictMap2Val struct {
	__Category  ObjectCategory
	__Data      map[string]uint64
	__Added     int
}

type DictMap2Ref struct {
	__Category  ObjectCategory
	__Data      map[string]unsafe.Pointer
	__Added     int
}

type DictMap2Variant struct {
	__Data   map[string]Object
	__Added  int
}

type IntMap2Val struct {
	__Category  ObjectCategory
	__Data      map[int]uint64
	__Added     int
}

type IntMap2Ref struct {
	__Category  ObjectCategory
	__Data      map[int]unsafe.Pointer
	__Added     int
}

type IntMap2Variant struct {
	__Data   map[int]Object
	__Added  int
}

