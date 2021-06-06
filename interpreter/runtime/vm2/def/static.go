package def

import (
	"fmt"
	"reflect"
)


type StaticValueSeed interface {
	Evaluate(ctx StaticValueSeedEvaluator) *Value
	fmt.Stringer
}
type StaticValueSeedEvaluator struct {
	GetFunctionReference  func(sym Symbol) *Value
}

type StaticValueSeedImmediate struct {
	ValuePointer  *Value
}
func (s StaticValueSeedImmediate) String() string {
	var v = *(s.ValuePointer)
	var t = reflect.TypeOf(v)
	return fmt.Sprintf("%s %v", t, v)
}
func (s StaticValueSeedImmediate) Evaluate(_ StaticValueSeedEvaluator) *Value {
	return s.ValuePointer
}

type StaticValueSeedFunctionReference struct {
	Symbol  Symbol
}
func (s StaticValueSeedFunctionReference) String() string {
	return fmt.Sprintf("ref func %s", s.Symbol)
}
func (s StaticValueSeedFunctionReference) Evaluate(ctx StaticValueSeedEvaluator) *Value {
	return ctx.GetFunctionReference(s.Symbol)
}


