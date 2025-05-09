package printer

import "fmt"

func PrintStruct(v ...any) {
	for _, i := range v {
		fmt.Printf("%#v\n", i)
	}
}
