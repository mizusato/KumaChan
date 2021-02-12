package rpc

import (
	"kumachan/rx"
	"kumachan/rpc/kmd"
)


type Service struct {
	Name         string
	Version      string
	Constructor  ServiceConstructor
	Methods      map[string] ServiceMethod
}
type ServiceInterface struct {
	Name         string
	Version      string
	Constructor  ServiceConstructorInterface
	Methods      map[string] ServiceMethodInterface
}

type ServiceConstructor struct {
	ArgType    *kmd.Type
	GetAction  func(object kmd.Object) rx.Action
}
type ServiceConstructorInterface struct {
	ArgType    *kmd.Type
}

type ServiceMethod struct {
	ArgType    *kmd.Type
	RetType    *kmd.Type
	GetAction  func(instance kmd.Object, arg kmd.Object) rx.Action
}
type ServiceMethodInterface struct {
	ArgType    *kmd.Type
	RetType    *kmd.Type
}

