package object

import "unsafe"

type Object struct {
    __Category  ObjectCategory
    __Inline    uint64
    __Pointer   unsafe.Pointer
}
