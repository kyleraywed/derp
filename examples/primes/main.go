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
	numbers := make([]int, size)
	//numbers := Range(size)

	var pipe derp.Pipeline[int]
	pipe.Map(func(index int, value int) int {
		return index + 1
	})

	start := time.Now()
	fmt.Print("Processing with Derp...\t")
	numbers, err := pipe.Apply(numbers)
	if err != nil {
		log.Fatal(err)
	}

	pipe.Filter(func(value int) bool {
		return isPrime(value)
	})

	_, err = pipe.Apply(numbers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Finished in %v\n", time.Since(start))

	start = time.Now()
	fmt.Print("Processing via range...\t")
	rangeHolder := make([]int, len(numbers))
	for _, val := range numbers {
		if isPrime(val) {
			rangeHolder = append(rangeHolder, val)
		}
	}
	fmt.Printf("Finished in %v\n", time.Since(start))
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
