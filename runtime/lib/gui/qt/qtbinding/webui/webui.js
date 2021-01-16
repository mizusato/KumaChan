/**
 *  @template T
 *  @typedef {{
 *      connect: (callback: T) => void,
 *      disconnect: (callback: T) => void
 *  }} Signal
 */
/**
 *  @typedef {{
 *      EmitEvent: (handler:string, event:Object) => void,
 *      LoadFinish: () => void,
 *      UpdateRootFontSize: Signal<(size:number) => void>,
 *      InjectCSS: Signal<(uuid:string, content:string) => void>,
 *      InjectJS: Signal<(uuid:string, content:string) => void>,
 *      InjectTTF: Signal<(uuid: string, content: string, family: string, weight: string, style: string) => void>,
 *      ApplyStyle: Signal<(id:string, key:string, value:string) => void>,
 *      EraseStyle: Signal<(id:string, key:string) => void>,
 *      SetAttr: Signal<(id:string, name:string, value:string) => void>,
 *      RemoveAttr: Signal<(id:string, name:string) => void>,
 *      AttachEvent: Signal<(id:string, name:string, prevent:boolean, stop:boolean, capture: boolean, handler: string) => void>,
 *      ModifyEvent: Signal<(id:string, name:string, prevent:boolean, stop:boolean, capture: boolean) => void>,
 *      DetachEvent: Signal<(id:string, name:string) => void>,
 *      SetText: Signal<(id:string, content:string) => void>,
 *      AppendNode: Signal<(parent:string, id:string, tag:string) => void>
 *      RemoveNode: Signal<(parent:string, id:string) => void>,
 *      UpdateNode: Signal<(old_id:string, new_id:string) => void>,
 *      ReplaceNode: Signal<(target:string, id:string, tag:string) => void>,
 *      SwapNode: Signal<(parent:string, a:string, b:string) => void>,
 *      PerformActualRendering: Signal<() => void>
 *  }} Bridge
 */

const SvgXmlNamespace = 'http://www.w3.org/2000/svg'
const RootElementSelector = 'html'

/** @type {{ [id:string]: HTMLElement | SVGElement }} */
let elementRegistry = {}
/** @type {{ [id:string]: { [name:string]: { listener: function, handler: string, capture: boolean } } }} */
let eventsRegistry = {}
/** @returns {void} */
function registerBodyElement() {
    elementRegistry[0] = document.body
}

/** @type {Array<() => void>} */
let updateHooks = []
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
    for (let f of updateHooks) {
        f()
    }
}

/** @type {Array<() => void>} */
let patchOperationQueue = []
/**
 *  @template T
 *  @param {Signal<T>} signal
 *  @param {T} callback 
 */
function connectPatchSignal(signal, callback) {
    // @ts-ignore
    signal.connect((...args) => {
        // @ts-ignore
        patchOperationQueue.push(() => callback(...args))
    })
}
function flushPatchOperationQueue() {
    for (let f of patchOperationQueue) {
        f()
    }
    patchOperationQueue.length = 0
}

window.addEventListener('load', _ => {
    let bridge = getBridgeObject()
    connectGeneralSignals(bridge)
    registerBodyElement()
    connectUpdateSignals(bridge)
    notifyInitialized(bridge)
})

/** @returns {Bridge} */
function getBridgeObject() {
    // @ts-ignore
    if (typeof WebUI == 'undefined') {
        console.error('[WebUI] Bridge Object Not Found')
        throw new Error('[WebUI] Initialization Failed - Bridge Object Not Found')
    }
    /** @type {Bridge} */
    // @ts-ignore
    let bridge = WebUI
    console.log('[WebUI] Bridge Object Found', bridge)
    return bridge
}

/** @param {Bridge} bridge */
function notifyInitialized(bridge) {
    bridge.LoadFinish()
    console.log('[WebUI] Bridge JS-Side Initialized')
}

/** @param {Bridge} bridge */
function connectGeneralSignals(bridge) {
    bridge.UpdateRootFontSize.connect(size => {
        let root = document.querySelector(RootElementSelector)
        root.style.fontSize = `${size}px`
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
}

/** @param {Bridge} bridge */
function connectUpdateSignals(bridge) {
    /** @type {(parent:string, tag:string) => HTMLElement | SVGElement} */
    let createElement = (parent, tag) => {
        let parent_el = elementRegistry[parent]
        if (tag == 'svg' || parent_el instanceof SVGElement) {
            return document.createElementNS(SvgXmlNamespace, tag)
        } else {
            return document.createElement(tag)
        }
    }
    /** @type {(prevent:boolean, stop:boolean, handler:string) => (ev:any) => void} */
    let createListener = (prevent, stop, handler) => ev => {
        if (prevent) { ev.preventDefault() }
        if (stop) { ev.stopPropagation() }
        /** @type {Element} */
        let target = ev.target
        /** @type {Element} */
        let currentTarget = ev.currentTarget
        if (target instanceof HTMLInputElement
            || target instanceof HTMLSelectElement) {
            ev['webuiValue'] = ev.target.value
            ev['webuiChecked'] = ev.target.checked
        }
        if (ev instanceof MouseEvent) {
            let bounds = currentTarget.getBoundingClientRect()
            ev['webuiCurrentTargetX'] = Math.round(ev.clientX - bounds.left)
            ev['webuiCurrentTargetY'] = Math.round(ev.clientY - bounds.top)
        }
        bridge.EmitEvent(handler, ev)
    }
    bridge.PerformActualRendering.connect(() => {
        flushPatchOperationQueue()
        runAllUpdateHooks()
    })
    connectPatchSignal(bridge.ApplyStyle, (id, key, val) => {
        try {
            elementRegistry[id].style[key] = val
        } catch (err) {
            console.log('ApplyStyle', { id, key, val }, err)
        }
    })
    connectPatchSignal(bridge.EraseStyle, (id, key) => {
        try {
            elementRegistry[id].style[key] = ''
        } catch (err) {
            console.log('EraseStyle', { id, key }, err)
        }
    })
    connectPatchSignal(bridge.SetAttr, (id, name, val) => {
        try {
            if (name == 'value') {
                elementRegistry[id]['value'] = val
            } else if (name == 'checked' || name == 'disabled') {
                elementRegistry[id][name] = true
            } else {
                elementRegistry[id].setAttribute(name, val)
            }
        } catch (err) {
            console.log('SetAttr', { id, name, val }, err)
        }
    })
    connectPatchSignal(bridge.RemoveAttr, (id, name) => {
        try {
            if (name == 'value') {
                elementRegistry[id]['value'] = ''
            } else if (name == 'checked' || name == 'disabled') {
                elementRegistry[id][name] = false
            } else {
                elementRegistry[id].removeAttribute(name)
            }
        } catch (err) {
            console.log('RemoveAttr', { id, name }, err)
        }
    })
    connectPatchSignal(bridge.AttachEvent, (id, name, prevent, stop, capture, handler) => {
        try {
            let listener = createListener(prevent, stop, handler)
            let el = elementRegistry[id]
            // TODO: convenient handling for key events
            let event_kind = name.replace(/\..*/, '')
            el.addEventListener(event_kind, listener, Boolean(capture))
            if (!(eventsRegistry[id])) { eventsRegistry[id] = {} }
            eventsRegistry[id][name] = { listener, handler, capture }
        } catch (err) {
            console.log('AttachEvent', { id, name, prevent, stop, capture, handler }, err)
        }
    })
    connectPatchSignal(bridge.ModifyEvent, (id, name, prevent, capture, stop) => {
        try {
            let el = elementRegistry[id]
            let events = eventsRegistry[id]
            let old_event = events[name]
            let handler = old_event.handler
            let old_listener = old_event.listener
            let old_capture = old_event.capture
            let listener = createListener(prevent, stop, handler)
            let event_kind = name.replace(/\..*/, '')
            // @ts-ignore
            el.removeEventListener(event_kind, old_listener, Boolean(old_capture))
            el.addEventListener(event_kind, listener, Boolean(capture))
            events[name] = { listener, handler, capture }
        } catch (err) {
            console.log('ModifyEvent', { id, name, prevent, stop, capture }, err)
        }
    })
    connectPatchSignal(bridge.DetachEvent, (id, name) => {
        try {
            let el = elementRegistry[id]
            let event = eventsRegistry[id][name]
            let event_kind = name.replace(/\..*/, '')
            // @ts-ignore
            el.removeEventListener(event_kind, event.listener)
            delete eventsRegistry[id][name]
        } catch (err) {
            console.log('DetachEvent', { id, name }, err)
        }
    })
    connectPatchSignal(bridge.SetText, (id, text) => {
        try {
            elementRegistry[id].textContent = text
        } catch (err) {
            console.log('SetText', { id, text }, err)
        }
    })
    connectPatchSignal(bridge.AppendNode, (parent, id, tag) => {
        try {
            let el = createElement(parent, tag)
            elementRegistry[id] = el
            elementRegistry[parent].appendChild(el)
        } catch (err) {
            console.log('AppendNode', { parent, id, tag }, err)
        }
    })
    connectPatchSignal(bridge.RemoveNode, (parent, id) => {
        try {
            elementRegistry[parent].removeChild(elementRegistry[id])
            delete elementRegistry[id]
            if (eventsRegistry[id]) {
                delete eventsRegistry[id]
            }
        } catch (err) {
            console.log('RemoveNode', { parent, id }, err)
        }
    })
    connectPatchSignal(bridge.UpdateNode, (old_id, new_id) => {
        try {
            let el = elementRegistry[old_id]
            delete elementRegistry[old_id]
            elementRegistry[new_id] = el
            if (eventsRegistry[old_id]) {
                eventsRegistry[new_id] = eventsRegistry[old_id]
                delete eventsRegistry[old_id]
            }
        } catch (err) {
            console.log('UpdateNode', { old_id, new_id }, err)
        }
    })
    connectPatchSignal(bridge.ReplaceNode, (parent, old_id, new_id, tag) => {
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
        } catch (err) {
            console.log('ReplaceNode', { parent, old_id, new_id, tag }, err)
        }
    })
    connectPatchSignal(bridge.SwapNode, (parent, a, b) => {
        /** @param {Element} el */
        let keepFocus = el => {
            let focus = document.activeElement
            if (focus && el.contains(focus)) {
                // @ts-ignore
                return () => { focus.focus && focus.focus() }
            } else {
                return () => {}
            }
        }
        try {
            // TODO: keep the focus of inserted node (or its descendent node)
            let parent_el = elementRegistry[parent]
            let a_el = elementRegistry[a]
            let b_el = elementRegistry[b]
            if (a_el.nextElementSibling === b_el) {
                let restore = keepFocus(b_el)
                parent_el.insertBefore(b_el, a_el)
                restore()
            } else if (b_el.nextElementSibling === a_el) {
                let restore = keepFocus(a_el)
                parent_el.insertBefore(a_el, b_el)
                restore()
            } else {
                let placeholder = createElement(parent, 'div')
                parent_el.insertBefore(placeholder, b_el)
                let restore = keepFocus(b_el)
                parent_el.insertBefore(b_el, a_el)
                restore()
                restore = keepFocus(a_el)
                parent_el.insertBefore(a_el, placeholder)
                restore()
                parent_el.removeChild(placeholder)
            }
        } catch (err) {
            console.log('SwapNode', { parent, a, b }, err)
        }
    })
}

