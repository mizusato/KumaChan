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
func (impl E_BoundNotSatisfied) TypeError() {}
type E_BoundNotSatisfied struct {
	Kind   string
	Bound  string
}
func (impl E_DuplicateField) TypeError() {}
type E_DuplicateField struct {
	FieldName  string
}
func (impl E_InvalidFieldName) TypeError() {}
type E_InvalidFieldName struct {
	Name  string
}
func (impl E_TooManyUnionItems) TypeError() {}
type E_TooManyUnionItems struct {
	Defined  uint
	Limit    uint
}
func (impl E_TooManyTupleBundleItems) TypeError() {}
type E_TooManyTupleBundleItems struct {
	Defined  uint
	Limit    uint
}
func (impl E_BoxedBadVariance) TypeError() {}
type E_BoxedBadVariance struct {
	BadParams  [] string
}
func (impl E_CaseBadVariance) TypeError() {}
type E_CaseBadVariance struct {
	CaseName   string
	UnionName  string
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
	case E_BoundNotSatisfied:
		msg.WriteText(TS_ERROR, "Type parameter bound")
		msg.WriteInnerText(TS_INLINE_CODE, e.Kind + " " + e.Bound)
		msg.WriteText(TS_ERROR, "not satisfied")
	case E_DuplicateField:
		msg.WriteText(TS_ERROR, "Duplicate field:")
		msg.WriteEndText(TS_INLINE_CODE, e.FieldName)
	case E_InvalidFieldName:
		msg.WriteText(TS_ERROR, "Invalid field name:")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_TooManyUnionItems:
		msg.WriteText(TS_ERROR, "Too many union items:")
		msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Defined))
		msg.WriteText(TS_ERROR,
			fmt.Sprintf("items (maximum is %d)", e.Limit))
	case E_TooManyTupleBundleItems:
		msg.WriteText(TS_ERROR, "Too many elements/fields:")
		msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Defined))
		msg.WriteText(TS_ERROR,
			fmt.Sprintf("elements/fields (maximum is %d)", e.Limit))
	case E_BoxedBadVariance:
		msg.WriteText(TS_ERROR,
			"Contradictory variance declaration on type parameter(s):")
		msg.Write(T_SPACE)
		for i, p := range e.BadParams {
			msg.WriteText(TS_INLINE_CODE, p)
			if i != (len(e.BadParams) - 1) {
				msg.WriteText(TS_ERROR, ", ")
			}
		}
	case E_CaseBadVariance:
		msg.WriteText(TS_ERROR,
			"The declared variance of the case type")
		msg.WriteInnerText(TS_INLINE_CODE, e.CaseName)
		msg.WriteText(TS_ERROR,
			"is not consistent with its parent union type")
		msg.WriteEndText(TS_INLINE_CODE, e.UnionName)
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *TypeError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
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
func (impl E_InvalidCaseTypeParam) TypeDeclError() {}
type E_InvalidCaseTypeParam struct {
	Name  string
}
func (impl E_InvalidTypeDecl) TypeDeclError() {}
type E_InvalidTypeDecl struct {
	TypeName  loader.Symbol
	Detail    *TypeError
}
func (impl E_TypeCircularDependency) TypeDeclError() {}
type E_TypeCircularDependency struct {
	TypeNames  [] loader.Symbol
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
	case E_InvalidCaseTypeParam:
		msg.WriteText(TS_ERROR, "Invalid type parameter")
		msg.WriteInnerText(TS_INLINE_CODE, e.Name)
		msg.WriteText(TS_ERROR, "(parameters of a case type should be a subset of its parent union type)")
	case E_InvalidTypeDecl:
		msg.WriteAll(e.Detail.Desc())
	case E_TypeCircularDependency:
		msg.WriteText(TS_ERROR, "Dependency cycle found among types:")
		msg.Write(T_SPACE)
		for i, t := range e.TypeNames {
			msg.WriteText(TS_INLINE_CODE, t.String())
			if i != (len(e.TypeNames) - 1) {
				msg.WriteText(TS_ERROR, ", ")
			}
		}
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *TypeDeclError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
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

func (impl E_InvalidTypeInFunction) FunctionError() {}
type E_InvalidTypeInFunction struct {
 	TypeError  *TypeError
}

func (impl E_InvalidImplicitContextType) FunctionError() {}
type E_InvalidImplicitContextType struct {
	Reason  string
}

func (impl E_ConflictImplicitContextField) FunctionError() {}
type E_ConflictImplicitContextField struct {
	FieldName  string
}

func (impl E_ImplicitContextOnNativeFunction) FunctionError() {}
type E_ImplicitContextOnNativeFunction struct {}

func (E_SignatureNonLocal) FunctionError() {}
type E_SignatureNonLocal struct {
	FuncName  string
}

func (E_InvalidOverload) FunctionError() {}
type E_InvalidOverload struct {
	BetweenLocal  bool
	AddedName     string
	AddedModule   string
	AddedType     string
	ExistingType  string
}

func (impl E_FunctionInvalidTypeParameterName) FunctionError() {}
type E_FunctionInvalidTypeParameterName struct {
	Name  string
}

func (impl E_FunctionVarianceDeclared) FunctionError() {}
type E_FunctionVarianceDeclared struct {}

func (err *FunctionError) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_InvalidFunctionName:
		msg.WriteText(TS_ERROR, "Invalid function name")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_InvalidTypeInFunction:
		msg.WriteAll(e.TypeError.Desc())
	case E_InvalidImplicitContextType:
		msg.WriteText(TS_ERROR, "Invalid implicit context type:")
		msg.WriteEndText(TS_ERROR, e.Reason)
	case E_ConflictImplicitContextField:
		msg.WriteText(TS_ERROR, "Conflict implicit context field:")
		msg.WriteEndText(TS_INLINE_CODE, e.FieldName)
	case E_ImplicitContextOnNativeFunction:
		msg.WriteText(TS_ERROR, "Cannot use implicit context on a native function")
	case E_SignatureNonLocal:
		msg.WriteText(TS_ERROR, "Function")
		msg.WriteInnerText(TS_INLINE_CODE, e.FuncName)
		msg.WriteText(TS_ERROR, "is declared to be public but has a non-local signature type")
	case E_InvalidOverload:
		msg.WriteText(TS_ERROR, "Cannot overload this function instance with the signature")
		msg.WriteInnerText(TS_INLINE_CODE, e.AddedType)
		msg.WriteText(TS_ERROR, "on the function name")
		msg.WriteInnerText(TS_INLINE_CODE, e.AddedName)
		msg.WriteText(TS_ERROR, "since a function with conflicting signature")
		msg.WriteInnerText(TS_INLINE_CODE, e.ExistingType)
		msg.WriteText(TS_ERROR, "already exists")
		if e.BetweenLocal {
			msg.WriteEndText(TS_ERROR, "in the current module")
		} else {
			msg.WriteInnerText(TS_ERROR, "in the module")
			msg.WriteText(TS_INLINE_CODE, e.AddedModule)
		}
	case E_FunctionInvalidTypeParameterName:
		msg.WriteText(TS_ERROR, "invalid type name")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_FunctionVarianceDeclared:
		msg.WriteText(TS_ERROR,
			"Cannot declare variance of type parameters of functions")
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *FunctionError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
}

func (err *FunctionError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, []ErrorMessage {
		err.Message(),
	})
	return msg.String()
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

func (err *ConstantError) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_InvalidConstName:
		msg.WriteText(TS_ERROR, "Invalid constant name")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_DuplicateConstDecl:
		msg.WriteText(TS_ERROR, "Duplicate declaration of constant")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_ConstTypeInvalid:
		msg.WriteAll(e.TypeError.Desc())
	case E_ConstConflictWithType:
		msg.WriteText(TS_ERROR, "The constant name")
		msg.WriteInnerText(TS_INLINE_CODE, e.Name)
		msg.WriteText(TS_ERROR, "conflict with existing type name")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *ConstantError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
}

func (err *ConstantError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, []ErrorMessage {
		err.Message(),
	})
	return msg.String()
}


type ExprError struct {
	Point     ErrorPoint
	Concrete  ConcreteExprError
}

type ConcreteExprError interface {
	ExprErrorDesc() ErrorMessage
}

type E_InvalidInteger struct {
	Value  string
}
func (e E_InvalidInteger) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Invalid integer literal")
	msg.WriteEndText(TS_INLINE_CODE, e.Value)
	return msg
}

type E_ExprDuplicateField struct {
	Name  string
}
func (e E_ExprDuplicateField) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Duplicate field")
	msg.WriteEndText(TS_INLINE_CODE, e.Name)
	return msg
}

type E_GetFromNonBundle struct {}
func (e E_GetFromNonBundle) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform field access on a value of non-bundle type")
	return msg
}

type E_GetFromLiteralBundle struct {}
func (e E_GetFromLiteralBundle) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Suspicious field access on bundle literal")
	return msg
}

type E_GetFromOpaqueBundle struct {}
func (e E_GetFromOpaqueBundle) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform field access on a value of opaque bundle type")
	return msg
}

type E_SetToNonBundle struct {}
func (e E_SetToNonBundle) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform field update on a value of non-bundle type")
	return msg
}

type E_SetToLiteralBundle struct {}
func (e E_SetToLiteralBundle) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Suspicious field update on bundle literal")
	return msg
}

type E_SetToOpaqueBundle struct {}
func (e E_SetToOpaqueBundle) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform field update on a value of opaque bundle type")
	return msg
}

type E_FieldDoesNotExist struct {
	Field   string
	Target  string
}
func (e E_FieldDoesNotExist) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "The field")
	msg.WriteInnerText(TS_INLINE_CODE, e.Field)
	msg.WriteText(TS_ERROR, "does not exist on the type")
	msg.WriteEndText(TS_INLINE_CODE, e.Target)
	return msg
}

type E_MissingField struct {
	Field  string
	Type   string
}
func (e E_MissingField) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Missing the field")
	msg.WriteInnerText(TS_INLINE_CODE, e.Field)
	msg.WriteText(TS_ERROR, "(type: ")
	msg.WriteText(TS_INLINE_CODE, e.Type)
	msg.WriteText(TS_ERROR, ")")
	return msg
}

type E_SuperfluousField struct {
	Field  string
}
func (e E_SuperfluousField) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Superfluous field")
	msg.WriteEndText(TS_INLINE_CODE, e.Field)
	return msg
}

type E_EntireValueIgnored struct {}
func (e E_EntireValueIgnored) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Entire value ignored suspiciously")
	return msg
}

type E_TupleSizeNotMatching struct {
	Required   int
	Given      int
	GivenType  string
}
func (e E_TupleSizeNotMatching) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Tuple size not matching:")
	msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Required))
	msg.WriteText(TS_ERROR, "required but")
	msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Given))
	msg.WriteText(TS_ERROR, "given (expected type: ")
	msg.WriteText(TS_INLINE_CODE, e.GivenType)
	msg.WriteText(TS_ERROR, ")")
	return msg
}

type E_NotAssignable struct {
	From    string
	To      string
	Reason  string
}
func (e E_NotAssignable) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "The value of type")
	msg.WriteInnerText(TS_INLINE_CODE, e.From)
	msg.WriteText(TS_ERROR, "cannot be assigned to the type")
	msg.WriteEndText(TS_INLINE_CODE, e.To)
	if e.Reason != "" {
		msg.WriteText(TS_ERROR, " (")
		msg.WriteText(TS_ERROR, e.Reason)
		msg.WriteText(TS_ERROR, ")")
	}
	return msg
}

type E_ExplicitTypeRequired struct {}
func (e E_ExplicitTypeRequired) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Explicit type cast desired")
	return msg
}

type E_DuplicateBinding struct {
	ValueName  string
}
func (e E_DuplicateBinding) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Duplicate binding of value name")
	msg.WriteEndText(TS_INLINE_CODE, e.ValueName)
	return msg
}

type E_MatchingNonTupleType struct {}
func (e E_MatchingNonTupleType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform tuple destruction on a value of non-tuple type")
	return msg
}

type E_MatchingOpaqueTupleType struct {}
func (e E_MatchingOpaqueTupleType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform tuple destruction on a value of opaque tuple type")
	return msg
}

type E_MatchingNonBundleType struct {}
func (e E_MatchingNonBundleType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform tuple destruction on a value of non-bundle type")
	return msg
}

type E_MatchingOpaqueBundleType struct {}
func (e E_MatchingOpaqueBundleType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform tuple destruction on a value of opaque bundle type")
	return msg
}

type E_LambdaAssignedToNonFuncType struct {
	NonFuncType  string
}
func (e E_LambdaAssignedToNonFuncType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot assign lambda to the non-function type")
	msg.WriteEndText(TS_INLINE_CODE, e.NonFuncType)
	return msg
}

type E_IntegerAssignedToNonIntegerType struct {
	NonIntegerType  string
}
func (e E_IntegerAssignedToNonIntegerType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot assign integer literal to the non-integer type")
	msg.WriteEndText(TS_INLINE_CODE, e.NonIntegerType)
	return msg
}

type E_IntegerOverflow struct {
	Kind  string
}
func (e E_IntegerOverflow) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Integer literal overflows")
	msg.WriteEndText(TS_INLINE, e.Kind)
	return msg
}

type E_InvalidCharacter struct {
	RawValue  string
}
func (e E_InvalidCharacter) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Invalid character")
	msg.WriteEndText(TS_INLINE_CODE, e.RawValue)
	return msg
}

type E_TupleAssignedToNonTupleType struct {
	NonTupleType  string
}
func (e E_TupleAssignedToNonTupleType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot assign tuple literal to the non-tuple type")
	msg.WriteEndText(TS_INLINE_CODE, e.NonTupleType)
	return msg
}

type E_BundleAssignedToNonBundleType struct {
	NonBundleType  string
}
func (e E_BundleAssignedToNonBundleType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot assign bundle literal to the non-bundle type")
	msg.WriteEndText(TS_INLINE_CODE, e.NonBundleType)
	return msg
}

type E_ArrayAssignedToNonArrayType struct {
	NonArrayType  string
}
func (e E_ArrayAssignedToNonArrayType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot assign array literal to the non-array type")
	msg.WriteEndText(TS_INLINE_CODE, e.NonArrayType)
	return msg

}

type E_RecursiveMarkUsedOnNonLambda struct {}
func (e E_RecursiveMarkUsedOnNonLambda) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Invalid usage of recursion mark")
	return msg
}

type E_TypeErrorInExpr struct {
	TypeError  *TypeError
}
func (e E_TypeErrorInExpr) ExprErrorDesc() ErrorMessage {
	return e.TypeError.Desc()
}

type E_InvalidSwitchArgType struct {
	ArgType  string
}
func (e E_InvalidSwitchArgType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot pattern match on the value of type")
	msg.WriteEndText(TS_INLINE_CODE, e.ArgType)
	return msg
}

type E_CheckedBranch struct {}
func (e E_CheckedBranch) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "This situation already checked in another branch")
	return msg
}

type E_DuplicateDefaultBranch struct {}
func (e E_DuplicateDefaultBranch) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Duplicate default branch")
	return msg
}

type E_SuperfluousDefaultBranch struct {}
func (e E_SuperfluousDefaultBranch) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Superfluous default branch")
	return msg
}

type E_TypeParametersUnnecessary struct {}
func (e E_TypeParametersUnnecessary) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Unnecessary type parameters")
	return msg
}

type E_MultiBranchTypesAllDefault struct {}
func (e E_MultiBranchTypesAllDefault) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Invalid case types: should use default branch instead")
	return msg
}

type E_WrongMultiBranchTypeQuantity struct {
	Required  uint
	Given     uint
}
func (e E_WrongMultiBranchTypeQuantity) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "A case in this switch requires")
	msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Required))
	msg.WriteText(TS_ERROR, "case types but")
	msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Given))
	msg.WriteText(TS_ERROR, "given")
	return msg
}

type E_NotBranchType struct {
	Union     string
	TypeName  string
}
func (e E_NotBranchType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "The type")
	msg.WriteInnerText(TS_INLINE_CODE, e.TypeName)
	msg.WriteText(TS_ERROR, "is not a branch type of the union type")
	msg.WriteEndText(TS_INLINE_CODE, e.Union)
	return msg
}

type E_IncompleteMatch struct {
	Missing  [] string
}
func (e E_IncompleteMatch) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Pattern matching is not exhaustive: missing branches ")
	for i, branch := range e.Missing {
		msg.WriteText(TS_INLINE_CODE, branch)
		if i != len(e.Missing)-1 {
			msg.WriteText(TS_ERROR, ", ")
		}
	}
	return msg
}

type E_IncompleteMultiMatch struct {
	MissingQuantity  uint
}
func (e E_IncompleteMultiMatch) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Pattern matching is not exhaustive: missing")
	msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.MissingQuantity))
	msg.WriteText(TS_ERROR, "situations")
	return msg
}

type E_ModuleNotFound struct {
	Name  string
}
func (e E_ModuleNotFound) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "No such module:")
	msg.WriteEndText(TS_INLINE_CODE, e.Name)
	return msg
}

type E_TypeOrValueNotFound struct {
	Symbol  loader.Symbol
}
func (e E_TypeOrValueNotFound) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "No such value or type:")
	msg.WriteEndText(TS_INLINE_CODE, e.Symbol.String())
	return msg
}

type E_ImplicitContextNotFound struct {
	Name    string
	Detail  *ExprError
}
func (e E_ImplicitContextNotFound) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Implicit context value")
	msg.WriteInnerText(TS_INLINE_CODE, e.Name)
	msg.WriteText(TS_ERROR, "not found:")
	msg.WriteAllWithIndent(e.Detail.Desc(), 1)
	return msg
}

type E_TypeParamInExpr struct {
	Name  string
}
func (e E_TypeParamInExpr) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot use type parameter")
	msg.WriteInnerText(TS_INLINE_CODE, e.Name)
	msg.WriteText(TS_ERROR, "as a value")
	return msg
}

type E_ExplicitTypeParamsRequired struct {}
func (e E_ExplicitTypeParamsRequired) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Explicit type parameters expected")
	return msg
}

type E_TypeUsedAsValue struct {
	TypeName  loader.Symbol
}
func (e E_TypeUsedAsValue) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot use type")
	msg.WriteInnerText(TS_INLINE_CODE, e.TypeName.String())
	msg.WriteText(TS_ERROR, "as a value")
	return msg
}

type E_FunctionWrongTypeParamsQuantity struct {
	FuncName  string
	Given     uint
	Required  uint
}
func (e E_FunctionWrongTypeParamsQuantity) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "The function")
	msg.WriteInnerText(TS_INLINE_CODE, e.FuncName)
	msg.WriteText(TS_ERROR, "requires")
	msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Required))
	msg.WriteText(TS_ERROR, "type parameters but")
	msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Given))
	msg.WriteText(TS_ERROR, "given")
	return msg
}

type E_NoneOfFunctionsAssignable struct {
	To          string
	Candidates  [] Candidate
}
func (e E_NoneOfFunctionsAssignable) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"None of function instances assignable to the type")
	msg.WriteEndText(TS_INLINE_CODE, e.To)
	msg.Write(T_LF)
	msg.WriteText(TS_INFO, "*** candidates are:")
	msg.Write(T_LF)
	for _, candidate := range e.Candidates {
		msg.Write(T_INDENT)
		msg.WriteText(TS_INLINE_CODE, DescribeCandidate(candidate))
		msg.WriteAllWithIndent(candidate.Error.Desc(), 2)
		msg.Write(T_LF)
	}
	return msg
}

type E_NoneOfFunctionsCallable struct {
	Candidates  [] Candidate
}
func (e E_NoneOfFunctionsCallable) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"None of function instances can be called")
	msg.Write(T_LF)
	msg.WriteText(TS_INFO, "*** candidates are:")
	msg.Write(T_LF)
	var err_listed = make(map[string] int)
	for i, candidate := range e.Candidates {
		msg.Write(T_INDENT)
		msg.WriteText(TS_INFO, fmt.Sprintf("(%d)", (i + 1)))
		msg.Write(T_SPACE)
		msg.WriteText(TS_INFO, DescribeCandidate(candidate))
		msg.Write(T_LF)
		var candidate_msg = candidate.Error.Message()
		var key = candidate_msg.String()
		var listed_index, listed = err_listed[key]
		if listed {
			msg.WriteRepeated(T_INDENT, 2)
			msg.WriteText(TS_ERROR,
				fmt.Sprintf("= (%d)", (listed_index + 1)))
		} else {
			msg.WriteAllWithIndent(candidate_msg, 2)
			err_listed[key] = i
		}
		msg.Write(T_LF)
	}
	return msg
}

type E_AmbiguousCall struct {
	Candidates  [] string
}
func (e E_AmbiguousCall) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Ambiguous function call")
	msg.Write(T_LF)
	msg.WriteText(TS_INFO, "*** candidates are:")
	msg.Write(T_LF)
	for _, candidate := range e.Candidates {
		msg.Write(T_INDENT)
		msg.WriteText(TS_INLINE_CODE, candidate)
		msg.Write(T_LF)
	}
	return msg
}

type E_ExprNotCallable struct {}
func (e E_ExprNotCallable) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "This value is not callable")
	return msg
}

type E_ExprTypeNotCallable struct {
	Type  string
}
func (e E_ExprTypeNotCallable) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "This value is not callable: type")
	msg.WriteInnerText(TS_INLINE_CODE, e.Type)
	msg.WriteText(TS_ERROR, "is not callable")
	return msg
}

type E_NoneOfTypesAssignable struct {
	From  [] string
	To    string
}
func (e E_NoneOfTypesAssignable) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot assign to type")
	msg.WriteInnerText(TS_INLINE_CODE, e.To)
	msg.WriteText(TS_ERROR, "from any of available return types: ")
	for i, item := range e.From {
		msg.WriteText(TS_INLINE_CODE, item)
		if i != len(e.From)-1 {
			msg.WriteText(TS_ERROR, ", ")
		}
	}
	return msg
}

type E_NotCaseType struct {
	Type   string
	Union  string
}
func (e E_NotCaseType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "The type")
	msg.WriteInnerText(TS_INLINE_CODE, e.Type)
	msg.WriteText(TS_ERROR, "is not a case type of the union type")
	msg.WriteEndText(TS_INLINE_CODE, e.Union)
	return msg
}

type E_BoxNonBoxedType struct {
	Type  string
}
func (e E_BoxNonBoxedType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot box a value into the non-boxed type")
	msg.WriteEndText(TS_INLINE_CODE, e.Type)
	return msg
}

type E_BoxProtectedType struct {
	Type  string
}
func (e E_BoxProtectedType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot box a value into the protected type")
	msg.WriteEndText(TS_INLINE_CODE, e.Type)
	return msg
}

type E_BoxOpaqueType struct {
	Type  string
}
func (e E_BoxOpaqueType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot box a value into the opaque type")
	msg.WriteEndText(TS_INLINE_CODE, e.Type)
	return msg
}

func (err *ExprError) Desc() ErrorMessage {
	return err.Concrete.ExprErrorDesc()
}

func (err *ExprError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
}

func (err *ExprError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, []ErrorMessage {
		err.Message(),
	})
	return msg.String()
}
