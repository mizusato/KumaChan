package ui

import (
	"fmt"
	"kumachan/lang"
	"kumachan/runtime/lib/ui/qt"
)


type EvaluatedObjectGroup struct {
	Name     string
	Widgets  map[string] qt.Widget
	Actions  map[string] qt.Action
}
var evaluatedObjectGroups = make(map[*lang.UiObjectGroup] EvaluatedObjectGroup)
func (g EvaluatedObjectGroup) GetObjectValue(name string) lang.Value {
	var widget, is_widget = g.Widgets[name]
	if is_widget {
		return widget
	}
	var action, is_action = g.Actions[name]
	if is_action {
		return action
	}
	panic(fmt.Sprintf("ui object '%s' not found in UI definition '%s'",
		name, g.Name))
}

func EvaluateObjectThunk(thunk lang.UiObjectThunk) lang.Value {
	var evaluated, exists = evaluatedObjectGroups[thunk.Group]
	if !(exists) {
		var group = thunk.Group
		var wait = make(chan struct{})
		qt.MakeSureInitialized()
		qt.CommitTask(func() {
			var root, ok = qt.LoadWidget(string(group.XmlDef), group.BaseDir)
			if !(ok) {
				panic(fmt.Sprintf("bad UI definition '%s'", group.GroupName))
			}
			var widgets = make(map[string] qt.Widget)
			var actions = make(map[string] qt.Action)
			widgets[group.RootName] = root
			for _, name := range group.Widgets {
				var widget, ok = qt.FindChild(root, name)
				if !(ok) {
					panic("something went wrong")
				}
				widgets[name] = widget
			}
			for _, name := range group.Actions {
				var action, ok = qt.FindChildAction(root, name)
				if !(ok) {
					panic("something went wrong")
				}
				actions[name] = action
			}
			evaluated = EvaluatedObjectGroup {
				Name:    group.GroupName,
				Widgets: widgets,
				Actions: actions,
			}
			evaluatedObjectGroups[group] = evaluated
			wait <- struct{}{}
		})
		<- wait
	}
	var stored = evaluated.GetObjectValue(thunk.Object)
	return lang.NativeFunctionValue(func(_ lang.Value, _ lang.InteropContext) lang.Value {
		return stored
	})
}

