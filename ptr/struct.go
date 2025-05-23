package ptr

import "fmt"

func Struct(v ...any) {
	for _, i := range v {
		fmt.Printf("%#v\n", i)
	}
}
