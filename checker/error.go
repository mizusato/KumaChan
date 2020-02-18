package checker

import (
	"fmt"
	. "kumachan/error"
	"kumachan/loader"
)


type TypeError struct {
	Point     ErrorPoint
	Concrete  ConcreteTypeError
}

type ConcreteTypeError interface { TypeError() }

func (impl E_ModuleOfTypeRefNotFound) TypeError() {}
type E_ModuleOfTypeRefNotFound struct {
	Name   string
}
func (impl E_TypeNotFound) TypeError() {}
type E_TypeNotFound struct {
	Name   loader.Symbol
}
func (impl E_WrongParameterQuantity) TypeError() {}
type E_WrongParameterQuantity struct {
	TypeName  loader.Symbol
	Required  uint
	Given     uint
}
func (impl E_DuplicateField) TypeError() {}
type E_DuplicateField struct {
	FieldName  string
}

func (err *TypeError) Error() string {
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
	ExprError  *TypeError
}

func (err *TypeDeclError) Error() string {
	var cause interface {}
	var errors = make([]string, 0)
	switch e := err.Concrete.(type) {
	case E_DuplicateTypeDecl:
		cause = e
		errors = append(errors, err.Point.GenErrMsg(fmt.Sprintf (
			"%vDuplicate type declaration: %v%s%v",
			Red, Bold, e.TypeName.SymbolName, Reset,
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
			Red, Bold, e.TypeName.SymbolName, Reset,
		)))
		errors = append(errors, e.ExprError.Error())
	default:
		panic("unknown concrete error type")
	}
	return GenCompilationFailedMessage(cause, errors)
}


type FunctionError struct {
	Point     ErrorPoint
	Concrete  ConcreteFunctionError
}

type ConcreteFunctionError interface { FunctionError() }

func (impl E_SignatureInvalid) FunctionError() {}
type E_SignatureInvalid struct {
	FuncName   string
	TypeError  *TypeError
}

func (E_SignatureNonLocal) FunctionError() {}
type E_SignatureNonLocal struct {
	FuncName  string
}

func (E_InvalidOverload) FunctionError() {}
type E_InvalidOverload struct {
	FuncName         string
	IsLocalConflict  bool
}


type ExprError struct {
	Point     ErrorPoint
	Concrete  ConcreteExprError
}

type ConcreteExprError interface { ExprError() }

func (impl E_NotAssignable) ExprError() {}
type E_NotAssignable struct {
	From    string
	To      string
	Reason  string
}
