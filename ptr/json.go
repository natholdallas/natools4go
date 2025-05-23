package ptr

import (
	"fmt"

	"github.com/natholdallas/natools4go/jsons"
)

func JSON(v ...any) {
	for _, i := range v {
		d, _ := jsons.String(i, true)
		fmt.Println(d)
	}
}
