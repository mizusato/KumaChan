package rpc

import (
	"kumachan/rx"
	"kumachan/rpc/kmd"
	. "kumachan/lang"
)


type Service struct {
	Name         string
	Version      string
	Constructor  ServiceConstructor
	Methods      map[string] ServiceMethod
}

type ServiceConstructor struct {
	ArgType    *kmd.Type
	GetAction  func(Value) rx.Action
}

type ServiceMethod struct {
	ArgType    *kmd.Type
	RetType    *kmd.Type
	GetAction  func(instance Value, arg Value) rx.Action
}

