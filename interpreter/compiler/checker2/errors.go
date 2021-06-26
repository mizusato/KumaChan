package checker2

import (
	"fmt"
	"strconv"
	"kumachan/standalone/util/richtext"
)


func makeErrorDescBlankBlock() richtext.Block {
	var b richtext.Block
	b.WriteSpan("Error: ", richtext.TAG_EM)
	return b
}
func makeErrorDescBlock(msg string) richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan(msg, richtext.TAG_ERR_NORMAL)
	return b
}

type E_BlankTypeDefinition struct {}
func (E_BlankTypeDefinition) DescribeError() richtext.Block {
	return makeErrorDescBlock("blank type definition")
}

type E_DuplicateAlias struct {
	Which  string
}
func (e E_DuplicateAlias) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("duplicate alias:", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	return b
}

type E_InvalidAlias struct {
	Which  string
}
func (e E_InvalidAlias) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("invalid alias:", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	b.WriteSpan("(alias cannot point to another alias)", richtext.TAG_ERR_NOTE)
	return b
}

type E_DuplicateTypeDefinition struct {
	Which  string
}
func (e E_DuplicateTypeDefinition) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("duplicate definition for type", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	return b
}

type E_InvalidMetadata struct {
	Reason  string
}
func (e E_InvalidMetadata) DescribeError() richtext.Block {
	var msg = fmt.Sprintf("invalid metadata: %s", e.Reason)
	return makeErrorDescBlock(msg)
}

type E_InvalidTypeName struct {
	Name  string
}
func (e E_InvalidTypeName) DescribeError() richtext.Block {
	var msg = fmt.Sprintf("invalid type name: %s", strconv.Quote(e.Name))
	return makeErrorDescBlock(msg)
}

type E_TypeParametersOnCaseType struct {}
func (e E_TypeParametersOnCaseType) DescribeError() richtext.Block {
	var msg = "cannot specify explicit type parameters on case types"
	return makeErrorDescBlock(msg)
}

type E_TypeConflictWithAlias struct {
	Which  string
}
func (e E_TypeConflictWithAlias) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("type name", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	b.WriteSpan("conflicts with alias declaration", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	return b
}

type E_TypeNotFound struct {
	Which  string
}
func (e E_TypeNotFound) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("no such type:", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	return b
}

type E_TypeWrongParameterQuantity struct {
	Which  string
	Given  uint
	Least  uint
	Total  uint
}
func (e E_TypeWrongParameterQuantity) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("wrong parameter quantity for type", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	var arity string
	if e.Least != e.Total {
		arity = fmt.Sprintf("total %d [at least %d]", e.Total, e.Least)
	} else {
		arity = fmt.Sprintf("total %d", e.Total)
	}
	var arity_note = fmt.Sprintf("(%s required but %d given)", arity, e.Given)
	b.WriteSpan(arity_note, richtext.TAG_ERR_NOTE)
	return b
}

type E_TypeDuplicateField struct {
	Which  string
}
func (e E_TypeDuplicateField) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("duplicate field:", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	return b
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
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("duplicated implemented type:", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	return b
}

type E_BadImplemented struct {
	Which  string
}
func (e E_BadImplemented) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("bad implemented type:", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.Which, richtext.TAG_ERR_INLINE)
	b.WriteSpan("(should be an interface type)", richtext.TAG_ERR_NOTE)
	return b
}

type E_TooManyTypeParameters struct {
	TypeName  string
}
func (e E_TooManyTypeParameters) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("too many type parameters declared on type", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.TypeName, richtext.TAG_ERR_INLINE)
	return b
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

type E_ImplementedIncompatibleParameters struct {
	TypeName       string
	InterfaceName  string
}
func (e E_ImplementedIncompatibleParameters) DescribeError() richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan("type", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.TypeName, richtext.TAG_ERR_INLINE)
	b.WriteSpan("cannot implement the interface", richtext.TAG_ERR_NORMAL)
	b.WriteSpan(e.InterfaceName, richtext.TAG_ERR_INLINE)
	b.WriteSpan("(parameters incompatible)", richtext.TAG_ERR_NOTE)
	return b
}


