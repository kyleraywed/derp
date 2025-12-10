package main

/*
	- Present order invoice
*/

import (
	"fmt"

	"github.com/kyleraywed/dee"
)

func main() {
	var enum dee.Dee[byte]

	enum.Filter(func(value byte) bool {
		return value%2 == 0
	}, "Get just evens")

	enum.Map(func(value byte) byte {
		return value / 2
	}, "Half the value")

	enum.Skip(2)

	enum.Take(2)

	// No comment, empty slice
	enum.Foreach(func(value byte) {
		fmt.Println(value)
	}, "cons", "Print fast.")

	// Notice the index for this Map will be 1 since it's the second time Map is called.
	enum.Map(func(value byte) byte {
		return value + 1
	}, "Increment value", "Check the index")

	fmt.Println(enum)
}
