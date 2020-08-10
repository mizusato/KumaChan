/**
 *  @interface Signal
 *  @property {function(T): void} connect
 *  @property {function(T): void} disconnect
 *  @template T
 */
/**
 *  @interface Bridge
 *  @property {function(handler:string, event:Object): void} EmitEvent
 *  @property {function(): void} LoadFinish
 *  @property {Signal.<function(id:string, key:string, value:string): void>} ApplyStyle
 *  @property {Signal.<function(id:string, key:string): void>} EraseStyle
 *  @property {Signal.<function(id:string, name:string, prevent:boolean, stop:boolean, handler: string): void>} AttachEvent
 *  @property {Signal.<function(id:string, name:string, prevent:boolean, stop:boolean): void>} ModifyEvent
 *  @property {Signal.<function(id:string, name:string): void>} DetachEvent
 *  @property {Signal.<function(id:string, content:string): void>} SetText
 *  @property {Signal.<function(parent:string, id:string, tag:string) void>} AppendNode
 *  @property {Signal.<function(parent:string, id:string): void>} RemoveNode
 *  @property {Signal.<function(old_id:string, new_id:string): void>} UpdateNode
 *  @property {Signal.<function(target:string, id:string, tag:string): void>} ReplaceNode
 */

/** @type {Object.<string,HTMLElement>} */
let ElementRegistry = {}

console.log('[WebUI] Script Loaded')
window.addEventListener('load', _ => {
    ElementRegistry[0] = document.body
    if (typeof WebUI == 'undefined') {
        console.error('[WebUI] Bridge Object Not Found')
        return
    }
    /** @type {Bridge} */
    let bridge = WebUI
    console.log('[WebUI] Bridge Object Found')
    console.log(bridge)
    let create_listener = (prevent, stop, handler) => ev => {
        if (prevent) { ev.preventDefault() }
        if (stop) { ev.stopPropagation() }
        bridge.EmitEvent(handler, ev)
    }
    bridge.ApplyStyle.connect((id, key, val) => {
        console.log('ApplyStyle', { id, key, val })
        ElementRegistry[id].style[key] = val
    })
    bridge.EraseStyle.connect((id, key) => {
        console.log('EraseStyle', { id, key })
        ElementRegistry[id].style[key] = ''
    })
    bridge.AttachEvent.connect((id, name, prevent, stop, handler) => {
        console.log('AttachEvent', { id, name, prevent, stop, handler })
        let listener = create_listener(prevent, stop, handler)
        let el = ElementRegistry[id]
        el.addEventListener(name, listener)
        if (!(el._events)) { el._events = {} }
        el._events[name] = { listener, handler }
    })
    bridge.ModifyEvent.connect((id, name, prevent, stop) => {
        let el = ElementRegistry[id]
        let old_event = el._events[name]
        let handler = old_event.handler
        let old_listener = old_event.listener
        let listener = create_listener(prevent, stop, handler)
        el.removeEventListener(name, old_listener)
        el.addEventListener(name, listener)
        el._events[name] = { listener, handler }
    })
    bridge.DetachEvent.connect((id, name) => {
        let el = ElementRegistry[id]
        let event = el._events[name]
        el.removeEventListener(name, event.listener)
        delete el._events[name]
    })
    bridge.SetText.connect((id, text) => {
        console.log('SetText', { id, text })
        ElementRegistry[id].textContent = text
    })
    bridge.AppendNode.connect((parent, id, tag) => {
        console.log('AppendNode', { parent, id, tag })
        let el = document.createElement(tag)
        ElementRegistry[id] = el
        ElementRegistry[parent].appendChild(el)
    })
    bridge.RemoveNode.connect((parent, id) => {
        console.log('RemoveNode', { parent, id })
        ElementRegistry[parent].removeChild(ElementRegistry[id])
        delete ElementRegistry[id]
    })
    bridge.UpdateNode.connect((old_id, new_id) => {
        console.log('UpdateNode', { old_id, new_id })
        let el = ElementRegistry[old_id]
        delete ElementRegistry[old_id]
        ElementRegistry[new_id] = el
    })
    bridge.ReplaceNode.connect((parent, old_id, new_id, tag) => {
        console.log('ReplaceNode', { parent, old_id, new_id, tag })
        let parent_el = ElementRegistry[parent]
        let old_el = ElementRegistry[old_id]
        let new_el = document.createElement(tag)
        parent_el.insertBefore(new_el, old_el)
        parent_el.removeChild(old_el)
    })
    bridge.LoadFinish()
    console.log('LoadFinish')
})
