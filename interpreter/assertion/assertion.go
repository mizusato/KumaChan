package assertion

import "os"
import "fmt"

func Assert (ok bool, msg string) {
    if ok {
        return
    }
    fmt.Fprintf (
        os.Stderr, "Internal Interpreter Error: %v: %v",
        "Assertion Failed", msg,
    )
    os.Exit(222)
}
