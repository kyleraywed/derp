package main

/*
	- Present order invoice
*/

import (
	"fmt"

	"github.com/kyleraywed/derp"
)

func main() {
	var enum derp.Pipeline[byte]

	enum.Filter(func(value byte) bool {
		return value%2 == 0
	}, "Get just evens")

	enum.Reduce(func(acc, value byte) byte {
		return acc + value
	}, "The reduce order is moved to the last order at Apply().", "This is the only instruction with a side-effect beside reset.")

	enum.Skip(2)

	// Notice the index for this Map will be 1 since it's the second time Map is called.
	enum.Map(func(index int, value byte) byte {
		return value + 1
	}, "Increment value", "Check the index")

	fmt.Println(enum)
	fmt.Printf("------------Apply()-------------\n\n")
	enum.Apply([]byte{0}) // Apply() returns early if slice is empty
	fmt.Println(enum)
}
