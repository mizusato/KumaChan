package checker

import (
	"fmt"
	"kumachan/loader"
	. "kumachan/error"
)

type TypeExprError struct {
	Point     ErrorPoint
	Concrete  ConcreteTypeExprError
}

type ConcreteTypeExprError interface { TypeExprError() }

func (impl E_ModuleOfTypeRefNotFound) TypeExprError() {}
type E_ModuleOfTypeRefNotFound struct {
	Name   string
}
func (impl E_TypeNotFound) TypeExprError() {}
type E_TypeNotFound struct {
	Name   loader.Symbol
}
func (impl E_NativeTypeNotFound) TypeExprError() {}
type E_NativeTypeNotFound struct {
	Name   string
}
func (impl E_WrongParameterQuantity) TypeExprError() {}
type E_WrongParameterQuantity struct {
	TypeName  loader.Symbol
	Required  uint
	Given     uint
}
func (impl E_DuplicateField) TypeExprError() {}
type E_DuplicateField struct {
	FieldName  string
}

func (err *TypeExprError) Error() string {
	var description string
	switch e := err.Concrete.(type) {
	case E_ModuleOfTypeRefNotFound:
		description = fmt.Sprintf (
			"%vNo such module: %v%s%v",
			Red, Bold, e.Name, Reset,
		)
	case E_TypeNotFound:
		description = fmt.Sprintf (
			"%vNo such type: %v%s%v",
			Red, Bold, e.Name, Reset,
		)
	case E_NativeTypeNotFound:
		description = fmt.Sprintf (
			"%vNo such native type: %v%s%v",
			Red, Bold, e.Name, Reset,
		)
	case E_WrongParameterQuantity:
		description = fmt.Sprintf (
			"%vWrong parameter quantity: %v%d%v required but %v%d%v given%v",
			Red, Bold, e.Required, Reset+Red, Bold, e.Given, Reset+Red, Reset,
		)
	case E_DuplicateField:
		description = fmt.Sprintf (
			"%vDuplicate field: %v%s%v",
			Red, Bold, e.FieldName, Reset,
		)
	default:
		panic("unknown concrete error type")
	}
	return err.Point.GenErrMsg(description)
}


type TypeDeclError struct {
	Point     ErrorPoint
	Concrete  ConcreteTypeDeclError
}

type ConcreteTypeDeclError interface { TypeDeclError() }

func (impl E_DuplicateTypeDecl) TypeDeclError() {}
type E_DuplicateTypeDecl struct {
	TypeName  loader.Symbol
}
func (impl E_GenericUnionSubType) TypeDeclError() {}
type E_GenericUnionSubType struct {
	TypeName  loader.Symbol
}
func (impl E_InvalidTypeDecl) TypeDeclError() {}
type E_InvalidTypeDecl struct {
	TypeName   loader.Symbol
	ExprError  *TypeExprError
}

func (err *TypeDeclError) Error() string {
	var cause interface {}
	var errors = make([]string, 0)
	switch e := err.Concrete.(type) {
	case E_DuplicateTypeDecl:
		cause = e
		errors = append(errors, err.Point.GenErrMsg(fmt.Sprintf (
			"%vDuplicate type declaration: %v%s%v",
			Red, Bold, e.TypeName, Reset,
		)))
	case E_GenericUnionSubType:
		cause = e
		errors = append(errors, err.Point.GenErrMsg(fmt.Sprintf (
			"%vCannot define generic paramters on a subtype of a union type%v",
			Red, Reset,
		)))
	case E_InvalidTypeDecl:
		cause = e.ExprError.Concrete
		errors = append(errors, err.Point.GenErrMsg(fmt.Sprintf (
			"%vInvalid definition of type %v%s%v",
			Red, Bold, e.TypeName, Reset,
		)))
		errors = append(errors, e.ExprError.Error())
	default:
		panic("unknown concrete error type")
	}
	return GenCompilationFailedMessage(cause, errors)
}