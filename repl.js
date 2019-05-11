#!/usr/bin/env node

let KumaChan = require(`${__dirname}/build/dev/runtime.js`)
let ChildProcess = require('child_process')
let REPL = require('repl')
let Compiler = `${__dirname}/build/dev/compiler --eval`


function k_eval (command, context, filename, callback) {
    let p = ChildProcess.exec(Compiler, (error, stdout) => {
        if (error === null) {
            try {
                callback(null, eval(stdout))
            } catch (error) {
                console.log(error.message)
                callback(null)
            }
        } else {
            console.log(error.message)
            callback(null)
        }
    })
    p.stdin.write(command)
    p.stdin.end()
}

let instance = REPL.start({ prompt: '>> ', eval: k_eval })
instance.context.KumaChan = KumaChan
