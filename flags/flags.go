// Package flags provides a set of functions to parse command line arguments
package flags

import "os"

func Run(s bool, script func(), continuer ...bool) {
	if !s {
		return
	}
	script()
	c := false
	if len(continuer) > 0 {
		c = continuer[0]
	}
	if !c {
		os.Exit(0)
	}
}
