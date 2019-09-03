package assertion

import "os"
import "fmt"

const __ColorRedBold = "\033[1m\033[31m"
const __ColorBrownBold = "\033[1m\033[33m"
const __ColorReset = "\033[0m"

func Assert (ok bool, msg string) {
    if ok {
        return
    }
    fmt.Fprintf (
        os.Stderr, "%v*** Internal Interpreter Error:%v %v%v\n",
        __ColorRedBold, __ColorBrownBold, msg, __ColorReset,
    )
    panic("Assertion Failed")
}
