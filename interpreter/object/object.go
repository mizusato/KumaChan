package object

import "unsafe"

type Object struct {
    __Category  ObjectCategory
    __Inline32  uint32
    __Inline64  uint64
    __Pointer   unsafe.Pointer
}
