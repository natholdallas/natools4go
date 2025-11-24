// Package defers provide some used defer function
package defers

import (
	"fmt"

	"github.com/natholdallas/natools4go/ptr"
)

func PrintErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func PrintValue(v ...any) {
	ptr.JSON(v)
}
