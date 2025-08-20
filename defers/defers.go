// Package defers provide some used defer function
package defers

import (
	"fmt"

	"github.com/natholdallas/natools4go/ptr"
)

func PrintErr(err error) {
	fmt.Println(err)
}

func PrintReturnValue(v ...any) {
	ptr.JSON(v)
}
