#!/usr/bin/env node

let KumaChan = require(`${__dirname}/build/dev/runtime.js`)
let ChildProcess = require('child_process')
let REPL = require('repl')
let Compiler = mode => `${__dirname}/build/dev/compiler ${mode}`

function load_modules () {
    let args = process.argv.map(a => a)
    args.shift()
    args.shift()
    if (args[0] == 'load') {
        for (let i=1; i<args.length; i++) {
            let module_path = args[i]
            console.log(`loading module ${module_path}`)
            let cmd = `${Compiler('module')} ${module_path}`
            let stdout = ChildProcess.execSync(cmd)
            eval(String(stdout))
        }
    }
}


function k_eval (command, context, filename, callback) {
    if (command.match(/^[ \t\n]*$/) != null) {
        callback(null)
        return
    }
    let p = ChildProcess.exec(Compiler('eval'), (error, stdout) => {
        if (error === null) {
            try {
                let value = eval(stdout)
                if (value !== KumaChan.Void) {
                    callback(null, value)
                } else {
                    callback(null)
                }
            } catch (error) {
                if (
                    error instanceof KumaChan.CustomError
                    || error instanceof KumaChan.RuntimeError
                ) {
                    console.log(error.message)
                    callback(null)
                } else {
                    callback(null)
                    throw error
                }
            }
        } else {
            console.log(error.message.replace(/^Command failed[^\n]*\n/, ''))
            callback(null)
        }
    })
    p.stdin.write(command)
    p.stdin.end()
}

load_modules()
let instance = REPL.start({ prompt: '>> ', eval: k_eval })
instance.context.KumaChan = KumaChan
