package checker2

import (
	"fmt"
	"strconv"
	"strings"
	"kumachan/standalone/util/richtext"
)


const BlockClassError = "error"
const BlockClassErrorContentItem = "error-content-item"

func makeErrorDescBlankBlock() richtext.Block {
	var b richtext.Block
	b.AddClass(BlockClassError)
	b.WriteSpan("Error: ", richtext.TAG_EM)
	return b
}
func makeErrorDescBlock(msg ...string) richtext.Block {
	var b = makeErrorDescBlankBlock()
	for i, span := range msg {
		b.WriteSpan(span, (func() string {
			if (i % 2) == 0 {
				if strings.HasPrefix(span, "(") && strings.HasSuffix(span, ")") {
					return richtext.TAG_ERR_NOTE
				} else {
					return richtext.TAG_ERR_NORMAL
				}
			} else {
				return richtext.TAG_ERR_INLINE
			}
		})())
	}
	return b
}
func makeEmptyErrorContentItemBlock() richtext.Block {
	var b richtext.Block
	b.AddClass(BlockClassErrorContentItem)
	return b
}

type ImplError struct {
	Concrete   string
	Interface  string
	Method     string
}
func (e ImplError) Describe(problem string) richtext.Block {
	return makeErrorDescBlock (
		"type", e.Concrete, "cannot implement interface", e.Interface,
		": method", e.Method, "not available:", problem,
	)
}

type SizeLimitError struct {
	Given  uint
	Limit  uint
}
func (e SizeLimitError) Describe(which string) richtext.Block {
	return makeErrorDescBlock (
		fmt.Sprintf("%s size limit exceeded (%d/%d)", which, e.Given, e.Limit),
	)
}

// ****************************************************************************

type E_ModuleNotFound struct {
	ModuleName  string
}
func (e E_ModuleNotFound) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"no such module: ", e.ModuleName,
	)
}

type E_BlankTypeDefinition struct {}
func (E_BlankTypeDefinition) DescribeError() richtext.Block {
	return makeErrorDescBlock("blank type definition")
}

type E_DuplicateAlias struct {
	Which  string
}
func (e E_DuplicateAlias) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"duplicate alias:", e.Which,
	)
}

type E_InvalidAlias struct {
	Which  string
}
func (e E_InvalidAlias) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"invalid alias:", e.Which,
		"(alias cannot point to another alias)",
	)
}

type E_DuplicateTypeDefinition struct {
	Which  string
}
func (e E_DuplicateTypeDefinition) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"duplicate definition for type", e.Which,
	)
}

type E_InvalidMetadata struct {
	Reason  string
}
func (e E_InvalidMetadata) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		fmt.Sprintf("invalid metadata: %s", e.Reason),
	)
}

type E_InvalidAliasName struct {
	Name  string
}
func (e E_InvalidAliasName) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		fmt.Sprintf("invalid alias name: %s", strconv.Quote(e.Name)),
	)
}

type E_InvalidTypeName struct {
	Name  string
}
func (e E_InvalidTypeName) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		fmt.Sprintf("invalid type name: %s", strconv.Quote(e.Name)),
	)
}

type E_TypeParametersOnCaseType struct {}
func (E_TypeParametersOnCaseType) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"cannot specify explicit type parameters on case types",
	)
}

type E_TypeConflictWithAlias struct {
	Which  string
}
func (e E_TypeConflictWithAlias) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"type name", e.Which, "conflicts with alias declaration", e.Which,
	)
}

type E_TypeNotFound struct {
	Which  string
}
func (e E_TypeNotFound) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"no such type:", e.Which,
	)
}

type E_TypeWrongParameterQuantity struct {
	Which  string
	Given  uint
	Least  uint
	Total  uint
}
func (e E_TypeWrongParameterQuantity) DescribeError() richtext.Block {
	var arity string
	if e.Least != e.Total {
		arity = fmt.Sprintf("total %d [at least %d]", e.Total, e.Least)
	} else {
		arity = fmt.Sprintf("total %d", e.Total)
	}
	var arity_note = fmt.Sprintf("(%s required but %d given)", arity, e.Given)
	return makeErrorDescBlock (
		"wrong parameter quantity for type", e.Which,
		arity_note,
	)
}

type E_TypeDuplicateField struct {
	Which  string
}
func (e E_TypeDuplicateField) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"duplicate field:", e.Which,
	)
}

type E_CircularSubtypingDefinition struct {
	Which  [] string
}
func (e E_CircularSubtypingDefinition) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("circular subtyping definition:", richtext.TAG_ERR_NORMAL)
	b.WriteSpan("cycle(s) detected within types:", richtext.TAG_ERR_NORMAL)
	for i, t := range e.Which {
		if i > 0 {
			b.WriteSpan(",", richtext.TAG_ERR_NORMAL)
		}
		b.WriteSpan(t, richtext.TAG_ERR_INLINE)
	}
	return b
}

type E_CircularInterfaceDefinition struct {
	Which  [] string
}
func (e E_CircularInterfaceDefinition) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("circular interface definition:", richtext.TAG_ERR_NORMAL)
	b.WriteSpan("cycle(s) detected within types:", richtext.TAG_ERR_NORMAL)
	for i, t := range e.Which {
		if i > 0 {
			b.WriteSpan(",", richtext.TAG_ERR_NORMAL)
		}
		b.WriteSpan(t, richtext.TAG_ERR_INLINE)
	}
	return b
}

type E_DuplicateImplemented struct {
	Which  string
}
func (e E_DuplicateImplemented) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"duplicated implemented type:", e.Which,
	)
}

type E_BadImplemented struct {
	Which  string
}
func (e E_BadImplemented) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"bad implemented type:", e.Which,
		"(should be an interface type)",
	)
}

type E_BlankTypeExpr struct {}
func (E_BlankTypeExpr) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"blank type expression",
	)
}

type E_TooManyTypeParameters struct {
	SizeLimitError
}
func (e E_TooManyTypeParameters) DescribeError() richtext.Block {
	return e.Describe("type parameter list")
}

type E_InvalidVarianceOnParameters struct {
	Which  [] string
}
func (e E_InvalidVarianceOnParameters) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("invalid variance declared on parameters:", richtext.TAG_ERR_NORMAL)
	for i, t := range e.Which {
		if i > 0 {
			b.WriteSpan(",", richtext.TAG_ERR_NORMAL)
		}
		b.WriteSpan(t, richtext.TAG_ERR_INLINE)
	}
	return b
}

type E_TooManyImplemented struct {
	SizeLimitError
}
func (e E_TooManyImplemented) DescribeError() richtext.Block {
	return e.Describe("implemented type list")
}

type E_ImplementedIncompatibleParameters struct {
	TypeName       string
	InterfaceName  string
}
func (e E_ImplementedIncompatibleParameters) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"type", e.TypeName,
		"cannot implement the interface", e.InterfaceName,
		"(parameters incompatible)",
	)
}

type E_InvalidFunctionName struct {
	Name  string
}
func (e E_InvalidFunctionName) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		fmt.Sprintf("invalid function name: %s", strconv.Quote(e.Name)),
	)
}

type E_FunctionConflictWithAlias struct {
	Which  string
}
func (e E_FunctionConflictWithAlias) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"function name", e.Which,
		"conflicts with alias declaration", e.Which,
	)
}

type E_ImplMethodNoSuchFunctionOrField struct {
	ImplError
}
func (e E_ImplMethodNoSuchFunctionOrField) DescribeError() richtext.Block {
	return e.Describe("no such function or field")
}

type E_ImplMethodNoneCompatible struct {
	ImplError
}
func (e E_ImplMethodNoneCompatible) DescribeError() richtext.Block {
	return e.Describe("none of method functions compatible")
}

type E_ImplMethodAmbiguous struct {
	ImplError
}
func (e E_ImplMethodAmbiguous) DescribeError() richtext.Block {
	return e.Describe(
		"ambiguous method implementation: " +
		"corresponding field and method function both exist",
	)
}

type E_ImplMethodDuplicateCompatible struct {
	ImplError
}
func (e E_ImplMethodDuplicateCompatible) DescribeError() richtext.Block {
	return e.Describe("multiple method functions compatible")
}

type E_ReceiverTypeNotFound struct {
	TypeName  string
}
func (e E_ReceiverTypeNotFound) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"bad receiver: no such type: ", e.TypeName,
	)
}

type E_IntegerNotRepresentableByFloatType struct {}
func (E_IntegerNotRepresentableByFloatType) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"integer literal is too big to be represented " +
		"using a floating-point type",
	)
}

type E_IntegerOverflowUnderflow struct {
	TypeName  string
}
func (e E_IntegerOverflowUnderflow) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"integer literal is not representable by the type", e.TypeName,
	)
}

type E_FloatOverflowUnderflow struct {}
func (E_FloatOverflowUnderflow) DescribeError() richtext.Block {
	return makeErrorDescBlock("float literal value too big")
}

type E_InvalidChar struct {
	Content  string
}
func (e E_InvalidChar) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"invalid character", e.Content,
	)
}

type E_NotAssignable struct {
	From  string
	To    string
}
func (e E_NotAssignable) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"type", e.From,
		"cannot be assigned to the type", e.To,
	)
}

type E_TupleAssignedToIncompatible struct {
	TypeName  string
}
func (e E_TupleAssignedToIncompatible) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"tuple literal cannot be assigned to incompatible type", e.TypeName,
	)
}

type E_TupleSizeNotMatching struct {
	Required  uint
	Given     uint
}
func (e E_TupleSizeNotMatching) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"tuple size not matching: " +
		fmt.Sprintf("%d required but %d given", e.Required, e.Given),
	)
}

type E_CannotMatchTuple struct {
	TypeName  string
}
func (e E_CannotMatchTuple) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"cannot match tuple from type", e.TypeName,
	)
}

type E_CannotMatchRecord struct {
	TypeName  string
}
func (e E_CannotMatchRecord) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"cannot match record from type", e.TypeName,
	)
}

type E_DuplicateBinding struct {
	BindingName  string
}
func (e E_DuplicateBinding) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"duplicate binding", e.BindingName,
	)
}

type E_FieldNotFound struct {
	FieldName  string
	TypeName   string
}
func (e E_FieldNotFound) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"field", e.FieldName,
		"does not exist in type", e.TypeName,
	)
}

type E_ExplicitTypeRequired struct {}
func (E_ExplicitTypeRequired) DescribeError() richtext.Block {
	return makeErrorDescBlock("explicit type cast expected")
}

type E_LambdaAssignedToIncompatible struct {
	TypeName  string
}
func (e E_LambdaAssignedToIncompatible) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"lambda cannot be assigned to incompatible type", e.TypeName,
	)
}

type E_TooManyTupleElements struct {
	SizeLimitError
}
func (e E_TooManyTupleElements) DescribeError() richtext.Block {
	return e.Describe("tuple")
}

type E_TooManyRecordFields struct {
	SizeLimitError
}
func (e E_TooManyRecordFields) DescribeError() richtext.Block {
	return e.Describe("record")
}

type E_TooManyEnumCases struct {
	SizeLimitError
}
func (e E_TooManyEnumCases) DescribeError() richtext.Block {
	return e.Describe("enum")
}

type E_DuplicateField struct {
	FieldName  string
}
func (e E_DuplicateField) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"duplicate field", e.FieldName,
	)
}

type E_UpdateOnNonRecord struct {
	TypeName  string
}
func (e E_UpdateOnNonRecord) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"cannot perform update operation on non-record type", e.TypeName,
	)
}

type E_RecordAssignedToIncompatible struct {
	TypeName  string
}
func (e E_RecordAssignedToIncompatible) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"cannot assign record literal to incompatible type", e.TypeName,
	)
}

type E_RecordSizeNotMatching struct {
	Given     uint
	Required  uint
}
func (e E_RecordSizeNotMatching) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"record size not matching: " +
		fmt.Sprintf("%d required but %d given", e.Required, e.Given),
	)
}

type E_TypeNotCallable struct {
	TypeName  string
}
func (e E_TypeNotCallable) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"type", e.TypeName,
		"is not callable",
	)
}

type E_NonLambdaRecursive struct {}
func (E_NonLambdaRecursive) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"non-lambda expression cannot be declared recursive",
	)
}

type E_NoSuchBindingOrFunction struct {
	Name  string
}
func (e E_NoSuchBindingOrFunction) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"no such binding or function: ", e.Name,
	)
}

type E_ImplicitContextNotFound struct {
	InnerError  richtext.Block
}
func (e E_ImplicitContextNotFound) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("implicit context not available: ", richtext.TAG_ERR_NORMAL)
	b.Append(e.InnerError)
	return b
}

type E_InvalidFunctionUsage struct {
	Candidates  [] OverloadCandidateDescription
}
type OverloadCandidateDescription struct {
	Signature  string
	Error      richtext.Block
}
func (e E_InvalidFunctionUsage) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteLine(
		"none of functions can be called/assigned:",
		richtext.TAG_ERR_NORMAL)
	for _, candidate := range e.Candidates {
		var item = makeEmptyErrorContentItemBlock()
		item.WriteLine(candidate.Signature, richtext.TAG_ERR_NOTE)
		item.Append(candidate.Error)
		b.Append(item)
	}
	return b
}

type E_AmbiguousFunctionUsage struct {
	Options  [] string
}
func (e E_AmbiguousFunctionUsage) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteLine(
		"callee/assignee not decidable within functions:",
		richtext.TAG_ERR_NORMAL)
	for _, option := range e.Options {
		var item = makeEmptyErrorContentItemBlock()
		item.WriteLine(option, richtext.TAG_ERR_NOTE)
		b.Append(item)
	}
	return b
}

type E_TypeParametersOnLocalBindingRef struct {}
func (E_TypeParametersOnLocalBindingRef) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"cannot specify type parameters on a local binding reference",
	)
}

type E_TypeParametersExceededArity struct {
	Arity  uint
}
func (e E_TypeParametersExceededArity) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		"the quantity of specified type parameters exceeded " +
		fmt.Sprintf("the arity (%d)", e.Arity),
	)
}

type E_InvalidTypeParameterOnFunction struct {
	Index  uint
}
func (e E_InvalidTypeParameterOnFunction) DescribeError() richtext.Block {
	return makeErrorDescBlock (
		fmt.Sprintf("a specified type parameter is invalid (#%d)", e.Index),
	)
}


