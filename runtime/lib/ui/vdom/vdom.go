package vdom

import (
	"reflect"
	"fmt"
)


type Map interface {
	Has(string) bool
	Lookup(string) (interface{}, bool)
	ForEach(func(string,interface{}))
}
type EmptyMap struct {}
func (_ EmptyMap) Has(string) bool { return false }
func (_ EmptyMap) Lookup(string) (interface{},bool) { return nil, false }
func (_ EmptyMap) ForEach(func(string,interface{})) {}

type Node struct {
	Tag  string
	Props
	Content
}

type Props struct {
	Attrs   *Attrs
	Styles  *Styles
	Events  *Events
}
type Attrs struct {
	// string -> string
	Data  Map
}
type Styles struct {
	// string -> string
	Data  Map
}
type Events struct {
	// string -> *EventHandler
	Data  Map
}

// TODO: change to functions to prevent them from being mutated
var EmptyAttrs = &Attrs { Data: EmptyMap {} }
var EmptyStyles = &Styles { Data: EmptyMap {} }
var EmptyEvents = &Events { Data: EmptyMap {} }
var EmptyContent = &Children {}

type Content interface { NodeContent() }
func (impl *Text) NodeContent() {}
type Text string
func (impl *Children) NodeContent() {}
type Children ([] *Node)
type EventHandler struct { Handler interface{} }

type DeltaNotifier struct {
	ApplyStyle   func(id string, key string, value string)
	EraseStyle   func(id string, key string)
	SetAttr      func(id string, key string, value string)
	RemoveAttr   func(id string, key string)
	AttachEvent  func(id string, name string, handler *EventHandler)
	ModifyEvent  func(id string, name string)
	DetachEvent  func(id string, name string, handler *EventHandler)
	SetText      func(id string, content string)
	AppendNode   func(parent string, id string, tag string)
	RemoveNode   func(parent string, id string)
	UpdateNode   func(old_id string, new_id string)
	ReplaceNode  func(parent string, old_id string, new_id string, tag string)
	SwapNode     func(parent string, a string, b string)
	MoveNode     func(parent string, id string, pivot string)
}

func assert(ok bool) {
	if !(ok) { panic("assertion failed") }
}

func get_addr(ptr interface{}) string {
	var n = reflect.ValueOf(ptr).Pointer()
	return string(fmt.Sprintf("%X", n))
}

func detach_all_events(ctx *DeltaNotifier, node *Node) {
	var node_id = get_addr(node)
	node.Events.Data.ForEach(func(name string, val interface{}) {
		var handler = val.(*EventHandler)
		ctx.DetachEvent(node_id, name, handler)
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
		if (old.Tag == new.Tag) {
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
			old_styles.Data.ForEach(func(key string, _ interface{}) {
				if !(new_styles.Data.Has(key)) {
					ctx.EraseStyle(id, key)
				}
			})
		}
		new_styles.Data.ForEach(func(key string, val_ interface{}) {
			var val = val_.(string)
			if old != nil {
				var old_val_, exists = old.Styles.Data.Lookup(key)
				if exists {
					if exists {
						var old_val = old_val_.(string)
						if old_val == val {
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
			old_attrs.Data.ForEach(func(name string, _ interface{}) {
				if !(new_attrs.Data.Has(name)) {
					ctx.RemoveAttr(id, name)
				}
			})
		}
		new_attrs.Data.ForEach(func(name string, val_ interface{}) {
			var val = val_.(string)
			if old != nil {
				var old_val_, exists = old.Attrs.Data.Lookup(name)
				if exists {
					var old_val = old_val_.(string)
					if old_val == val {
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
			old_events.Data.ForEach(func(key string, val interface{}) {
				var old_name = key
				var old_handler = val.(*EventHandler)
				var new_handler_, name_in_new = new_events.Data.Lookup(old_name)
				if name_in_new {
					var name = old_name
					var new_handler = new_handler_.(*EventHandler)
					if new_handler == old_handler {
						goto skip_event_opts
					}
					ctx.DetachEvent(id, name, old_handler)
					ctx.AttachEvent(id, name, new_handler)
					skip_event_opts:
				} else {
					ctx.DetachEvent(id, old_name, old_handler)
				}
			})
		}
		new_events.Data.ForEach(func(key string, val interface{}) {
			var new_name = key
			var new_handler = val.(*EventHandler)
			if !(old != nil && old.Events.Data.Has(new_name)) {
				 ctx.AttachEvent(id, new_name, new_handler)
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
				if is_text && (*old_content == *new_content) {
					goto skip_text
				}
			}
			ctx.SetText(id, string(*new_content))
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
					ctx.SetText(id, string(""))
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

