let ElementRegistry = {}

console.log('[WebUI] Script Loaded')
window.addEventListener('load', _ => {
    ElementRegistry[0] = document.body
    if (typeof WebUI == 'undefined') {
        console.error('[WebUI] Bridge Object Not Found')
        return
    }
    console.log('[WebUI] Bridge Object Found')
    console.log(WebUI)
    WebUI.AppendNode.connect((parent, id, tag) => {
        console.log('AppendNode', { parent, id, tag })
        let el = document.createElement(tag)
        ElementRegistry[id] = el
        ElementRegistry[parent].appendChild(el)
    })
    WebUI.SetText.connect((id, text) => {
        console.log('SetText', { id, text })
        ElementRegistry[id].textContent = text
    })
    WebUI.LoadFinish()
    console.log('LoadFinish')
})
