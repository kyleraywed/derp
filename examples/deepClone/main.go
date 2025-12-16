package main

/*
	- Demonstrate WithDeepClone()
	- NOTE: Auto deep-cloning is now the default
*/

import (
	"fmt"

	"github.com/kyleraywed/derp"
)

type person struct {
	name string
	tags []string
	meta map[int]string
}

func main() {
	p := []person{
		{
			name: "Kyle",
			tags: []string{"x", "y", "z"},
			meta: map[int]string{
				1: "one",
				2: "two",
				3: "three",
			},
		},
	}

	var pipe derp.Derp[person]

	pipe.Map(func(value person) person {
		value.tags = append(value.tags, "a")
		value.meta[4] = "four"
		return value
	})

	new := pipe.Apply(p)

	fmt.Println("New:\t", new)
	fmt.Println("Old:\t", p)
}
