package main

/*
	- Demonstrate WithDeepClone()
*/

import (
	"fmt"
	"maps"
	"slices"

	"github.com/kyleraywed/dei"
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

	var enum dei.Dei[person]

	// Without this, Apply() will affect p because p contains reference types []string and map.
	// It can be placed anywhere before Apply().
	enum.WithDeepClone(func(value person) person {
		out := value
		out.tags = slices.Clone(value.tags)
		out.meta = maps.Clone(value.meta)
		return out
	})

	enum.Map(func(value person) person {
		value.tags = append(value.tags, "a")
		value.meta[4] = "four"
		return value
	})

	new := enum.Apply(p)

	fmt.Println("New:\t", new)
	fmt.Println("Old:\t", p)
}
