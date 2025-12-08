# Dee

A fast, concurrency-driven **d**eferred-**e**xecution **e**numerator library.

```go
// Keep only the elements where in returns true. Optional comment strings.
func (enum *Dee[T]) Filter(in func(value T) bool, comments ...string)

// Perform logic using each element as an input. No changes to the underlying elements are made.
// Set the first optional comment to "con" for concurrent execution of input functions.
// Concurrent execution will be slower for most use-cases, while the order in which the funcs are
// evaluated is non-deterministic. Be careful when using "con".
func (enum *Dee[T]) Foreach(in func(value T), comments ...string)

// Transform each element by applying a function. Optional comment strings.
func (enum *Dee[T]) Map(in func(value T) T, comments ...string)

// Skip the first n items and yields the rest. Comments inferred.
func (enum *Dee[T]) Skip(n int)

// Yield only the first n items from the enumerator. Comments inferred.
func (enum *Dee[T]) Take(n int)

// Interpret orders on data. Return new slice.
func (enum *Dee[T]) Apply(input []T) []T

// If your element type contains any reference fields and you want to guarantee
// that the original input is never mutated, provide a deep clone function here.
// The clone function must return a fully independent copy of the element. By
// default, all values, reference or not, are shallowly cloned.
func (enum *Dee[T]) WithDeepClone(in func(value T) T, comments ...string)
```

Usage

```go
package main

import (
    "github.com/kyleraywed/dee"
)

func main() {
    // create a Dee object of the type you want to enumerate over
    var enum dee.Dee[int]

    // Each adapter is applied in the order in which they are declared.
    // So this would happen first.
    enum.Filter(func(value int) bool {
        return value % 2 == 0
    }, "Get just the evens")

    // Second. Also notice the optional comment. String() is implemented;
    // fmt.Print()ing the object presents everything.
    enum.Map(func(value int) int {
        return value * 2
    }, "Double them")

    // Third.
    enum.Filter(func(value int) bool { // get just values > 10
        return value > 10
    })

    // Fourth. 
    enum.Foreach(func(value int) { // print the values
        fmt.Println(value)
    })

    // Last. Take and skip still log inferred comments.
    enum.Take(2) // get just the first 2 elements

    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    output := enum.Apply(numbers) // []int{12, 16}
    // [12, 16, 20] will print when Apply is run since
    // Foreach() was called before Take()
}
```
Notes and design
- 
- If the type you're enumerating over is or contains reference types, remember to
use the WithDeepClone() method. Otherwise, you may end up with unintended side-effects
on the input data. If that's of no consequence, the default shallow copy is quick.