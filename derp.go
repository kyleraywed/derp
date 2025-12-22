package derp

// Deferred Execution Reusable data-processing Pipeline

/*
	Notes:
		- Dynamic scheduling slows operations by roughly 165x due to a necessary mutex used by the channel.
			- ~200ms -> ~33s in examples/primes
			- Stick with static chunking
*/

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	clone "github.com/huandu/go-clone/generic"
)

type Option byte

const (
	NoCopyOpt Option = iota
	CloneOpt
	DpcOpt
	CfeOpt
	Power25Opt
	Power50Opt
	Power75Opt
)

type order struct {
	method   string
	index    int
	comments []string
}

type Derp[T any] struct {
	filterInstructs  []func(t T) bool
	mapInstructs     []func(t T) T
	foreachInstructs []func(t T)
	reduceInstruct   func(a T, v T) T
	takeCounts       []int
	skipCounts       []int

	orders []order
}

func (pipeline Derp[T]) String() string {
	var out strings.Builder

	for idx, val := range pipeline.orders {
		var prettyComments strings.Builder

		if len(val.comments) == 0 {
			prettyComments.WriteString("[ N/A ]\n")
		}

		for _, cmt := range val.comments {
			prettyComments.WriteString("[ " + cmt + " ]\n\t\t")
		}

		fmt.Fprintf(&out, "Order %v:\n\tAdapter: %v\n\tIndex: %v\n\tComments: \n\t\t%v\n",
			idx+1, val.method, val.index, prettyComments.String())
	}

	return out.String()
}

// Keep only the elements where in returns true. Optional comment strings.
func (pipeline *Derp[T]) Filter(in func(value T) bool, comments ...string) {
	pipeline.filterInstructs = append(pipeline.filterInstructs, in)
	pipeline.orders = append(pipeline.orders, order{
		method:   "filter",
		index:    len(pipeline.filterInstructs) - 1,
		comments: comments,
	})
}

// Perform logic using each element as an input. No changes to the underlying elements are made.
// Optional comment strings.
func (pipeline *Derp[T]) Foreach(in func(value T), comments ...string) {
	pipeline.foreachInstructs = append(pipeline.foreachInstructs, in)
	pipeline.orders = append(pipeline.orders, order{
		method:   "foreach",
		index:    len(pipeline.foreachInstructs) - 1,
		comments: comments,
	})
}

// Transform each value by applying a function. Optional comment strings.
func (pipeline *Derp[T]) Map(in func(value T) T, comments ...string) {
	pipeline.mapInstructs = append(pipeline.mapInstructs, in)
	pipeline.orders = append(pipeline.orders, order{
		method:   "map",
		index:    len(pipeline.mapInstructs) - 1,
		comments: comments,
	})
}

// Reduce sets a terminal operation that aggregates all elements of the pipeline into a single value.
//
// The provided function `in` is called with an accumulator and each element of the slice,
// in order. The result of each call becomes the new accumulator for the next element.
//
// Only one Reduce can be set per pipeline. It is automatically executed last
// regardless of the order in which it was added.
//
// When Apply() is run, Apply()'s output will be a []T with a single element.
func (pipeline *Derp[T]) Reduce(in func(acc T, value T) T, comments ...string) error {
	if pipeline.reduceInstruct != nil {
		return fmt.Errorf("Reduce has already been set.")
	}

	pipeline.reduceInstruct = in
	pipeline.orders = append(pipeline.orders, order{
		method:   "reduce",
		comments: comments,
	})

	return nil
}

// Skip the first n items and yield the rest. Comment inferred.
func (pipeline *Derp[T]) Skip(n int) error {
	if n < 1 {
		return fmt.Errorf("Skip(%v): No order submitted.", n)
	}

	pipeline.skipCounts = append(pipeline.skipCounts, n)
	pipeline.orders = append(pipeline.orders, order{
		method:   "skip",
		index:    len(pipeline.skipCounts) - 1,
		comments: []string{"skip(" + strconv.Itoa(n) + ")"},
	})

	return nil
}

// Yield only the first n items from the pipeline. Comment inferred.
func (pipeline *Derp[T]) Take(n int) error {
	if n < 1 {
		return fmt.Errorf("Take(%v): No order submitted.", n)
	}

	pipeline.takeCounts = append(pipeline.takeCounts, n)
	pipeline.orders = append(pipeline.orders, order{
		method:   "take",
		index:    len(pipeline.takeCounts) - 1,
		comments: []string{"take(" + strconv.Itoa(n) + ")"},
	})

	return nil
}

// Interpret orders on data. Return new slice.
// Non-pointer-cycle-safe deep-cloning by default.
//
// Options:
//   - "nocopy"; operate directly on the input backing array. Expect mutations on reference types. Default for value types.
//   - "clone"; deep-clone non pointer cycle data. Default for reference types and structs.
//   - "dpc" : "(d)eep-clone (p)ointer (c)ycles"; eg. doubly-linked lists. Implements clone.Slowly().
//   - "cfe" : "(c)oncurrent (f)or(e)ach"; function eval order is non-deterministic. Use with caution.
//   - "power-[25, 50, 75]"; throttle cpu usage to 25, 50, or 75%. Default is 100%.
func (pipeline *Derp[T]) Apply(input []T, options ...Option) ([]T, error) {
	// Ensure reduce is the last instruction in the orders
	if pipeline.reduceInstruct != nil && pipeline.orders[len(pipeline.orders)-1].method != "reduce" {
		for idx, ord := range pipeline.orders {
			if ord.method == "reduce" {
				pipeline.orders = append(pipeline.orders[:idx], pipeline.orders[idx+1:]...) // remove it where it is
				pipeline.orders = append(pipeline.orders, ord)                              // put it on the end
				break
			}
		}
	}

	hasMultipleCloneOpts := func(in []Option) bool {
		targets := []Option{NoCopyOpt, CloneOpt, DpcOpt}
		count := 0

		for _, val := range targets {
			if slices.Contains(in, val) {
				count++
			}
			if count >= 2 {
				return true
			}
		}
		return false
	}

	// check that multiple clone options aren't invoked
	if hasMultipleCloneOpts(options) {
		return nil, fmt.Errorf("error: cannot invoke multiple cloning options")
	}

	inputType := reflect.TypeOf(input[0])

	hasExplicitCloneOption := slices.Contains(options, DpcOpt) || slices.Contains(options, NoCopyOpt) || slices.Contains(options, CloneOpt)

	if !hasExplicitCloneOption {
		switch inputType.Kind() {
		case reflect.Slice, reflect.Map, reflect.Pointer, reflect.Struct:
			options = append(options, CloneOpt)
		default:
			options = append(options, NoCopyOpt)
		}
	}

	workingSlice := make([]T, len(input))

	if len(options) > 0 && slices.Contains(options, DpcOpt) {
		workingSlice = clone.Slowly(input) // for pointer cycles
	} else if len(options) > 0 && slices.Contains(options, NoCopyOpt) {
		workingSlice = input
	} else {
		workingSlice = clone.Clone(input) // regular deep clone by default
	}

	throttleMult := 1.0
	for _, opt := range options {
		switch opt {
		case Power25Opt:
			throttleMult = 0.25
		case Power50Opt:
			throttleMult = 0.5
		case Power75Opt:
			throttleMult = 0.75
		}
	}

	//log.Printf("Running at %v%% power", throttleMult*100)
	numWorkers := int(math.Ceil(float64(runtime.GOMAXPROCS(0)) * throttleMult))

	// init chunksize
	chunkSize := (len(workingSlice) + numWorkers - 1) / numWorkers

	for _, order := range pipeline.orders {
		switch order.method {
		case "filter":
			workOrder := pipeline.filterInstructs[order.index]
			results := make([][]T, numWorkers)

			var wg sync.WaitGroup
			wg.Add(numWorkers)

			for idx := range numWorkers {
				start := idx * chunkSize

				if start >= len(workingSlice) {
					wg.Done()
					continue
				}

				// If the end marker runs longer than the slice, you've reached the end.
				end := min(start+chunkSize, len(workingSlice))

				chunk := workingSlice[start:end]

				go func(idx int) {
					defer wg.Done()

					out := make([]T, 0, len(chunk))
					for _, v := range chunk {
						if workOrder(v) {
							out = append(out, v)
						}
					}
					results[idx] = out
				}(idx)
			}

			wg.Wait()

			// Flatten
			newlength := 0
			for _, r := range results {
				newlength += len(r)
			}
			//log.Printf("Flattening:\n\tOld length: %v\n\tNew length: %v\n", len(workingSlice), newlength)
			tempSlice := make([]T, 0, newlength)

			for _, r := range results {
				tempSlice = append(tempSlice, r...)
			}

			workingSlice = tempSlice

		case "foreach":
			workOrder := pipeline.foreachInstructs[order.index]

			if len(options) > 0 && slices.Contains(options, CfeOpt) {
				var wg sync.WaitGroup
				wg.Add(numWorkers)

				for idx := range numWorkers {
					start := idx * chunkSize

					if start >= len(workingSlice) {
						wg.Done()
						continue
					}

					end := min(start+chunkSize, len(workingSlice))

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
			workOrder := pipeline.mapInstructs[order.index]

			var wg sync.WaitGroup
			wg.Add(numWorkers)

			for idx := range numWorkers {
				start := idx * chunkSize

				if start >= len(workingSlice) {
					wg.Done()
					continue
				}

				end := min(start+chunkSize, len(workingSlice))

				chunk := workingSlice[start:end]

				go func(c []T) {
					defer wg.Done()
					for i := range chunk {
						c[i] = workOrder(c[i])
					}
				}(chunk)
			}

			wg.Wait()

		case "reduce":
			workOrder := pipeline.reduceInstruct

			if len(workingSlice) == 0 {
				return []T{}, nil
			}

			acc := workingSlice[0]
			for _, v := range workingSlice[1:] {
				acc = workOrder(acc, v)
			}

			workingSlice = []T{acc}

		case "skip":
			skipUntilIndex := pipeline.skipCounts[order.index]

			if skipUntilIndex > len(workingSlice) {
				workingSlice = workingSlice[:0] // skip all
			} else {
				workingSlice = workingSlice[skipUntilIndex:]
			}

		case "take":
			takeUntilIndex := pipeline.takeCounts[order.index]

			if takeUntilIndex < len(workingSlice) {
				workingSlice = workingSlice[:takeUntilIndex]
			}
		}

		// redistribute work evenly among workers after every order
		//old := chunkSize
		chunkSize = (len(workingSlice) + numWorkers - 1) / numWorkers
		//log.Printf("Redistributing work:\n\tOld chunksize: %v\n\tNew chunksize: %v", old, chunkSize)
	}

	return workingSlice, nil
}
