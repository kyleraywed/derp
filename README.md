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

// Reduce sets a terminal operation that aggregates all elements of the pipeline into a single value.
//
// The provided function `in` is called with an accumulator and each element of the slice,
// in order. The result of each call becomes the new accumulator for the next element.
//
// Only one Reduce can be set per pipeline. It is automatically executed last
// regardless of the order in which it was added.
//
// Returns a slice containing a single value: the final accumulator.
func (pipeline *Derp[T]) Reduce(in func(acc T, value T) T, comments ...string) error

// Skip the first n items and yield the rest. Comment inferred.
func (pipeline *Derp[T]) Skip(n int) error

// Yield only the first n items. Comment inferred.
func (pipeline *Derp[T]) Take(n int) error

// Interpret orders on data. Return new slice.
//
// Options:
//   - "dpc" : "(d)eep-clone (p)ointer (c)ycles"; eg. doubly-linked lists. Implements clone.Slowly().
//   - "cfe" : "(c)oncurrent (f)or(e)ach"; function eval order is non-deterministic. Use with caution.
//   - "power-[25, 50, 75]"; throttle cpu usage to 25, 50, or 75%. Default is 100%.
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

    // Upon calling Apply(), orders are fulfilled in the order in which they are declared.
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

    // Fourth? NO! Reduce will ALWAYS be the LAST thing to run, it can only be declared
    // one time or it returns an error, and it returns a slice of a single element.
    // Apply() will error if the working slice is empty when the reduce order is given.
    err := pipeline.Reduce(func(acc int, value int) int {
        return acc + value
    })
    if err != nil {
        log.Fatal(err) 
    }

    // Fourth. 
    pipeline.Foreach(func(value int) { // print the values
        fmt.Println(value)
    })

    // Fifth. Take and skip still log inferred comments.
    if err = pipeline.Take(2); err != nil {
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