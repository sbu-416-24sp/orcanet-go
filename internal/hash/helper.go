package hash

import "fmt"

func Assert(pred bool, format string, args ...any) {
	if !pred {
		panic(fmt.Sprintf(format, args...))
	}
}
