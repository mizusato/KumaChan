console.log('[WebUI] Script Loaded')
window.addEventListener('load', _ => {
    if (typeof WebUI == 'undefined') {
        console.error('[WebUI] Bridge Object Not Found')
        return
    }
    console.log('[WebUI] Bridge Object Found')
    console.log(WebUI)
})
