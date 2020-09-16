package vdom

import (
	"reflect"
	"fmt"
)


type String = [] rune
type Map interface {
	Has(String) bool
	Lookup(String) (interface{}, bool)
	ForEach(func(String,interface{}))
}
func StringEqual(a String, b String) bool {
	if len(a) != len(b) { return false }
	var L = len(a)
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
	Styles  *Styles
	Events  *Events
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
	Handler  EventHandler
}
func EventOptionsEqual(a *EventOptions, b *EventOptions) bool {
	if a == b {
		return true
	} else {
		if a.Handler != b.Handler { return false }
		if a.Prevent != b.Prevent { return false }
		if a.Stop != b.Stop { return false }
		return true
	}
}

type Content interface { NodeContent() }
func (impl *Text) NodeContent() {}
type Text String
func (impl *Children) NodeContent() {}
type Children ([] *Node)
type EventHandler = interface{}

type DeltaNotifier struct {
	ApplyStyle   func(id String, key String, value String)
	EraseStyle   func(id String, key String)
	AttachEvent  func(id String, name String, prevent bool, stop bool, handler EventHandler)
	ModifyEvent  func(id String, name String, prevent bool, stop bool)
	DetachEvent  func(id String, name String, handler EventHandler)
	SetText      func(id String, content String)
	AppendNode   func(parent String, id String, tag String)
	// InsertNode   func(parent String, ref String, id String, tag String)
	RemoveNode   func(parent String, id String)
	UpdateNode   func(old_id String, new_id String)
	ReplaceNode  func(parent String, old_id String, new_id String, tag String)
}

func assert(ok bool) {
	if !(ok) { panic("assertion failed") }
}

func get_addr(ptr interface{}) String {
	var n = reflect.ValueOf(ptr).Pointer()
	return String(fmt.Sprintf("%X", n))
}

func str_equal(a String, b String) bool {
	if len(a) != len(b) { return false }
	var l = len(a)
	for i := 0; i < l; i += 1 {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
		old.Events.Data.ForEach(func(name String, val interface{}) {
			var opts = val.(*EventOptions)
			ctx.DetachEvent(old_id, name, opts.Handler)
		})
		ctx.RemoveNode(parent_id, old_id)
	} else {
		if str_equal(old.Tag, new.Tag) {
			ctx.UpdateNode(old_id, new_id)
		} else {
			ctx.ReplaceNode(parent_id, old_id, new_id, new.Tag)
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
			ctx.ApplyStyle(id, key, val)
		})
		skip_styles:
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
							new_opts.Prevent, new_opts.Stop)
					} else {
						ctx.DetachEvent(id, name,
							old_opts.Handler)
						ctx.AttachEvent(id, name,
							new_opts.Prevent, new_opts.Stop, new_opts.Handler)
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
				 	new_opts.Prevent, new_opts.Stop, new_opts.Handler)
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
				if is_text && StringEqual(*old_content, *new_content) {
					break
				}
			}
			ctx.SetText(id, String(*new_content))
		case *Children:
			var new_children = *new_content
			var diff_children = func(old_children Children, new_children Children) {
				var L int
				if len(old_children) > len(new_children) {
					L = len(old_children)
				} else {
					L = len(new_children)
				}
				for i := 0; i < L; i += 1 {
					var old_child *Node = nil
					var new_child *Node = nil
					if i < len(old_children) {
						old_child = old_children[i]
					}
					if i < len(new_children) {
						new_child = new_children[i]
					}
					Diff(ctx, node, old_child, new_child)
				}
			}
			if old != nil {
				switch old_content := old.Content.(type) {
				case *Text:
					ctx.SetText(old_id, String(""))
					diff_children(Children([] *Node {}), new_children)
				case *Children:
					var old_children = *old_content
					diff_children(old_children, new_children)
				}
			} else {
				diff_children(Children([] *Node {}), new_children)
			}
		}
		skip_content:
	}
}