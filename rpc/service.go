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

type ServiceConstructor struct {
	ArgType    *kmd.Type
	GetAction  func(object kmd.Object) rx.Action
}

type ServiceMethod struct {
	ArgType    *kmd.Type
	RetType    *kmd.Type
	GetAction  func(instance kmd.Object, arg kmd.Object) rx.Action
}

