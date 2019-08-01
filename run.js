#!/usr/bin/env node


let KumaChan = require(`${__dirname}/build/dev/runtime.js`)
let Compiler = (mode, f) => `${__dirname}/build/dev/compiler ${mode} ${f}`
let ChildProcess = require('child_process')
let FS = require('fs')


function run_module (file) {
    let transpiled = ChildProcess.execSync (
        Compiler('module', file),
        { encoding: 'utf8' }
    )
    eval(transpiled)
}


for (let i = 2; i < process.argv.length; i += 1) {
    run_module(process.argv[i])
}
