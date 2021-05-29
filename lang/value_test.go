package lang

import (
	"testing"
	"unsafe"
)


type Stub struct {
	Something  int
	Extra      bool
}
func (*Stub) Method() {}
func (*Stub) AnotherMethod() {}

func TestRefEqual(t *testing.T) {
	var x = 1
	var f = func() { x++ }
	var g = func() { x++ }
	var h = func() { x-- }
	var stub1 Stub
	var stub2 Stub
	var chan1 = make(chan struct{})
	var chan2 = make(chan struct{})
	var slice1 = [] int { 1, 2 }
	var slice2 = [] int { 1, 2 }
	var slice3 = [] int { 0xbadbeef }
	var map1 = map[int] int { 1: 1, 2: 2 }
	var map2 = map[int] int { 1: 1, 2: 2 }
	var map3 = map[int] int { 0: 0xbadbeef }
	var s1 = "Hello"
	var s2 = string([] rune {'H','e','l','l','o'})
	var s3 = "World"
	var check = func(p bool, expected bool, msg string) {
		if p != expected {
			t.Fatal("wrong behaviour of RefEqual: " + msg)
		}
	}
	// different kind
	check(RefEqual(stub1, f), false, "different kind 1")
	check(RefEqual("1", 1), false, "different kind 2")
	check(RefEqual(int(1), uint(1)), false, "different kind 3")
	check(RefEqual(uint32(1), uint64(1)), false, "different kind 4")
	// struct with 0 element (always true)
	check(RefEqual(struct{}{}, struct{}{}), true, "struct[0]")
	// struct with 1 element (forward)
	check(RefEqual(struct{a int}{a: 1}, struct{a int}{a: 1}), true, "struct[1] 1")
	check(RefEqual(struct{a int}{a: 1}, struct{a int}{a: 2}), false, "struct[1] 2")
	// struct with 2+ elements (value type: always false)
	check(RefEqual(stub1, stub1), false, "struct[2] 1")
	check(RefEqual(stub1, stub2), false, "struct[2] 2")
	// array (value type: always false)
	check(RefEqual([...] int {1,2}, [...] int {1,2}), false, "array 1")
	check(RefEqual([...] int {1,2}, [...] int {0xbadbeef}), false, "array 2")
	// pointer
	check(RefEqual(&stub1, &stub1), true, "pointer (same)")
	check(RefEqual(&stub1, &stub2), false, "pointer (different)")
	check(RefEqual(unsafe.Pointer(&stub1), unsafe.Pointer(&stub1)), true,
		"unsafe pointer (same)")
	check(RefEqual(unsafe.Pointer(&stub1), unsafe.Pointer(&stub2)), false,
		"unsafe pointer (different)")
	// channel
	check(RefEqual(chan1, chan1), true, "channel (same)")
	check(RefEqual(chan1, chan2), false, "channel (different)")
	// function:closure
	check(RefEqual(f, f), true, "closure (same)")
	check(RefEqual(f, g), false, "closure (same underlying)")
	check(RefEqual(f, h), false, "closure (different)")
	// function:method (new pointer returned each time: always false)
	check(RefEqual(stub1.Method, stub1.Method), false, "method 1")
	check(RefEqual(stub1.Method, stub2.Method), false, "method 2")
	check(RefEqual(stub1.Method, stub1.AnotherMethod), false, "method 3")
	// slice
	check(RefEqual(slice1, slice1), true, "slice (same)")
	check(RefEqual(slice1, slice1[:]), true, "slice (same, further sliced)")
	check(RefEqual(slice1, slice2), false, "slice (value equal)")
	check(RefEqual(slice1, slice3), false, "slice (different)")
	check(RefEqual(slice1, slice1[1:]), false, "slice (different, further sliced 1)")
	check(RefEqual(slice1, slice1[:1]), false, "slice (different, further sliced 2)")
	// map
	check(RefEqual(map1, map1), true, "map (same)")
	check(RefEqual(map1, map2), false, "map (value equal)")
	check(RefEqual(map1, map3), false, "map (different)")
	// string
	check(RefEqual(s1, s1), true, "string (same)")
	check(RefEqual(s1, s2), false, "string (value equal)")
	check(RefEqual(s1, s3), false, "string (different)")
	// primitives
	check(RefEqual('1', '1'), true, "rune (same)")
	check(RefEqual('1', '2'), false, "rune (different)")
	check(RefEqual(1.0, 1.0), true, "float64 (same)")
	check(RefEqual(1.0, 2.0), false, "float64 (different)")
}

