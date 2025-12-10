package dee

/*
	REMEMBER:
		Keep the methods in alphabetical order with Apply() & String() at the bottom.
		This applies to the switch in Apply() as well. Keep them in order.
*/

import (
	"fmt"
	"log"
	"runtime"
	"slices"
	"strconv"
	"sync"
)

type order struct {
	method   string
	index    int
	comments []string
}

type Dee[T any] struct {
	filters    []func(t T) bool
	mappers    []func(t T) T
	foreachers []func(t T)

	takeCounts []int
	skipCounts []int

	orders []order

	userDeepClone func(t T) T
}

// Keep only the elements where in returns true. Optional comment strings.
func (iter *Dee[T]) Filter(in func(value T) bool, comments ...string) {
	iter.filters = append(iter.filters, in)
	iter.orders = append(iter.orders, order{
		method: "filter", index: len(iter.filters) - 1, comments: comments,
	})
}

// Perform logic using each element as an input. No changes to the underlying elements are made.
// Set the first optional comment to "con" for concurrent execution of input functions.
// Concurrent execution will be slower for most use-cases, while the order in which the funcs are
// evaluated is non-deterministic. Be careful when using "con".
func (iter *Dee[T]) Foreach(in func(value T), comments ...string) {
	iter.foreachers = append(iter.foreachers, in)
	iter.orders = append(iter.orders, order{
		method: "foreach", index: len(iter.foreachers) - 1, comments: comments,
	})
}

// Transform each element by applying a function. Optional comment strings.
func (iter *Dee[T]) Map(in func(value T) T, comments ...string) {
	iter.mappers = append(iter.mappers, in)
	iter.orders = append(iter.orders, order{
		method: "map", index: len(iter.mappers) - 1, comments: comments,
	})
}

// Skip the first n items and yields the rest. Comments inferred.
func (iter *Dee[T]) Skip(n int) {
	if n < 1 {
		log.Printf("Skip(%v): No order submitted.", n)
		return
	}

	iter.skipCounts = append(iter.skipCounts, n)
	iter.orders = append(iter.orders, order{
		method: "skip", index: len(iter.skipCounts) - 1, comments: []string{strconv.Itoa(n)},
	})
}

// Yield only the first n items from the iterator. Comments inferred.
func (iter *Dee[T]) Take(n int) {
	if n < 1 {
		log.Printf("Take(%v): No order submitted.", n)
		return
	}

	iter.takeCounts = append(iter.takeCounts, n)
	iter.orders = append(iter.orders, order{
		method: "take", index: len(iter.takeCounts) - 1, comments: []string{strconv.Itoa(n)},
	})
}

// Interpret orders on data. Return new slice.
func (iter *Dee[T]) Apply(input []T) []T {
	workingSlice := make([]T, len(input))
	if iter.userDeepClone != nil {
		for i := range input {
			workingSlice[i] = iter.userDeepClone(input[i])
		}
	} else {
		workingSlice = slices.Clone(input) // shallow copy
	}

	numWorkers := runtime.NumCPU()
	chunkSize := (len(workingSlice) + numWorkers - 1) / numWorkers

	for _, order := range iter.orders {
		switch order.method {
		case "filter":
			workOrder := iter.filters[order.index]
			results := make([][]T, numWorkers)

			var wg sync.WaitGroup
			wg.Add(numWorkers)

			for idx := range numWorkers {
				start := idx * chunkSize

				if start >= len(workingSlice) {
					wg.Done()
					continue
				}

				end := start + chunkSize
				if end > len(workingSlice) {
					end = len(workingSlice)
				}

				chunk := workingSlice[start:end]

				go func() {
					defer wg.Done()

					out := make([]T, 0, len(chunk))
					for _, v := range chunk {
						if workOrder(v) {
							out = append(out, v)
						}
					}
					results[idx] = out
				}()
			}

			wg.Wait()

			// Flatten
			tempSlice := make([]T, 0, len(workingSlice))
			for _, r := range results {
				tempSlice = append(tempSlice, r...)
			}

			workingSlice = tempSlice

		case "foreach":
			workOrder := iter.foreachers[order.index]

			if len(order.comments) > 0 && order.comments[0] == "con" {
				var wg sync.WaitGroup
				wg.Add(numWorkers)

				for idx := range numWorkers {
					start := idx * chunkSize

					if start >= len(workingSlice) {
						wg.Done()
						continue
					}

					end := start + chunkSize
					if end > len(workingSlice) {
						end = len(workingSlice)
					}

					chunk := workingSlice[start:end]

					go func() {
						defer wg.Done()

						for _, v := range chunk {
							workOrder(v)
						}
					}()
				}

				wg.Wait()

			} else {
				for _, val := range workingSlice {
					workOrder(val)
				}
			}

		case "map":
			workOrder := iter.mappers[order.index]

			var wg sync.WaitGroup
			wg.Add(numWorkers)

			for idx := range numWorkers {
				start := idx * chunkSize

				if start >= len(workingSlice) {
					wg.Done()
					continue
				}

				end := start + chunkSize
				if end > len(workingSlice) {
					end = len(workingSlice)
				}

				chunk := workingSlice[start:end]

				go func() {
					defer wg.Done()
					for i := range chunk {
						chunk[i] = workOrder(chunk[i])
					}
				}()
			}

			wg.Wait()

		case "skip":
			skipUntilIndex := iter.skipCounts[order.index] - 1

			if skipUntilIndex > len(workingSlice)-1 {
				log.Printf("index %v out of range. skipping order...", skipUntilIndex)
				continue
			}

			workingSlice = workingSlice[skipUntilIndex+1:]

		case "take":
			takeUntilIndex := iter.takeCounts[order.index] - 1

			if takeUntilIndex > len(workingSlice)-1 {
				log.Printf("index %v out of range, skipping order...", takeUntilIndex)
				continue
			}

			workingSlice = workingSlice[:takeUntilIndex+1]
		}
	}

	return workingSlice
}

// If your element type contains any reference fields and you want to guarantee
// that the original input is never mutated, provide a deep clone function here.
// The clone function must return a fully independent copy of the element. By
// default, all values, reference or not, are shallowly cloned.
func (iter *Dee[T]) WithDeepClone(in func(value T) T, comments ...string) {
	iter.userDeepClone = in
}

func (iter Dee[T]) String() string {
	var out string

	out += fmt.Sprintf(
		"Deep clone implemented: %v\n", iter.userDeepClone != nil,
	)

	for idx, val := range iter.orders {
		var prettyComments string

		if len(val.comments) == 0 {
			prettyComments += "[ N/A ]\n"
		}

		for _, cmt := range val.comments {
			prettyComments += "[ " + cmt + " ]\n\t\t"
		}

		out += fmt.Sprintf(
			"Order %v:\n\tAdapter: %v\n\tIndex: %v\n\tComments: \n\t\t%v\n",
			idx+1, val.method, val.index, prettyComments,
		)
	}

	return out
}
