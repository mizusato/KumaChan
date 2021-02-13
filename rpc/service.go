package rpc

import (
	"kumachan/rx"
	"kumachan/rpc/kmd"
)



type Service struct {
	ServiceIdentifier
	Constructor  ServiceConstructor
	Methods      map[string] ServiceMethod
}
type ServiceInterface struct {
	ServiceIdentifier
	Constructor  ServiceConstructorInterface
	Methods      map[string] ServiceMethodInterface
}
type ServiceIdentifier struct {
	Vendor   string
	Project  string
	Name     string
	Version  string
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

