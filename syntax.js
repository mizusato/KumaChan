const TOKEN_ORDER = [
    'string',
    'extend_string',
    'comment',
    'space',
    'line_feed',
    'symbol',
    'identifier'
]


const TOKEN = {
    string: {
        pattern: /^'([^']*)'/,
        extract: 1
    },
    extend_string: {
        pattern: /^"[^"]*"/,
        extract: 1
    },
    comment: {
        pattern: /^\/\*(.*)\*\//,
        extract: 1
    },
    space: {
        pattern: /^[ \t\r　]+/,
        extract: 0
    },
    line_feed: {
        pattern: /^\n+/,
        extract: 0
    },
    symbol: {
        pattern: /^[\+\-\*\/%^!~\&\|><=\{\}\[\]\(\)\,]/,
        extract: 0
    },
    identifier: {
        pattern: (
/^[^0-9\~\&\- \t\r\n　\*\.\,\{\[\('"\)\]\}\/][^ \t\r\n　\*\.\,\{\[\('"\)\]\}\/]*/
        ),
        extract: 0
    }
}
