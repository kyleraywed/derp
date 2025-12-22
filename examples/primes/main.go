package main

/*
	- Benchmark the time it takes to filter the prime numbers from 100MB of random ints.
*/

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"
	"unsafe"

	"github.com/kyleraywed/derp"
)

const size = (1024 * 1024 * 100) / int(unsafe.Sizeof(int(0)))

func main() {
	fmt.Printf("Size: %v ints / %v bytes\n\n", size, size*int(unsafe.Sizeof(int(0))))
	start := time.Now()
	fmt.Print("Allocating... ")

	numbers := make([]int, size)
	var allocPipe derp.Pipeline[int]
	allocPipe.Map(func(value int) int {
		return rand.IntN(256)
	})
	numbers, err := allocPipe.Apply(numbers)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Finished in %v\n", time.Since(start))

	start = time.Now()
	fmt.Print("Processing... ")
	// new pipeline required as running Apply doesn't consume
	var primePipe derp.Pipeline[int]

	primePipe.Filter(func(value int) bool {
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

	_, err = primePipe.Apply(numbers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Finished in %v\n", time.Since(start))
}
