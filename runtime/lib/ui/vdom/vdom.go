package vdom

import (
	"reflect"
	"fmt"
)


type String = ([] rune)
type Map interface {
	Has(String) bool
	Lookup(String) (interface{}, bool)
	ForEach(func(String,interface{}))
}
type EmptyMap struct {}
func (_ EmptyMap) Has(String) bool { return false }
func (_ EmptyMap) Lookup(String) (interface{},bool) { return nil, false }
func (_ EmptyMap) ForEach(func(String,interface{})) {}
func str_equal(a String, b String) bool {
	if len(a) != len(b) { return false }
	var L = len(a)
	for i := 0; i < L; i += 1 {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func str_limited_equal(a String, b String) bool {
	const max_chars = 1024
	if len(a) != len(b) { return false }
	var L = len(a)
	if L > max_chars { return false }
	for i := 0; i < L; i += 1 {
		if a[i] != b[i] { return false }
	}
	return true
}

type Node struct {
	Tag  String
	Props
	Content
}

type Props struct {
	Attrs   *Attrs
	Styles  *Styles
	Events  *Events
}
type Attrs struct {
	Data  Map
}
type Styles struct {
	Data  Map
}
type Events struct {
	Data  Map
}
type EventOptions struct {
	Prevent  bool
	Stop     bool
	Capture  bool
	Handler  *EventHandler
}
func EventOptionsEqual(a *EventOptions, b *EventOptions) bool {
	if a == b {
		return true
	} else {
		return *a == *b
	}
}

// TODO: change to functions to prevent them from being mutated
var EmptyAttrs = &Attrs { Data: EmptyMap {} }
var EmptyStyles = &Styles { Data: EmptyMap {} }
var EmptyEvents = &Events { Data: EmptyMap {} }
var EmptyContent = &Children {}

type Content interface { NodeContent() }
func (impl *Text) NodeContent() {}
type Text String
func (impl *Children) NodeContent() {}
type Children ([] *Node)
type EventHandler struct { Handler interface{} }

type DeltaNotifier struct {
	ApplyStyle   func(id String, key String, value String)
	EraseStyle   func(id String, key String)
	SetAttr      func(id String, key String, value String)
	RemoveAttr   func(id String, key String)
	AttachEvent  func(id String, name String, prevent bool, stop bool, capture bool, handler *EventHandler)
	ModifyEvent  func(id String, name String, prevent bool, stop bool, capture bool)
	DetachEvent  func(id String, name String, handler *EventHandler)
	SetText      func(id String, content String)
	AppendNode   func(parent String, id String, tag String)
	RemoveNode   func(parent String, id String)
	UpdateNode   func(old_id String, new_id String)
	ReplaceNode  func(parent String, old_id String, new_id String, tag String)
	SwapNode     func(parent String, a String, b String)
	MoveNode     func(parent String, id String, pivot String)
}

func assert(ok bool) {
	if !(ok) { panic("assertion failed") }
}

func get_addr(ptr interface{}) String {
	var n = reflect.ValueOf(ptr).Pointer()
	return String(fmt.Sprintf("%X", n))
}

func detach_all_events(ctx *DeltaNotifier, node *Node) {
	var node_id = get_addr(node)
	node.Events.Data.ForEach(func(name String, val interface{}) {
		var opts = val.(*EventOptions)
		ctx.DetachEvent(node_id, name, opts.Handler)
	})
	var children, has_children = node.Content.(*Children)
	if has_children {
		for _, child := range *children {
			detach_all_events(ctx, child)
		}
	}
}

func Diff(ctx *DeltaNotifier, parent *Node, old *Node, new *Node) {
	assert(ctx != nil)
	assert(old != nil || new != nil)
	if old == new { return }
	var parent_id = get_addr(parent)
	var old_id = get_addr(old)
	var new_id = get_addr(new)
	if old == nil {
		ctx.AppendNode(parent_id, new_id, new.Tag)
	} else if new == nil {
		detach_all_events(ctx, old)
		ctx.RemoveNode(parent_id, old_id)
	} else {
		if str_equal(old.Tag, new.Tag) {
			ctx.UpdateNode(old_id, new_id)
		} else {
			ctx.ReplaceNode(parent_id, old_id, new_id, new.Tag)
			old = nil
		}
	}
	if new != nil {
		var id = new_id
		var node = new
		var new_styles = new.Styles
		if old != nil {
			var old_styles = old.Styles
			if new_styles == old_styles {
				goto skip_styles
			}
			old_styles.Data.ForEach(func(key String, _ interface{}) {
				if !(new_styles.Data.Has(key)) {
					ctx.EraseStyle(id, key)
				}
			})
		}
		new_styles.Data.ForEach(func(key String, val_ interface{}) {
			var val = val_.(String)
			if old != nil {
				var old_val_, exists = old.Styles.Data.Lookup(key)
				if exists {
					if exists {
						var old_val = old_val_.(String)
						if str_equal(old_val, val) {
							goto skip_this_style
						}
					}
				}
			}
			ctx.ApplyStyle(id, key, val)
			skip_this_style:
		})
		skip_styles:
		var new_attrs = new.Attrs
		if old != nil {
			var old_attrs = old.Attrs
			if new_attrs == old_attrs {
				goto skip_attrs
			}
			old_attrs.Data.ForEach(func(name String, _ interface{}) {
				if !(new_attrs.Data.Has(name)) {
					ctx.RemoveAttr(id, name)
				}
			})
		}
		new_attrs.Data.ForEach(func(name String, val_ interface{}) {
			var val = val_.(String)
			if old != nil {
				var old_val_, exists = old.Attrs.Data.Lookup(name)
				if exists {
					var old_val = old_val_.(String)
					if str_equal(old_val, val) {
						goto skip_this_attr
					}
				}
			}
			ctx.SetAttr(id, name, val)
			skip_this_attr:
		})
		skip_attrs:
		var new_events = new.Events
		if old != nil {
			var old_events = old.Events
			if old_events == new_events {
				goto skip_events
			}
			old_events.Data.ForEach(func(key String, val interface{}) {
				var old_name = key
				var old_opts = val.(*EventOptions)
				var new_opts_, name_in_new = new_events.Data.Lookup(old_name)
				if name_in_new {
					var name = old_name
					var new_opts = new_opts_.(*EventOptions)
					if EventOptionsEqual(new_opts, old_opts) {
						goto skip_event_opts
					}
					if new_opts.Handler == old_opts.Handler {
						ctx.ModifyEvent(id, name,
							new_opts.Prevent, new_opts.Stop, new_opts.Capture)
					} else {
						ctx.DetachEvent(id, name,
							old_opts.Handler)
						ctx.AttachEvent(id, name,
							new_opts.Prevent, new_opts.Stop, new_opts.Capture,
							new_opts.Handler)
					}
					skip_event_opts:
				} else {
					ctx.DetachEvent(id, old_name, old_opts.Handler)
				}
			})
		}
		new_events.Data.ForEach(func(key String, val interface{}) {
			var new_name = key
			var new_opts = val.(*EventOptions)
			if !(old != nil && old.Events.Data.Has(new_name)) {
				 ctx.AttachEvent(id, new_name,
				 	new_opts.Prevent, new_opts.Stop, new_opts.Capture,
				 	new_opts.Handler)
			}
		})
		skip_events:
		if old != nil && old.Content == new.Content {
			goto skip_content
		}
		switch new_content := new.Content.(type) {
		case *Text:
			if old != nil {
				var old_content, is_text = old.Content.(*Text)
				if is_text && str_limited_equal(*old_content, *new_content) {
					goto skip_text
				}
			}
			ctx.SetText(id, String(*new_content))
			skip_text:;
		case *Children:
			var new_children = *new_content
			var diff_children = func(old_children Children, new_children Children) {
				var build_index = func(nodes Children) (map[*Node] int, func(*Node) bool) {
					var index = make(map[*Node] int)
					for i, node := range nodes {
						index[node] = i
					}
					var has = func(node *Node) bool {
						var _, exists = index[node]
						return exists
					}
					return index, has
				}
				var old_index, old_has = build_index(old_children)
				var _, new_has = build_index(new_children)
				var old_skip = make(map[int] bool)
				var i = 0
				var j = 0
				for (i < len(old_children)) || (j < len(new_children)) {
					if old_skip[i] {
						i += 1
						continue
					}
					var old_child *Node = nil
					var new_child *Node = nil
					if i < len(old_children) {
						old_child = old_children[i]
					}
					if j < len(new_children) {
						new_child = new_children[j]
					}
					if new_child != old_child && old_has(new_child) {
						var node_id = get_addr(node)
						var old_child_id = get_addr(old_child)
						var new_child_id = get_addr(new_child)
						old_skip[old_index[new_child]] = true
						ctx.MoveNode(node_id, new_child_id, old_child_id)
						if !(new_has(old_child)) {
							Diff(ctx, node, old_child, nil) // remove
							i += 1
						}
						j += 1
						continue
					} else if new_child != old_child && new_has(old_child) {
						var node_id = get_addr(node)
						var old_child_id = get_addr(old_child)
						var new_child_id = get_addr(new_child)
						Diff(ctx, node, nil, new_child)  // append
						ctx.MoveNode(node_id, new_child_id, old_child_id)
						j += 1
						continue
					} else {
						Diff(ctx, node, old_child, new_child)
					}
					i += 1
					j += 1
				}
			}
			if old != nil {
				switch old_content := old.Content.(type) {
				case *Text:
					ctx.SetText(id, String(""))
					diff_children(Children([] *Node {}), new_children)
				case *Children:
					var old_children = *old_content
					diff_children(old_children, new_children)
				}
			} else {
				diff_children(Children([] *Node {}), new_children)
			}
		default:
			panic("impossible branch")
		}
		skip_content:
	}
}

