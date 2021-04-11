package ui

import (
	"kumachan/runtime/lib/container"
	"kumachan/runtime/lib/ui/vdom"
	. "kumachan/lang"
)


type VdomMap struct {
	Data  container.Map
}
func VdomAdaptMap(m container.Map) VdomMap {
	return VdomMap { m }
}
func VdomAdaptMapValue(v Value) Value {
	var str, is_str = v.(String)
	if is_str {
		return RuneSliceFromString(str)
	} else {
		return v
	}
}
func (m VdomMap) Has(key vdom.String) bool {
	var key_str = StringFromRuneSlice(key)
	var _, ok = m.Data.Lookup(key_str)
	return ok
}
func (m VdomMap) Lookup(key vdom.String) (interface{}, bool) {
	var key_str = StringFromRuneSlice(key)
	var v, ok = m.Data.Lookup(key_str)
	return VdomAdaptMapValue(v), ok
}
func (m VdomMap) ForEach(f func(key vdom.String, val interface{})) {
	m.Data.ForEach(func(k Value, v Value) {
		f(RuneSliceFromString(k.(String)), VdomAdaptMapValue(v))
	})
}

func VdomMergeStyles(list container.List) *vdom.Styles {
	var styles = container.NewMapOfStringKey()
	list.ForEach(func(_ uint, part_ Value) {
		var part = part_.(*vdom.Styles)
		part.Data.ForEach(func(k_runes vdom.String, v_obj interface{}) {
			var v_runes = v_obj.([] rune)
			var k_str = StringFromRuneSlice(k_runes)
			var v_str = StringFromRuneSlice(v_runes)
			styles, _ = styles.Inserted(k_str, v_str)
		})
	})
	return &vdom.Styles { Data: VdomAdaptMap(styles) }
}
func VdomMergeAttrs(list container.List) *vdom.Attrs {
	var class = StringFromRuneSlice([] rune ("class"))
	var attrs = container.NewMapOfStringKey()
	list.ForEach(func(_ uint, part_ Value) {
		var part = part_.(*vdom.Attrs)
		part.Data.ForEach(func(k_runes vdom.String, v_obj interface{}) {
			var k_str = StringFromRuneSlice(k_runes)
			var v_runes = v_obj.([] rune)
			if container.StringCompare(k_str, class) == Equal {
				var existing, exists = attrs.Lookup(k_str)
				if exists {
					// merge class list
					var existing_runes = RuneSliceFromString(existing.(String))
					var buf = make([] rune, 0)
					buf = append(buf, existing_runes...)
					buf = append(buf, ' ')
					buf = append(buf, v_runes...)
					var v_merged_str = StringFromRuneSlice(buf)
					attrs, _ = attrs.Inserted(k_str, v_merged_str)
				} else {
					var v_str = StringFromRuneSlice(v_runes)
					attrs, _ = attrs.Inserted(k_str, v_str)
				}
			} else {
				var v_str = StringFromRuneSlice(v_runes)
				attrs, _ = attrs.Inserted(k_str, v_str)
			}
		})
	})
	return &vdom.Attrs { Data: VdomAdaptMap(attrs) }
}
func VdomMergeEvents(list container.List) *vdom.Events {
	var events = container.NewMapOfStringKey()
	list.ForEach(func(_ uint, part_ Value) {
		var part = part_.(*vdom.Events)
		part.Data.ForEach(func(k_ vdom.String, v interface{}) {
			var k = StringFromRuneSlice(k_)
			events, _ = events.Inserted(k, v)
		})
	})
	return &vdom.Events { Data: VdomAdaptMap(events) }
}

