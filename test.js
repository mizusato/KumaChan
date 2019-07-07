#!/usr/bin/env node

let KumaChan = require(`${__dirname}/build/dev/runtime.js`)
let ChildProcess = require('child_process')
let FS = require('fs')
let Compiler = mode => `${__dirname}/build/dev/compiler ${mode}`
let e = String.fromCodePoint(27)
let Bold = `${e}[1m`
let Reset = `${e}[0m`
let Red = `${e}[31m`
let Green = `${e}[32m`


function assert (expr) {
    if (!expr) {
        throw Error('Assertion Failed')
    }
    return expr
}


let previous_length = 0

function display (info) {
    assert(typeof info == 'string')
    process.stdout.write('\r')
    process.stdout.write(' '.repeat(previous_length))
    process.stdout.write('\r')
    process.stdout.write(info)
    previous_length = info.length
}

function print (info) {
    assert(typeof info == 'string')
    process.stdout.write(info)
    process.stdout.write('\n')
    previous_length = 0
}


function eval_test (code) {
    let resolve = null
    let reject = null
    let promise = new Promise((res, rej) => { resolve = res, reject = rej })
    let p = ChildProcess.exec(Compiler('eval'), (error, stdout) => {
        if (error != null) {
            print(`${Bold}${Red} Error occured during compiling:${Reset}`)
            console.log(error)
            process.exit(-1)
        }
        try {
            let result = eval(stdout)
            if (result instanceof Promise) {
                result
                    .then(_ => { resolve({ ok: true }) })
                    .catch(error => { reject({ ok: false, error }) })
            } else {
                resolve({ ok: true })
            }
        } catch (error) {
            resolve({ ok: false, error })
        }
    })
    p.stdin.write(code)
    p.stdin.end()
    return promise
}


async function main () {
    let index_text = FS.readFileSync(`${__dirname}/test/index.json`)
    let index = JSON.parse(index_text)
    assert(index.sequence instanceof Array)
    let total = 0
    let passed = 0
    for (let group of index.sequence) {
        assert(typeof group == 'string')
        let units = FS.readdirSync(`${__dirname}/test/${group}`)
        for (let unit of units) {
            let unit_name = unit.replace(/\.k$/, '')
            let info_base = (
                `${Bold}Running Unit Test:${Reset} ${group}/${unit_name}`
            )
            display(`${info_base} ...`)
            let code = FS.readFileSync(`${__dirname}/test/${group}/${unit}`)
            let result = await eval_test(code)
            if (result.ok) {
                display(`${info_base} [ ${Bold}${Green}OK${Reset} ]`)
                passed += 1
            } else {
                display(`${info_base} [ ${Bold}${Red}FAILED${Reset} ]`)
                print('')
                let error = result.error
                if (
                    error instanceof KumaChan.CustomError
                    || error instanceof KumaChan.RuntimeError
                ) {
                    print(error.message)
                } else {
                    console.log(error)
                }
            }
            total += 1
        }
    }
    if (passed == total) {
        display(`${Bold}Test Finished, ${Green}All Passed.${Reset}`)
        print('')
    } else {
        print('')
        print(`${Bold}Test Finished, ${Red}${passed}/${total} Passed.${Reset}`)
        process.exit(1)
    }
}


main()
