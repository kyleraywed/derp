package main

import (
	"fmt"
	"log"
	"slices"
	"time"
	"unsafe"

	"github.com/kyleraywed/derp"
)

// the size in bytes of numbers
const size = 1024 * 1024 * 100 // 100 MB

func main() {
	numbers := getList(size)
	numbers2 := slices.Clone(numbers)

	{ // derp clone
		fmt.Print("Processing with Derp/Clone...\t")
		start := time.Now()

		var pipe derp.Pipeline[int]

		pipe.Filter(func(value int) bool {
			return isPrime(value)
		})

		_, err := pipe.Apply(numbers)
		if err != nil {
			log.Fatal(err)
		}

		done := time.Since(start)

		fmt.Printf("Finished in %v\n", done)
	}

	{ // derp inplace
		fmt.Print("Processing with Derp/InPlace...\t")
		start := time.Now()

		var pipe2 derp.Pipeline[int]

		pipe2.Filter(func(value int) bool {
			return isPrime(value)
		})

		_, err := pipe2.Apply(numbers2, derp.Opt_InPlace)
		if err != nil {
			log.Fatal(err)
		}

		done := time.Since(start)

		fmt.Printf("Finished in %v\n", done)
	}

	{ // range
		rangeHolder := make([]int, 0, len(numbers))
		fmt.Print("Processing via range...\t")

		start := time.Now()

		for _, val := range numbers {
			if isPrime(val) {
				//lint:ignore SA4010 benchmarking
				rangeHolder = append(rangeHolder, val)
			}
		}

		done := time.Since(start)

		fmt.Printf("Finished in %v\n", done)
	}

}

// return a slice of ints, incrementally valued, that takes up 'size' bytes
func getList(size int) []int {
	if size/int(unsafe.Sizeof(int(0))) < 1 {
		return nil
	}

	numbers := make([]int, size/int(unsafe.Sizeof(int(0))))
	var pipe derp.Pipeline[int]

	pipe.Map(func(index int, value int) int {
		return index + 1
	}, "Map the same numbers every time for consistency")

	numbers, err := pipe.Apply(numbers)
	if err != nil {
		log.Fatal(err)
	}

	return numbers
}

func isPrime(value int) bool {
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
}
