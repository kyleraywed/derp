package main

/*
	- Benchmark the time it takes to filter the prime numbers from 100MB of random bytes.
*/

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/kyleraywed/dei"
)

const size = 1024 * 1024 * 100

func main() {
	fmt.Printf("Size: %v bytes\n\n", size)
	start := time.Now()
	fmt.Print("Allocating... ")
	numbers := make([]byte, 0, size)
	for range size {
		numbers = append(numbers, byte(rand.IntN(256)))
	}
	fmt.Printf("Finished in %v\n", time.Since(start))

	start = time.Now()
	fmt.Print("Processing... ")
	var iter dei.Dei[byte]

	iter.Filter(func(value byte) bool {
		if value < 2 {
			return false
		}
		if value == 2 || value == 3 {
			return true
		}
		if value%2 == 0 || value%3 == 0 {
			return false
		}

		for i := 5; i*i <= int(value); i += 6 {
			if int(value)%i == 0 || int(value)%(i+2) == 0 {
				return false
			}
		}

		return true
	})

	_ = iter.Apply(numbers)
	fmt.Printf("Finished in %v\n", time.Since(start))
}
