package checker

import (
	"fmt"
	"kumachan/interpreter/def"
	. "kumachan/standalone/util/error"
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
	Name def.Symbol
}
func (impl E_WrongParameterQuantity) TypeError() {}
type E_WrongParameterQuantity struct {
	TypeName def.Symbol
	Required uint
	Given    uint
}
func (impl E_BoundNotSatisfied) TypeError() {}
type E_BoundNotSatisfied struct {
	Kind  TypeBoundKind
	Bound string
}
func (impl E_InvalidBoundType) TypeError() {}
type E_InvalidBoundType struct {
	Type  string
}
func (impl E_DuplicateField) TypeError() {}
type E_DuplicateField struct {
	FieldName  string
}
func (impl E_InvalidFieldName) TypeError() {}
type E_InvalidFieldName struct {
	Name  string
}
func (impl E_TooManyEnumItems) TypeError() {}
type E_TooManyEnumItems struct {
	Defined  uint
	Limit    uint
}
func (impl E_TooManyTupleRecordItems) TypeError() {}
type E_TooManyTupleRecordItems struct {
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
	EnumName  string
}
func (impl E_CaseBadBounds) TypeError() {}
type E_CaseBadBounds struct {
	CaseName   string
	EnumName  string
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
		msg.WriteInnerText(TS_INLINE_CODE, fmt.Sprintf("%c %s", e.Kind, e.Bound))
		msg.WriteText(TS_ERROR, "not satisfied")
	case E_InvalidBoundType:
		msg.WriteText(TS_ERROR, "Invalid bound type")
		msg.WriteInnerText(TS_INLINE_CODE, e.Type)
		msg.WriteText(TS_ERROR, "(parameter types cannot be used as bounds)")
	case E_DuplicateField:
		msg.WriteText(TS_ERROR, "Duplicate field:")
		msg.WriteEndText(TS_INLINE_CODE, e.FieldName)
	case E_InvalidFieldName:
		msg.WriteText(TS_ERROR, "Invalid field name:")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_TooManyEnumItems:
		msg.WriteText(TS_ERROR, "Too many enum items:")
		msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Defined))
		msg.WriteText(TS_ERROR,
			fmt.Sprintf("items (maximum is %d)", e.Limit))
	case E_TooManyTupleRecordItems:
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
			"Some parameter variance declaration(s) on the case type")
		msg.WriteInnerText(TS_INLINE_CODE, e.CaseName)
		msg.WriteText(TS_ERROR,
			"is not consistent with its parent enum type")
		msg.WriteEndText(TS_INLINE_CODE, e.EnumName)
	case E_CaseBadBounds:
		msg.WriteText(TS_ERROR,
			"Some parameter bound declaration(s) on the case type")
		msg.WriteInnerText(TS_INLINE_CODE, e.CaseName)
		msg.WriteText(TS_ERROR,
			"is not consistent with its parent enum type")
		msg.WriteEndText(TS_INLINE_CODE, e.EnumName)
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *TypeError) ErrorPoint() ErrorPoint {
	return err.Point
}

func (err *TypeError) ErrorConcrete() interface{} {
	return err.Concrete
}

func (err *TypeError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
}

func (err *TypeError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, [] ErrorMessage {
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
	TypeName def.Symbol
}
func (impl E_InvalidTypeTag) TypeDeclError() {}
type E_InvalidTypeTag struct {
	Tag   string
	Info  string
}
func (impl E_InvalidFieldTag) TypeDeclError() {}
type E_InvalidFieldTag struct {
	Tag   string
	Info  string
}
func (impl E_InvalidTypeTags) TypeDeclError() {}
type E_InvalidTypeTags struct {
	Info  string
}
func (impl E_InvalidCaseTypeParam) TypeDeclError() {}
type E_InvalidCaseTypeParam struct {
	Name  string
}
func (impl E_InvalidTypeDecl) TypeDeclError() {}
type E_InvalidTypeDecl struct {
	TypeName def.Symbol
	Detail   *TypeError
}
func (impl E_TypeCircularDependency) TypeDeclError() {}
type E_TypeCircularDependency struct {
	TypeNames  [] def.Symbol
}
func (impl E_TypeIncompleteDefaultParameters) TypeDeclError() {}
type E_TypeIncompleteDefaultParameters struct {}

func (err *TypeDeclError) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_InvalidTypeName:
		msg.WriteText(TS_ERROR, "Invalid type name:")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_DuplicateTypeDecl:
		msg.WriteText(TS_ERROR, "Duplicate type declaration:")
		msg.WriteEndText(TS_INLINE_CODE, e.TypeName.SymbolName)
	case E_InvalidTypeTag:
		msg.WriteText(TS_ERROR, "Invalid type tag:")
		msg.WriteInnerText(TS_INLINE_CODE, fmt.Sprintf("'%s'", e.Tag))
		msg.WriteText(TS_ERROR, fmt.Sprintf("(%s)", e.Info))
	case E_InvalidFieldTag:
		msg.WriteText(TS_ERROR, "Invalid field tag:")
		msg.WriteInnerText(TS_INLINE_CODE, fmt.Sprintf("'%s'", e.Tag))
		msg.WriteText(TS_ERROR, fmt.Sprintf("(%s)", e.Info))
	case E_InvalidTypeTags:
		msg.WriteText(TS_ERROR, "Invalid type tags:")
		msg.WriteEndText(TS_ERROR, e.Info)
	case E_InvalidCaseTypeParam:
		msg.WriteText(TS_ERROR, "Invalid type parameter")
		msg.WriteInnerText(TS_INLINE_CODE, e.Name)
		msg.WriteText(TS_ERROR, "(parameters of a case type should be a subset of its parent enum type)")
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
	case E_TypeIncompleteDefaultParameters:
		msg.WriteText(TS_ERROR, "Incomplete default type parameters. " +
			"If a parameter has a default value, then all of its following " +
			"parameters should also have default values.")
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *TypeDeclError) ErrorPoint() ErrorPoint {
	return err.Point
}

func (err *TypeDeclError) ErrorConcrete() interface{} {
	return err.Concrete
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


type KmdError struct {
	Point     ErrorPoint
	Concrete  ConcreteKmdError
}

type ConcreteKmdError interface { KmdError() }

func (impl E_KmdDuplicateType) KmdError() {}
type E_KmdDuplicateType struct {
	Id  string
}

func (impl E_KmdOnNative) KmdError() {}
type E_KmdOnNative struct {}

func (impl E_KmdFieldNotSerializable) KmdError() {}
type E_KmdFieldNotSerializable struct {
	FieldName  string
}

func (impl E_KmdElementNotSerializable) KmdError() {}
type E_KmdElementNotSerializable struct {
	ElementIndex  uint
}

func (impl E_KmdCaseNotSerializable) KmdError() {}
type E_KmdCaseNotSerializable struct {
	CaseName  string
}

func (impl E_KmdTypeNotSerializable) KmdError() {}
type E_KmdTypeNotSerializable struct {}

func (impl E_KmdDuplicateAdapter) KmdError() {}
type E_KmdDuplicateAdapter struct {}

func (impl E_KmdDuplicateValidator) KmdError() {}
type E_KmdDuplicateValidator struct {}

func (impl E_KmdValidatorNotInSameModule) KmdError() {}
type E_KmdValidatorNotInSameModule struct {}

func (impl E_KmdMissingValidator) KmdError() {}
type E_KmdMissingValidator struct {}

func (impl E_KmdSuspiciousValidator) KmdError() {}
type E_KmdSuspiciousValidator struct {}

func (err *KmdError) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_KmdDuplicateType:
		msg.WriteText(TS_ERROR, "duplicate type definition for KMD type id:")
		msg.WriteEndText(TS_INLINE, e.Id)
	case E_KmdOnNative:
		msg.WriteText(TS_ERROR, "KMD serialization on native types " +
			"is not supported")
	case E_KmdFieldNotSerializable:
		msg.WriteText(TS_ERROR, "Field")
		msg.WriteInnerText(TS_INLINE_CODE, e.FieldName)
		msg.WriteText(TS_ERROR, "is not KMD serializable")
	case E_KmdElementNotSerializable:
		msg.WriteText(TS_ERROR, "Element")
		msg.WriteInnerText(TS_INLINE, fmt.Sprintf("#%d", e.ElementIndex))
		msg.WriteText(TS_ERROR, "is not KMD serializable")
	case E_KmdCaseNotSerializable:
		msg.WriteText(TS_ERROR, "Case type")
		msg.WriteInnerText(TS_INLINE_CODE, e.CaseName)
		msg.WriteText(TS_ERROR, "is not KMD serializable")
	case E_KmdTypeNotSerializable:
		msg.WriteText(TS_ERROR, "This type is not KMD serializable")
	case E_KmdDuplicateAdapter:
		msg.WriteText(TS_ERROR, "Duplicate adapter")
	case E_KmdDuplicateValidator:
		msg.WriteText(TS_ERROR, "Duplicate validator")
	case E_KmdValidatorNotInSameModule:
		msg.WriteText(TS_ERROR,
			"Validator should be defined in the same module with its input type")
	case E_KmdMissingValidator:
		msg.WriteText(TS_ERROR, "Missing validator for this type")
	case E_KmdSuspiciousValidator:
		msg.WriteText(TS_ERROR, "Suspicious validator defined on this type")
		msg.WriteEndText(TS_INFO, "(maybe the type should be protected or opaque)")
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *KmdError) ErrorPoint() ErrorPoint {
	return err.Point
}

func (err *KmdError) ErrorConcrete() interface{} {
	return err.Concrete
}

func (err *KmdError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
}

func (err *KmdError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, [] ErrorMessage {
		err.Message(),
	})
	return msg.String()
}


type ServiceError struct {
	Point     ErrorPoint
	Concrete  ConcreteServiceError
}
type ConcreteServiceError interface { ServiceError() }

func (impl E_ServiceDuplicateModule) ServiceError() {}
type E_ServiceDuplicateModule struct {
	Id  string
}

func (impl E_ServiceMethodInvalidSignature) ServiceError() {}
type E_ServiceMethodInvalidSignature struct {
	Reason  string
}

func (err *ServiceError) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_ServiceDuplicateModule:
		msg.WriteText(TS_ERROR, "duplicate module for service:")
		msg.WriteEndText(TS_INLINE, e.Id)
	case E_ServiceMethodInvalidSignature:
		msg.WriteText(TS_ERROR, "invalid service method signature:")
		msg.WriteEndText(TS_ERROR, e.Reason)
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *ServiceError) ErrorPoint() ErrorPoint {
	return err.Point
}

func (err *ServiceError) ErrorConcrete() interface{} {
	return err.Concrete
}

func (err *ServiceError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
}

func (err *ServiceError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, [] ErrorMessage {
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

func (impl E_InvalidFunctionTag) FunctionError() {}
type E_InvalidFunctionTag struct {
	Tag   string
	Info  string
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

func (impl E_TooManyImplicitContextField) FunctionError() {}
type E_TooManyImplicitContextField struct {
	Defined  uint
	Limit    uint
}

func (impl E_ImplicitContextOnNativeFunction) FunctionError() {}
type E_ImplicitContextOnNativeFunction struct {}

func (impl E_ImplicitContextOnServiceMethod) FunctionError() {}
type E_ImplicitContextOnServiceMethod struct {}

func (impl E_NativeFunctionOutsideStandardLibrary) FunctionError() {}
type E_NativeFunctionOutsideStandardLibrary struct {}

func (impl E_MissingFunctionDefinition) FunctionError() {}
type E_MissingFunctionDefinition struct {
	FuncName  string
}

func (E_InvalidOverload) FunctionError() {}
type E_InvalidOverload struct {
	BetweenLocal  bool
	AddedName     string
	AddedModule   string
}

func (impl E_FunctionInvalidTypeParameterName) FunctionError() {}
type E_FunctionInvalidTypeParameterName struct {
	Name  string
}

func (impl E_FunctionVarianceDeclared) FunctionError() {}
type E_FunctionVarianceDeclared struct {}

func (impl E_FunctionDefaultTypeParameterDeclared) FunctionError() {}
type E_FunctionDefaultTypeParameterDeclared struct {}

func (err *FunctionError) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	switch e := err.Concrete.(type) {
	case E_InvalidFunctionName:
		msg.WriteText(TS_ERROR, "Invalid function name")
		msg.WriteEndText(TS_INLINE_CODE, e.Name)
	case E_InvalidFunctionTag:
		msg.WriteText(TS_ERROR, "Invalid function tag:")
		msg.WriteInnerText(TS_INLINE_CODE, fmt.Sprintf("'%s'", e.Tag))
		msg.WriteText(TS_ERROR, fmt.Sprintf("(%s)", e.Info))
	case E_InvalidTypeInFunction:
		msg.WriteAll(e.TypeError.Desc())
	case E_InvalidImplicitContextType:
		msg.WriteText(TS_ERROR, "Invalid implicit context type:")
		msg.WriteEndText(TS_ERROR, e.Reason)
	case E_ConflictImplicitContextField:
		msg.WriteText(TS_ERROR, "Conflict implicit context field:")
		msg.WriteEndText(TS_INLINE_CODE, e.FieldName)
	case E_TooManyImplicitContextField:
		msg.WriteText(TS_ERROR, "Too many implicit context:")
		msg.WriteInnerText(TS_INLINE, fmt.Sprint(e.Defined))
		msg.WriteText(TS_ERROR,
			fmt.Sprintf("fields (maximum is %d)", e.Limit))
	case E_ImplicitContextOnNativeFunction:
		msg.WriteText(TS_ERROR, "Cannot use implicit context on a native function")
	case E_ImplicitContextOnServiceMethod:
		msg.WriteText(TS_ERROR, "Cannot use implicit context on a service method")
	case E_NativeFunctionOutsideStandardLibrary:
		msg.WriteText(TS_ERROR, "Cannot define native function outside standard library")
	case E_MissingFunctionDefinition:
		msg.WriteText(TS_ERROR, "Missing function definition for")
		msg.WriteEndText(TS_INLINE_CODE, e.FuncName)
	case E_InvalidOverload:
		msg.WriteText(TS_ERROR, "Cannot define this function with the name")
		msg.WriteInnerText(TS_INLINE_CODE, e.AddedName)
		msg.WriteText(TS_ERROR, "since a function with identical signature exists")
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
	case E_FunctionDefaultTypeParameterDeclared:
		msg.WriteText(TS_ERROR,
			"Cannot declare default type parameter value on functions")
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *FunctionError) ErrorPoint() ErrorPoint {
	return err.Point
}

func (err *FunctionError) ErrorConcrete() interface{} {
	return err.Concrete
}

func (err *FunctionError) Message() ErrorMessage {
	return FormatErrorAt(err.Point, err.Desc())
}

func (err *FunctionError) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, [] ErrorMessage {
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

type E_GetFromNonRecord struct {}
func (e E_GetFromNonRecord) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform field access on a value of non-record type")
	return msg
}

type E_GetFromLiteralRecord struct {}
func (e E_GetFromLiteralRecord) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Suspicious field access on record literal")
	return msg
}

type E_GetFromOpaqueRecord struct {}
func (e E_GetFromOpaqueRecord) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform field access on a value of opaque record type")
	return msg
}

type E_SetToNonRecord struct {}
func (e E_SetToNonRecord) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform field update on a value of non-record type")
	return msg
}

type E_SetToLiteralRecord struct {}
func (e E_SetToLiteralRecord) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Suspicious field update on record literal")
	return msg
}

type E_SetToOpaqueRecord struct {}
func (e E_SetToOpaqueRecord) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform field update on a value of opaque record type")
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

type E_MatchingNonTupleType struct {}  // TODO: type description
func (e E_MatchingNonTupleType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform tuple destruction on a value of non-tuple type")
	return msg
}

type E_MatchingOpaqueTupleType struct {}  // TODO: type description
func (e E_MatchingOpaqueTupleType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform tuple destruction on a value of opaque tuple type")
	return msg
}

type E_MatchingNonRecordType struct {}  // TODO: type description
func (e E_MatchingNonRecordType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform record destruction on a value of non-record type")
	return msg
}

type E_MatchingOpaqueRecordType struct {}  // TODO: type description
func (e E_MatchingOpaqueRecordType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot perform record destruction on a value of opaque record type")
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

type E_IntegerNotRepresentableByFloatType struct {}
func (e E_IntegerNotRepresentableByFloatType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Integer literal is too big to be represented " +
		"by a floating-point type")
	return msg
}

type E_FloatOverflow struct {}
func (e E_FloatOverflow) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Float literal overflow")
	return msg
}

type E_IntegerOverflow struct {
	Kind  string
}
func (e E_IntegerOverflow) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Integer literal overflow")
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

type E_RecordAssignedToNonRecordType struct {
	NonRecordType  string
}
func (e E_RecordAssignedToNonRecordType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot assign record literal to the non-record type")
	msg.WriteEndText(TS_INLINE_CODE, e.NonRecordType)
	return msg
}

type E_ArrayAssignedToNonArrayType struct {
	NonArrayType  string
}
func (e E_ArrayAssignedToNonArrayType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Cannot assign list literal to the non-list type")
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

type E_SuperfluousTypeArgs struct {}
func (e E_SuperfluousTypeArgs) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Superfluous type arguments")
	return msg
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

type E_InvalidTypeForReactiveSwitch struct {
	Type  string
}
func (e E_InvalidTypeForReactiveSwitch) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot assign reactive switch to the type")
	msg.WriteInnerText(TS_INLINE_CODE, e.Type)
	msg.WriteText(TS_ERROR, "(a type which is not a multi-valued effect)")
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
	Enum     string
	TypeName  string
}
func (e E_NotBranchType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "The type")
	msg.WriteInnerText(TS_INLINE_CODE, e.TypeName)
	msg.WriteText(TS_ERROR, "is not a branch type of the enum type")
	msg.WriteEndText(TS_INLINE_CODE, e.Enum)
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
	Symbol def.Symbol
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
	msg.Write(T_LF)
	msg.WriteAllWithIndent(e.Detail.Desc(), 1)
	return msg
}

type E_BadTypeArg struct {
	Index   uint
	Name    string
	Detail  *TypeError
}
func (e E_BadTypeArg) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Bad type parameter")
	msg.WriteInnerText(TS_INLINE, fmt.Sprintf("(#%d %s)", e.Index, e.Name))
	msg.WriteText(TS_ERROR, ":")
	msg.Write(T_LF)
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
	TypeName def.Symbol
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

type E_AmbiguousFunctionAssign struct {
	Candidates  [] string
}
func (e E_AmbiguousFunctionAssign) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR,
		"Ambiguous function assignment")
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

type E_NoneOfFunctionsAssignable struct {
	To          string
	Candidates  [] UnavailableFuncInfo
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
		msg.WriteText(TS_INLINE_CODE, candidate.FuncDesc)
		msg.Write(T_LF)
		msg.WriteAllWithIndent(candidate.Error.Desc(), 2)
		msg.Write(T_LF)
	}
	return msg
}

type E_NoneOfFunctionsCallable struct {
	Candidates  [] UnavailableFuncInfo
}
type UnavailableFuncInfo struct {
	FuncDesc  string
	Error     *ExprError
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
		msg.WriteText(TS_INFO, candidate.FuncDesc)
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
	Enum  string
}
func (e E_NotCaseType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "The type")
	msg.WriteInnerText(TS_INLINE_CODE, e.Type)
	msg.WriteText(TS_ERROR, "is not a case type of the enum type")
	msg.WriteEndText(TS_INLINE_CODE, e.Enum)
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

type E_UnboxOpaqueType struct {
	Type  string
}
func (e E_UnboxOpaqueType) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot unbox a value from the opaque type")
	msg.WriteEndText(TS_INLINE_CODE, e.Type)
	return msg
}

type E_UnboxFailed struct {
	Type  string
}
func (e E_UnboxFailed) ExprErrorDesc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, "Cannot unbox a value from the type")
	msg.WriteEndText(TS_INLINE_CODE, e.Type)
	return msg
}

func (err *ExprError) ErrorPoint() ErrorPoint {
	return err.Point
}

func (err *ExprError) ErrorConcrete() interface{} {
	return err.Concrete
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

func (err* ExprError) GetInnerMost() *ExprError {
	// NOTE: there are some problems about the behavior of this function,
	//       but it is OK since it is only used for the ad-hoc plugin for Atom
	type item struct {
		error  *ExprError
		depth  int
	}
	var q = [] item { { error: err, depth: 0 } }
	var leaf_list = make([] item, 0)
	for len(q) > 0 {
		var i = q[0]
		q = q[1:]
		var e = i.error
		var d = i.depth
		switch e := e.Concrete.(type) {
		case E_NoneOfFunctionsCallable:
			var nest = false
			for _, candidate := range e.Candidates {
				switch candidate.Error.Concrete.(type) {
				case E_NoneOfFunctionsCallable:
					nest = true
					q = append(q, item {
						error: candidate.Error,
						depth: (d + 1),
					})
				}
			}
			if !(nest) {
				leaf_list = append(leaf_list, i)
			}
		default:
			return i.error
		}
	}
	if len(leaf_list) == 1 {
		return leaf_list[0].error
	} else if len(leaf_list) > 1 {
		var max = -1
		var max_index = -1
		for i, leaf := range leaf_list {
			if leaf.depth > max {
				max = leaf.depth
				max_index = i
			}
		}
		return leaf_list[max_index].error
	} else {
		panic("impossible branch")
	}
}

