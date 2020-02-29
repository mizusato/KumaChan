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

func (impl E_InvalidTypeName) TypeDeclError() {}
type E_InvalidTypeName struct {
	Name  string
}
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
	case E_InvalidTypeName:
		cause = e
		errors = append(errors, err.Point.GenErrMsg(fmt.Sprintf (
			"%vInvalid type name: %v%s%v",
			Red, Bold, e.Name, Reset,
		)))
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

func (impl E_InvalidFunctionName) FunctionError() {}
type E_InvalidFunctionName struct {
	Name  string
}

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


type ConstantError struct {
	Point     ErrorPoint
	Concrete  ConcreteConstantError
}

type ConcreteConstantError interface { ConstantError() }

func (impl E_InvalidConstName) ConstantError() {}
type E_InvalidConstName struct {
	Name  string
}

func (impl E_DuplicateConstDecl) ConstantError() {}
type E_DuplicateConstDecl struct {
	Name  string
}

func (impl E_ConstTypeInvalid) ConstantError() {}
type E_ConstTypeInvalid struct {
	ConstName  string
	TypeError  *TypeError
}

func (impl E_ConstConflictWithType) ConstantError() {}
type E_ConstConflictWithType struct {
	Name  string
}


type ExprError struct {
	Point     ErrorPoint
	Concrete  ConcreteExprError
}

type ConcreteExprError interface { ExprError() }

func (impl E_InvalidInteger) ExprError() {}
type E_InvalidInteger struct {
	Value  string
}

func (impl E_ExprDuplicateField) ExprError() {}
type E_ExprDuplicateField struct {
	Name  string
}

func (impl E_HeterogeneousArray) ExprError() {}
type E_HeterogeneousArray struct {}

func (impl E_GetFromNonBundle) ExprError() {}
type E_GetFromNonBundle struct {}

func (impl E_GetFromLiteralBundle) ExprError() {}
type E_GetFromLiteralBundle struct {}

func (impl E_GetFromOpaqueBundle) ExprError() {}
type E_GetFromOpaqueBundle struct {}

func (impl E_SetToNonBundle) ExprError() {}
type E_SetToNonBundle struct {}

func (impl E_SetToLiteralBundle) ExprError() {}
type E_SetToLiteralBundle struct {}

func (impl E_SetToOpaqueBundle) ExprError() {}
type E_SetToOpaqueBundle struct {}

func (impl E_FieldDoesNotExist) ExprError() {}
type E_FieldDoesNotExist struct {
	Field   string
	Target  string
}

func (impl E_MissingField) ExprError() {}
type E_MissingField struct {
	Field  string
	Type   string
}

func (impl E_SurplusField) ExprError() {}
type E_SurplusField struct {
	Field  string
}

func (impl E_EntireValueIgnored) ExprError() {}
type E_EntireValueIgnored struct {}

func (impl E_TupleSizeNotMatching) ExprError() {}
type E_TupleSizeNotMatching struct {
	Required   int
	Given      int
	GivenType  string
}

func (impl E_NotAssignable) ExprError() {}
type E_NotAssignable struct {
	From    string
	To      string
	Reason  string
}

func (impl E_NotConstructable) ExprError() {}
type E_NotConstructable struct {
	From    string
	To      string
}

func (impl E_ExplicitTypeRequired) ExprError() {}
type E_ExplicitTypeRequired struct {}

func (impl E_DuplicateBinding) ExprError() {}
type E_DuplicateBinding struct {
	ValueName  string
}

func (impl E_MatchingNonTupleType) ExprError() {}
type E_MatchingNonTupleType struct {}

func (impl E_MatchingOpaqueTupleType) ExprError() {}
type E_MatchingOpaqueTupleType struct {}

func (impl E_MatchingNonBundleType) ExprError() {}
type E_MatchingNonBundleType struct {}

func (impl E_MatchingOpaqueBundleType) ExprError() {}
type E_MatchingOpaqueBundleType struct {}

func (impl E_LambdaAssignedToNonFuncType) ExprError() {}
type E_LambdaAssignedToNonFuncType struct {
	NonFuncType  string
}

func (impl E_IntegerAssignedToNonIntegerType) ExprError() {}
type E_IntegerAssignedToNonIntegerType struct {
	NonIntegerType  string
}

func (impl E_IntegerOverflow) ExprError() {}
type E_IntegerOverflow struct {
	Kind  string
}

func (impl E_TupleAssignedToNonTupleType) ExprError() {}
type E_TupleAssignedToNonTupleType struct {
	NonTupleType  string
}

func (impl E_BundleAssignedToNonBundleType) ExprError() {}
type E_BundleAssignedToNonBundleType struct {
	NonBundleType  string
}
