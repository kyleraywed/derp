# Derp

A concurrency-driven, **d**eferred-**e**xecution, **r**eusable, data-processing **p**ipeline. 

```go
// Keep only the elements where in returns true. Optional comment strings.
func (pipeline *Derp[T]) Filter(in func(value T) bool, comments ...string)

// Perform logic using each element as an input. No changes to the underlying elements are made.
// Optional comment strings.
func (pipeline *Derp[T]) Foreach(in func(value T), comments ...string)

// Transform each value by applying a function. Optional comment strings.
func (pipeline *Derp[T]) Map(in func(value T) T, comments ...string)

// Skip the first n items and yield the rest. Comment inferred.
func (pipeline *Derp[T]) Skip(n int) error

// Yield only the first n items. Comment inferred.
func (pipeline *Derp[T]) Take(n int) error

// Interpret orders on data. Return new slice.
//
// Options:
//   - "dpc" : "(d)eep-clone (p)ointer (c)ycles"; eg. doubly-linked lists. Implements clone.Slowly().
//   - "cfe" : "(c)oncurrent (f)or(e)ach"; function eval order is non-deterministic. Use with caution.
//   - "power-[30, 50, 70]"; throttle cpu usage to 30, 50, or 70%. Default is 100%. Last power comment wins.
func (pipeline *Derp[T]) Apply(input []T, options ...string) ([]T, error)
```

Usage

```go
package main

import (
    "log"

    "github.com/kyleraywed/derp"
)

func main() {
    // Create a new instruction pipeline for the type of data being processed.
    var pipeline derp.Derp[int]

    // Instruct methods occur in the order in which they are declared upon calling Apply()
    // So this would happen first.
    pipeline.Filter(func(value int) bool {
        return value % 2 == 0
    }, "Get just the evens")

    // Second. Also notice the optional comment. String() is implemented;
    // fmt.Print()ing the object presents a detailed order invoice.
    pipeline.Map(func(value int) int {
        return value * 2
    }, "Double them")

    // Third.
    pipeline.Filter(func(value int) bool { // get just values > 10
        return value > 10
    })

    // Fourth. 
    pipeline.Foreach(func(value int) { // print the values
        fmt.Println(value)
    })

    // Last. Take and skip still log inferred comments.
    if err := pipeline.Take(2); err != nil {
        log.Println(err)
    } // get just the first 2 elements

    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    // Apply() performs the actual work orders declared above.
    // Input is implicitly deep-cloned and won't see side-effects.
    // The pipeline does not consume, and is safely reusable.
    output, err := pipeline.Apply(numbers)
    if err != nil {
        log.Println(err)
    }
    
    // []int{12, 16}
    // [12, 16, 20] will print when Apply is run since
    // Foreach() was called before Take()
}
```
Notes and design
-
- Deep cloning is handled via [go-clone](https://github.com/huandu/go-clone).
- Derp is **not** safe for concurrent use.