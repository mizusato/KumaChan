invoke {
    function props_to_string (props: Hash) -> String {
        return (
            props -> map_entry -> .{ ' ${1}="${2}"' << [.name, str(.value)] }
                    -> join('')
        )
    }
    function div (props: Hash, children: List) -> String {
        return '<div${1}>${2}</div>' << [
            props_to_string(props),
            children -> map -> .{
                when {
                    .child is Iterable: .child,
                    otherwise: [.child]
                }
            }
            -> flat
            -> map -> .{ str .child }
            -> join('')
        ]
    }
    let d1 = tree { div {} }
    let d1_ = |-> div {}
    assert d1 == d1_
    assert d1 == '<div></div>'
    let d2 = |-> div { x=1; y=2 }
    let ok = lambda(x) { return x == '<div x="1" y="2"></div>' }
    assert ok(d2)
    assert ok() -> tree { div { x=1; y=2 } }
    let d3 = |-> div {
        div { = 'Hello'; = ' '; = 'World' }
        div { = 'Change The World' }
    }
    assert d3 == '<div><div>Hello World</div><div>Change The World</div></div>'
    let d4 = tree {
        div {
            id = 'root'
            class = 'container'
            div {
                id = 'header'
                tip = 'Header'
            }
            = 'List:'
            = [ |-> div {
                'data-count' = i
                div {
                    = i+1
                }
            }, for i in seq(3) ]
        }
    }
    assert d4 == '<div id="root" class="container"><div id="header" tip="Header"></div>List:<div data-count="0"><div>1</div></div><div data-count="1"><div>2</div></div><div data-count="2"><div>3</div></div></div>'
}
