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
 *  @property {Signal.<function(id:string, name:string, prevent:boolean, stop:boolean, handler: string): void>} AttachEvent
 *  @property {Signal.<function(id:string, name:string, prevent:boolean, stop:boolean): void>} ModifyEvent
 *  @property {Signal.<function(id:string, name:string): void>} DetachEvent
 *  @property {Signal.<function(id:string, content:string): void>} SetText
 *  @property {Signal.<function(parent:string, id:string, tag:string) void>} AppendNode
 *  @property {Signal.<function(parent:string, id:string): void>} RemoveNode
 *  @property {Signal.<function(old_id:string, new_id:string): void>} UpdateNode
 *  @property {Signal.<function(target:string, id:string, tag:string): void>} ReplaceNode
 */

const SVGNS = 'http://www.w3.org/2000/svg'
const S_Root = 'html'
const S_GlobalStyle = '#__webui_global_style'

/** @type {Object.<string,HTMLElement>} */
let elementRegistry = {}
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
        if (ev.target) {
            if (ev.target instanceof HTMLInputElement
                || ev.target instanceof HTMLSelectElement) {
                ev.value = ev.target.value
                ev.checked = ev.target.checked
            }
        }
        bridge.EmitEvent(handler, ev)
    }
    bridge.ApplyStyle.connect((id, key, val) => {
        // console.log('ApplyStyle', { id, key, val })
        elementRegistry[id].style[key] = val
        runAllUpdateHooks()
    })
    bridge.EraseStyle.connect((id, key) => {
        // console.log('EraseStyle', { id, key })
        elementRegistry[id].style[key] = ''
        runAllUpdateHooks()
    })
    bridge.SetAttr.connect((id, name, val) => {
        // console.log('SetAttribute', { id, name, val })
        if (name == 'value') {
            elementRegistry[id].value = val
        } else if (name == 'checked' || name == 'disabled') {
            elementRegistry[id][name] = true
        } else {
            elementRegistry[id].setAttribute(name, val)
        }
        runAllUpdateHooks()
    })
    bridge.RemoveAttr.connect((id, name) => {
        // console.log('RemoveAttribute', { id, name })
        if (name == 'value') {
            elementRegistry[id].value = ''
        } else if (name == 'checked' || name == 'disabled') {
            elementRegistry[id][name] = false
        } else {
            elementRegistry[id].removeAttribute(name)
        }
        runAllUpdateHooks()
    })
    bridge.AttachEvent.connect((id, name, prevent, stop, handler) => {
        // console.log('AttachEvent', { id, name, prevent, stop, handler })
        let listener = create_listener(prevent, stop, handler)
        let el = elementRegistry[id]
        el.addEventListener(name, listener)
        if (!(el._events)) { el._events = {} }
        el._events[name] = { listener, handler }
        runAllUpdateHooks()
    })
    bridge.ModifyEvent.connect((id, name, prevent, stop) => {
        // console.log('ModifyEvent', { id, name, prevent, stop })
        let el = elementRegistry[id]
        let old_event = el._events[name]
        let handler = old_event.handler
        let old_listener = old_event.listener
        let listener = create_listener(prevent, stop, handler)
        el.removeEventListener(name, old_listener)
        el.addEventListener(name, listener)
        el._events[name] = { listener, handler }
        runAllUpdateHooks()
    })
    bridge.DetachEvent.connect((id, name) => {
        // console.log('DetachEvent', { id, name })
        let el = elementRegistry[id]
        let event = el._events[name]
        el.removeEventListener(name, event.listener)
        delete el._events[name]
        runAllUpdateHooks()
    })
    bridge.SetText.connect((id, text) => {
        // console.log('SetText', { id, text })
        elementRegistry[id].textContent = text
        runAllUpdateHooks()
    })
    bridge.AppendNode.connect((parent, id, tag) => {
        // console.log('AppendNode', { parent, id, tag })
        let el = createElement(parent, tag)
        elementRegistry[id] = el
        elementRegistry[parent].appendChild(el)
        runAllUpdateHooks()
    })
    bridge.RemoveNode.connect((parent, id) => {
        // console.log('RemoveNode', { parent, id })
        elementRegistry[parent].removeChild(elementRegistry[id])
        delete elementRegistry[id]
        runAllUpdateHooks()
    })
    bridge.UpdateNode.connect((old_id, new_id) => {
        // console.log('UpdateNode', { old_id, new_id })
        let el = elementRegistry[old_id]
        delete elementRegistry[old_id]
        elementRegistry[new_id] = el
        runAllUpdateHooks()
    })
    bridge.ReplaceNode.connect((parent, old_id, new_id, tag) => {
        // console.log('ReplaceNode', { parent, old_id, new_id, tag })
        let parent_el = elementRegistry[parent]
        let old_el = elementRegistry[old_id]
        let new_el = createElement(parent, tag)
        parent_el.insertBefore(new_el, old_el)
        parent_el.removeChild(old_el)
        delete elementRegistry[old_id]
        elementRegistry[new_id] = new_el
        runAllUpdateHooks()
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

