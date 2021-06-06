package ui

import (
	"kumachan/interpreter/runtime/lib/container"
	"kumachan/interpreter/runtime/lib/ui/vdom"
	. "kumachan/interpreter/runtime/def"
	"strings"
)


type VdomMap struct {
	Data  container.Map
}
func VdomAdaptMap(m container.Map) VdomMap {
	return VdomMap { m }
}
func (m VdomMap) Has(key string) bool {
	var _, ok = m.Data.Lookup(key)
	return ok
}
func (m VdomMap) Lookup(key string) (interface{}, bool) {
	var v, ok = m.Data.Lookup(key)
	return v, ok
}
func (m VdomMap) ForEach(f func(key string, val interface{})) {
	m.Data.ForEach(func(k Value, v Value) {
		f(k.(string), v)
	})
}

func VdomMergeStyles(list container.List) *vdom.Styles {
	var styles = container.NewMapOfStringKey()
	list.ForEach(func(_ uint, part_ Value) {
		var part = part_.(*vdom.Styles)
		part.Data.ForEach(func(k string, v_obj interface{}) {
			var v = v_obj.(string)
			styles, _ = styles.Inserted(k, v)
		})
	})
	return &vdom.Styles { Data: VdomAdaptMap(styles) }
}
func VdomMergeAttrs(list container.List) *vdom.Attrs {
	const class = "class"
	var attrs = container.NewMapOfStringKey()
	list.ForEach(func(_ uint, part_ Value) {
		var part = part_.(*vdom.Attrs)
		part.Data.ForEach(func(k string, v_obj interface{}) {
			var v = v_obj.(string)
			if container.StringCompare(k, class) == Equal {
				var existing_, exists = attrs.Lookup(k)
				if exists {
					// merge class list
					var existing = existing_.(string)
					var buf strings.Builder
					buf.WriteString(existing)
					buf.WriteRune(' ')
					buf.WriteString(v)
					var v_merged_str = buf.String()
					attrs, _ = attrs.Inserted(k, v_merged_str)
				} else {
					attrs, _ = attrs.Inserted(k, v)
				}
			} else {
				attrs, _ = attrs.Inserted(k, v)
			}
		})
	})
	return &vdom.Attrs { Data: VdomAdaptMap(attrs) }
}
func VdomMergeEvents(list container.List) *vdom.Events {
	var events = container.NewMapOfStringKey()
	list.ForEach(func(_ uint, part_ Value) {
		var part = part_.(*vdom.Events)
		part.Data.ForEach(func(k string, v interface{}) {
			events, _ = events.Inserted(k, v)
		})
	})
	return &vdom.Events { Data: VdomAdaptMap(events) }
}

