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
 *      PatchActualDOM: Signal<(data:string) => void>
 *  }} Bridge
 */
/**
 *  @typedef {{
 *      ApplyStyle: (id:string, key:string, value:string) => void,
 *      EraseStyle: (id:string, key:string) => void,
 *      SetAttr: (id:string, name:string, value:string) => void,
 *      RemoveAttr: (id:string, name:string) => void,
 *      AttachEvent: (id:string, name:string, prevent:boolean, stop:boolean, capture: boolean, handler: string) => void,
 *      ModifyEvent: (id:string, name:string, prevent:boolean, stop:boolean, capture: boolean) => void,
 *      DetachEvent: (id:string, name:string) => void,
 *      SetText: (id:string, content:string) => void,
 *      AppendNode: (parent:string, id:string, tag:string) => void
 *      RemoveNode: (parent:string, id:string) => void,
 *      UpdateNode: (old_id:string, new_id:string) => void,
 *      ReplaceNode: (target:string, id:string, tag:string) => void,
 *      SwapNode: (parent:string, a:string, b:string) => void,
 *      MoveNode: (parent:string, id:string, pivot:string) => void
 *  }} PatchOperations
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

window.addEventListener('load', _ => {
    let bridge = getBridgeObject()
    connectGeneralSignals(bridge)
    registerBodyElement()
    initializeInteraction(bridge)
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
function initializeInteraction(bridge) {
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
    /** @type {PatchOperations} */
    let patchOperations = {
        ApplyStyle: (id, key, val) => {
            try {
                elementRegistry[id].style[key] = val
            } catch (err) {
                console.log('ApplyStyle', { id, key, val }, err)
            }
        },
        EraseStyle: (id, key) => {
            try {
                elementRegistry[id].style[key] = ''
            } catch (err) {
                console.log('EraseStyle', { id, key }, err)
            }
        },
        SetAttr: (id, name, val) => {
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
        },
        RemoveAttr: (id, name) => {
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
        },
        AttachEvent: (id, name, prevent, stop, capture, handler) => {
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
        },
        ModifyEvent: (id, name, prevent, capture, stop) => {
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
        },
        DetachEvent: (id, name) => {
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
        },
        SetText: (id, text) => {
            try {
                elementRegistry[id].textContent = text
            } catch (err) {
                console.log('SetText', { id, text }, err)
            }
        },
        AppendNode: (parent, id, tag) => {
            try {
                let el = createElement(parent, tag)
                elementRegistry[id] = el
                elementRegistry[parent].appendChild(el)
            } catch (err) {
                console.log('AppendNode', { parent, id, tag }, err)
            }
        },
        RemoveNode: (parent, id) => {
            try {
                elementRegistry[parent].removeChild(elementRegistry[id])
                delete elementRegistry[id]
                if (eventsRegistry[id]) {
                    delete eventsRegistry[id]
                }
            } catch (err) {
                console.log('RemoveNode', { parent, id }, err)
            }
        },
        UpdateNode: (old_id, new_id) => {
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
        },
        ReplaceNode: (parent, old_id, new_id, tag) => {
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
        },
        SwapNode: (parent, a, b) => {
            try {
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
        },
        MoveNode: (parent, id, pivot) => {
            try {
                let parent_el = elementRegistry[parent]
                let el = elementRegistry[id]
                let pivot_el = elementRegistry[pivot]
                let restore = keepFocus(el)
                parent_el.insertBefore(el, pivot_el)
                restore()
            } catch (err) {
                console.log('MoveNode', { parent, id, pivot }, err)
            }
        }
    }
    bridge.PatchActualDOM.connect(data => {
        try {
            /** @type {Array<string|boolean>} */
            let items = JSON.parse(data)
            let i = 0
            let L = items.length
            while (i < L) {
                /** @type {string} */
                // @ts-ignore
                let name = items[i]
                if (patchOperations[name] instanceof Function) {
                    /** @type {Function} */
                    let op = patchOperations[name]
                    let args = []
                    for (let j = 0; j < op.length; j += 1) {
                        i += 1
                        args.push(items[i])
                    }
                    op.apply(null, args)
                } else {
                    throw new Error(`unknown patch operation: ${name}`)
                }
                i += 1
            }
            runAllUpdateHooks()
        } catch (err) {
            console.log(`error patching DOM: ${data}`, err)
        }
    })
}

