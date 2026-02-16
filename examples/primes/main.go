package main

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/kyleraywed/derp"
)

// the number of ints to fill 200 MB
const size = (1024 * 1024 * 200) / int(unsafe.Sizeof(int(0)))

func main() {
	fmt.Printf("Size: %v ints / %v bytes\n", size, size*int(unsafe.Sizeof(int(0))))
	numbers := getList()

	start := time.Now()
	fmt.Print("Processing with Derp...\t")

	var pipe derp.Pipeline[int]

	pipe.Filter(func(value int) bool {
		return isPrime(value)
	})

	_, err := pipe.Apply(numbers)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Finished in %v\n", time.Since(start))

	rangeHolder := make([]int, 0, len(numbers))

	start = time.Now()
	fmt.Print("Processing via range...\t")

	for _, val := range numbers {
		if isPrime(val) {
			//lint:ignore SA4010 benchmarking
			rangeHolder = append(rangeHolder, val)
		}
	}

	fmt.Printf("Finished in %v\n", time.Since(start))
}

func getList() []int {
	numbers := make([]int, size)

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
