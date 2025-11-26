// Package spew used to print anything
package spew

import (
	"bufio"
	"fmt"
	"os"

	"github.com/natholdallas/natools4go/jsons"
)

// Err used to print error
func Err(errs ...error) {
	for i := range errs {
		if errs[i] != nil {
			fmt.Println(errs[i])
		}
	}
}

// JSON used to jsonify any value then print
func JSON(v ...any) {
	for _, i := range v {
		d, _ := jsons.String(i, true)
		fmt.Println(d)
	}
}

// Struct used to print any value
func Struct(v ...any) {
	for _, i := range v {
		fmt.Printf("%+v\n", i)
	}
}

// File used to print file content, like cat
func File(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
