package errors
// error check and panic

import "os"

// check aborts the current execution if err is non-nil.
func Check(err os.Error) {
        if err != nil {
		println(err.String())
                panic(err)
        }
}
