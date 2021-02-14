package rpc

import (
	"fmt"
	"errors"
	"strings"
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
type ServiceIndex  map[ServiceIdentifier] ServiceInterface

type ServiceIdentifier struct {
	Vendor   string
	Project  string
	Name     string
	Version  string
}
func DescribeServiceIdentifier(id ServiceIdentifier) string {
	return fmt.Sprintf("%s:%s:%s:%s",
		id.Vendor, id.Project, id.Name, id.Version)
}
func ParseServiceIdentifier(str string) (ServiceIdentifier, error) {
	var t = strings.Split(str, ":")
	if len(t) != 4 {
		return ServiceIdentifier{}, errors.New("bad service identifier")
	}
	// TODO: maybe more validations needed
	return ServiceIdentifier {
		Vendor:  t[0],
		Project: t[1],
		Name:    t[2],
		Version: t[3],
	}, nil
}

type ServiceConstructor struct {
	ServiceConstructorInterface
	GetAction  func(object kmd.Object) rx.Action
}
type ServiceConstructorInterface struct {
	ArgType  *kmd.Type
}

type ServiceMethod struct {
	ServiceMethodInterface
	GetAction  func(instance kmd.Object, arg kmd.Object) rx.Action
}
type ServiceMethodInterface struct {
	ArgType     *kmd.Type
	RetType     *kmd.Type
	MultiValue  bool
}

