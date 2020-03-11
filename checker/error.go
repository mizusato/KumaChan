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
func (impl E_InvalidFieldName) TypeError() {}
type E_InvalidFieldName struct {
	Name  string
}

func (err *TypeError) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_ModuleOfTypeRefNotFound:
		msg.WriteText(TS_ERROR, "No such module:")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_TypeNotFound:
		msg.WriteText(TS_ERROR, "No such type:")
		msg.WriteEndText(TS_INLINE_CODE, e.Name.String())
	case E_WrongParameterQuantity:
		msg.WriteText(TS_ERROR, "Wrong parameter quantity:")
		msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Required))
		msg.WriteText(TS_ERROR, "required but")
		msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Given))
		msg.WriteText(TS_ERROR, "given")
	case E_DuplicateField:
		msg.WriteText(TS_ERROR, "Duplicate field:")
		msg.WriteEndText(TS_INLINE_CODE, e.FieldName)
	case E_InvalidFieldName:
		msg.WriteText(TS_ERROR, "Invalid field name:")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *TypeError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc(), nil)
}

func (err *TypeError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, []ErrorMessage {
		err.Message(),
	})
	return msg.String()
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
	TypeName  loader.Symbol
	Detail    *TypeError
}

func (err *TypeDeclError) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_InvalidTypeName:
		msg.WriteText(TS_ERROR, "Invalid type name:")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_DuplicateTypeDecl:
		msg.WriteText(TS_ERROR, "Duplicate type declaration:")
		msg.WriteEndText(TS_INLINE_CODE, e.TypeName.SymbolName)
	case E_GenericUnionSubType:
		msg.WriteText(TS_ERROR, "Cannot define generic parameters on a union item")
	case E_InvalidTypeDecl:
		msg = e.Detail.Message()
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *TypeDeclError) Message() ErrorMessage {
	switch e := err.Concrete.(type) {
	case E_InvalidTypeDecl:
		return e.Detail.Message()
	default:
		return FormatErrorAt(err.Point, err.Desc(), nil)
	}
}

func (err *TypeDeclError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, []ErrorMessage {
		err.Message(),
	})
	return msg.String()
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

func (impl E_ArrayAssignedToNonArrayType) ExprError() {}
type E_ArrayAssignedToNonArrayType struct {
	NonArrayType  string
}

func (impl E_RecursiveMarkUsedOnNonLambda) ExprError() {}
type E_RecursiveMarkUsedOnNonLambda struct {}

func (impl E_TypeErrorInExpr) ExprError() {}
type E_TypeErrorInExpr struct {
	TypeError  *TypeError
}

func (impl E_InvalidMatchArgType) ExprError() {}
type E_InvalidMatchArgType struct {
	ArgType  string
}

func (impl E_DuplicateDefaultBranch) ExprError() {}
type E_DuplicateDefaultBranch struct {}

func (impl E_TypeParametersUnnecessary) ExprError() {}
type E_TypeParametersUnnecessary struct {}

func (impl E_NotSubtype) ExprError() {}
type E_NotSubtype struct {
	Union     string
	TypeName  string
}

func (impl E_IncompleteMatch) ExprError() {}
type E_IncompleteMatch struct {
	Missing  [] string
}

func (impl E_NonBooleanCondition) ExprError() {}
type E_NonBooleanCondition struct {
	Typed  bool
	Type   string
}

func (impl E_ModuleNotFound) ExprError() {}
type E_ModuleNotFound struct {
	Name  string
}

func (impl E_TypeOrValueNotFound) ExprError() {}
type E_TypeOrValueNotFound struct {
	Symbol  loader.Symbol
}

func (impl E_TypeParamInExpr) ExprError() {}
type E_TypeParamInExpr struct {
	Name  string
}

func (impl E_ExplicitTypeParamsRequired) ExprError() {}
type E_ExplicitTypeParamsRequired struct {}

func (impl E_TypeUsedAsValue) ExprError() {}
type E_TypeUsedAsValue struct {
	TypeName  loader.Symbol
}

func (impl E_FunctionWrongTypeParamsQuantity) ExprError() {}
type E_FunctionWrongTypeParamsQuantity struct {
	FuncName  string
	Given     uint
	Required  uint
}

func (impl E_NoneOfFunctionsAssignable) ExprError() {}
type E_NoneOfFunctionsAssignable struct {
	To          string
	Candidates  [] string
}

func (impl E_NoneOfFunctionsCallable) ExprError() {}
type E_NoneOfFunctionsCallable struct {
	Candidates  [] string
}

func (impl E_ExprNotCallable) ExprError() {}
type E_ExprNotCallable struct {}

func (impl E_ExprTypeNotCallable) ExprError() {}
type E_ExprTypeNotCallable struct {
	Type  string
}

func (impl E_NoneOfTypesAssignable) ExprError() {}
type E_NoneOfTypesAssignable struct {
	Types  [] string
}

func (impl E_BoxNonBoxedType) ExprError() {}
type E_BoxNonBoxedType struct {
	Type  string
}

func (impl E_BoxProtectedType) ExprError() {}
type E_BoxProtectedType struct {
	Type  string
}

func (impl E_BoxOpaqueType) ExprError() {}
type E_BoxOpaqueType struct {
	Type  string
}
