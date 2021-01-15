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
 *  @property {function(size: number): void} UpdateRootFontSize
 *  @property {function(uuid: string, content: string): void} InjectCSS
 *  @property {function(uuid: string, content: string): void} InjectJS
 *  @property {function(uuid: string, content: string, family: string, weight: string, style: string): void} InjectTTF
 *  @property {Signal.<function(id:string, key:string, value:string): void>} ApplyStyle
 *  @property {Signal.<function(id:string, key:string): void>} EraseStyle
 *  @property {Signal.<function(id:string, name:string, value:string): void>} SetAttr
 *  @property {Signal.<function(id:string, name:string): void>} RemoveAttr
 *  @property {Signal.<function(id:string, name:string, prevent:boolean, stop:boolean, capture: boolean, handler: string): void>} AttachEvent
 *  @property {Signal.<function(id:string, name:string, prevent:boolean, stop:boolean, capture: boolean): void>} ModifyEvent
 *  @property {Signal.<function(id:string, name:string): void>} DetachEvent
 *  @property {Signal.<function(id:string, content:string): void>} SetText
 *  @property {Signal.<function(parent:string, id:string, tag:string) void>} AppendNode
 *  @property {Signal.<function(parent:string, id:string): void>} RemoveNode
 *  @property {Signal.<function(old_id:string, new_id:string): void>} UpdateNode
 *  @property {Signal.<function(target:string, id:string, tag:string): void>} ReplaceNode
 *  @property {Signal.<function(parent:string, a:string, b:string): void>} SwapNode
 */

const SVGNS = 'http://www.w3.org/2000/svg'
const S_Root = 'html'
const S_GlobalStyle = '#__webui_global_style'

/** @type {Object.<string,HTMLElement>} */
let elementRegistry = {}
/** @type {Object.<string, Object.<string, { listener: function, handler: string, capture: boolean }>>} */
let eventsRegistry = {}
function createElement(parent, tag) {
    let parent_el = elementRegistry[parent]
    if (tag == 'svg' || parent_el instanceof SVGElement) {
        return document.createElementNS(SVGNS, tag)
    } else {
        return document.createElement(tag)
    }
}

/** @type {Array.<function():void>} */
let updateHooks = []
let updateHooksQueued = false
function WebUiRegisterUpdateHook(f) {
    updateHooks.push(f)
    f()
    return () => {
        for (let i = 0; i < updateHooks.length; i += 1) {
            if (updateHooks[i] === f) {
                updateHooks.splice(i, 1)
                break
            }
        }
    }
}
function runAllUpdateHooks() {
    if (!(updateHooksQueued)) {
        setTimeout(function () {
            for (let f of updateHooks) {
                f()
            }
            updateHooksQueued = false
        }, 0)
        updateHooksQueued = true
    }
}

console.log('[WebUI] Script Loaded')
window.addEventListener('load', _ => {
    elementRegistry[0] = document.body
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
        /** @type {Element} */
        let target = ev.target
        /** @type {Element} */
        let currentTarget = ev.currentTarget
        if (target instanceof HTMLInputElement
            || target instanceof HTMLSelectElement) {
            ev.webuiValue = ev.target.value
            ev.webuiChecked = ev.target.checked
        }
        if (ev instanceof MouseEvent) {
            let bounds = currentTarget.getBoundingClientRect()
            ev.webuiCurrentTargetX = Math.round(ev.clientX - bounds.left)
            ev.webuiCurrentTargetY = Math.round(ev.clientY - bounds.top)
        }
        bridge.EmitEvent(handler, ev)
    }
    bridge.ApplyStyle.connect((id, key, val) => {
        try {
            elementRegistry[id].style[key] = val
            runAllUpdateHooks()
        } catch (err) {
            console.log('ApplyStyle', { id, key, val }, err)
        }
    })
    bridge.EraseStyle.connect((id, key) => {
        try {
            elementRegistry[id].style[key] = ''
            runAllUpdateHooks()
        } catch (err) {
            console.log('EraseStyle', { id, key }, err)
        }
    })
    bridge.SetAttr.connect((id, name, val) => {
        try {
            if (name == 'value') {
                elementRegistry[id].value = val
            } else if (name == 'checked' || name == 'disabled') {
                elementRegistry[id][name] = true
            } else {
                elementRegistry[id].setAttribute(name, val)
            }
            runAllUpdateHooks()
        } catch (err) {
            console.log('SetAttr', { id, name, val }, err)
        }
    })
    bridge.RemoveAttr.connect((id, name) => {
        try {
            if (name == 'value') {
                elementRegistry[id].value = ''
            } else if (name == 'checked' || name == 'disabled') {
                elementRegistry[id][name] = false
            } else {
                elementRegistry[id].removeAttribute(name)
            }
            runAllUpdateHooks()
        } catch (err) {
            console.log('RemoveAttr', { id, name }, err)
        }
    })
    bridge.AttachEvent.connect((id, name, prevent, stop, capture, handler) => {
        try {
            let listener = create_listener(prevent, stop, handler)
            let el = elementRegistry[id]
            // TODO: convenient handling for key events
            let event_kind = name.replace(/\..*/, '')
            el.addEventListener(event_kind, listener, Boolean(capture))
            if (!(eventsRegistry[id])) { eventsRegistry[id] = {} }
            eventsRegistry[id][name] = { listener, handler, capture }
            runAllUpdateHooks()
        } catch (err) {
            console.log('AttachEvent', { id, name, prevent, stop, capture, handler }, err)
        }
    })
    bridge.ModifyEvent.connect((id, name, prevent, capture, stop) => {
        try {
            let el = elementRegistry[id]
            let events = eventsRegistry[id]
            let old_event = events[name]
            let handler = old_event.handler
            let old_listener = old_event.listener
            let old_capture = old_event.capture
            let listener = create_listener(prevent, stop, handler)
            let event_kind = name.replace(/\..*/, '')
            el.removeEventListener(event_kind, old_listener, Boolean(old_capture))
            el.addEventListener(event_kind, listener, Boolean(capture))
            events[name] = { listener, handler, capture }
            runAllUpdateHooks()
        } catch (err) {
            console.log('ModifyEvent', { id, name, prevent, stop, capture }, err)
        }
    })
    bridge.DetachEvent.connect((id, name) => {
        try {
            let el = elementRegistry[id]
            let event = eventsRegistry[id][name]
            let event_kind = name.replace(/\..*/, '')
            el.removeEventListener(event_kind, event.listener)
            delete eventsRegistry[id][name]
            runAllUpdateHooks()
        } catch (err) {
            console.log('DetachEvent', { id, name }, err)
        }
    })
    bridge.SetText.connect((id, text) => {
        try {
            elementRegistry[id].textContent = text
            runAllUpdateHooks()
        } catch (err) {
            console.log('SetText', { id, text }, err)
        }
    })
    bridge.AppendNode.connect((parent, id, tag) => {
        try {
            let el = createElement(parent, tag)
            elementRegistry[id] = el
            elementRegistry[parent].appendChild(el)
            runAllUpdateHooks()
        } catch (err) {
            console.log('AppendNode', { parent, id, tag }, err)
        }
    })
    bridge.RemoveNode.connect((parent, id) => {
        try {
            elementRegistry[parent].removeChild(elementRegistry[id])
            delete elementRegistry[id]
            if (eventsRegistry[id]) {
                delete eventsRegistry[id]
            }
            runAllUpdateHooks()
        } catch (err) {
            console.log('RemoveNode', { parent, id }, err)
        }
    })
    bridge.UpdateNode.connect((old_id, new_id) => {
        try {
            let el = elementRegistry[old_id]
            delete elementRegistry[old_id]
            elementRegistry[new_id] = el
            if (eventsRegistry[old_id]) {
                eventsRegistry[new_id] = eventsRegistry[old_id]
                delete eventsRegistry[old_id]
            }
            runAllUpdateHooks()
        } catch (err) {
            console.log('UpdateNode', { old_id, new_id }, err)
        }
    })
    bridge.ReplaceNode.connect((parent, old_id, new_id, tag) => {
        try {
            let parent_el = elementRegistry[parent]
            let old_el = elementRegistry[old_id]
            let new_el = createElement(parent, tag)
            parent_el.insertBefore(new_el, old_el)
            parent_el.removeChild(old_el)
            delete elementRegistry[old_id]
            elementRegistry[new_id] = new_el
            if (eventsRegistry[old_id]) {
                delete eventsRegistry[old_id]
            }
            runAllUpdateHooks()
        } catch (err) {
            console.log('ReplaceNode', { parent, old_id, new_id, tag }, err)
        }
    })
    bridge.SwapNode.connect((parent, a, b) => {
        try {
            // TODO: keep the focus of inserted node (or its descendent node)
            let parent_el = elementRegistry[parent]
            let a_el = elementRegistry[a]
            let b_el = elementRegistry[b]
            if (a_el.nextElementSibling === b_el) {
                parent_el.insertBefore(b_el, a_el)
            } else if (b_el.nextElementSibling === a_el) {
                parent_el.insertBefore(a_el, b_el)
            } else {
                let placeholder = createElement(parent, 'div')
                parent_el.insertBefore(placeholder, b_el)
                parent_el.insertBefore(b_el, a_el)
                parent_el.insertBefore(a_el, placeholder)
                parent_el.removeChild(placeholder)
            }
            runAllUpdateHooks()
        } catch (err) {
            console.log('SwapNode', { parent, a, b }, err)
        }
    })
    bridge.UpdateRootFontSize.connect(size => {
        document.querySelector(S_Root).style.fontSize = `${size}px`
    })
    bridge.InjectCSS.connect((uuid, path) => {
        let link_tag = document.createElement('link')
        link_tag.dataset['uuid'] = uuid
        link_tag.rel = 'stylesheet'
        link_tag.href = `asset://${path}`
        document.head.appendChild(link_tag)
    })
    bridge.InjectJS.connect((uuid, path) => {
        let script_tag = document.createElement('script')
        script_tag.dataset['uuid'] = uuid
        script_tag.type = 'text/javascript'
        script_tag.src = `asset://${path}`
        document.head.appendChild(script_tag)
    })
    bridge.InjectTTF.connect((uuid, path, family, weight, style) => {
        let style_tag = document.createElement('style')
        style_tag.dataset['uuid'] = uuid
        style_tag.textContent = `@font-face {
            font-family: '${family}';
            src: url(asset://${path}) format('truetype');
            font-weight: ${weight};
            font-style: ${style};
        }`
        document.head.appendChild(style_tag)
    })
    bridge.LoadFinish()
    console.log('LoadFinish')
})

