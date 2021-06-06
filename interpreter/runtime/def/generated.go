package def

import (
	"fmt"
	"strconv"
)


type GeneratedNativeFunctionSeed interface {
	Evaluate(ctx GeneratedNativeFunctionSeedEvaluator) Value
	fmt.Stringer
}
type GeneratedNativeFunctionSeedEvaluator struct {
	EvaluateUiObjectSeed  func(seed *UiObjectSeed) Value
}

type UiObjectSeed struct {
	Object  string
	Group   *UiObjectGroup
}
type UiObjectGroup struct {
	GroupName  string
	BaseDir    string
	XmlDef     string
	RootName   string
	Widgets    [] string
	Actions    [] string
}
func (seed *UiObjectSeed) String() string {
	return fmt.Sprintf("%s %s %s",
		strconv.Quote(seed.Object),
		strconv.Quote(seed.Group.GroupName),
		strconv.Quote(seed.Group.BaseDir))
}
func (seed *UiObjectSeed) Evaluate(ctx GeneratedNativeFunctionSeedEvaluator) Value {
	return ctx.EvaluateUiObjectSeed(seed)
}


