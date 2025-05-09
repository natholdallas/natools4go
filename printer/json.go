package printer

import (
	"fmt"

	"github.com/natholdallas/natools4go/jsons"
)

func PrintJSON(v ...any) {
	for _, i := range v {
		d, _ := jsons.Stringify(i, true)
		fmt.Println(d)
	}
}
