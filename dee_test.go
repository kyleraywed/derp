package dei

import (
	"maps"
	"slices"
	"strconv"
	"sync"
	"testing"
)

func TestOrder(t *testing.T) {
	var iter Dei[int]

	iter.Filter(func(value int) bool {
		return value%2 == 0
	}, "Foo")

	iter.Map(func(value int) int {
		return value * 2
	}, "Bar")

	iter.Take(3)

	iter.Skip(1)

	iter.Map(func(value int) int {
		return value + 1
	}, "baz")

	iter.Filter(func(value int) bool {
		return value%2 != 0
	}, "boo")

	iter.Take(3)

	expected := []order{
		{method: "filter", index: 0, comments: []string{"Foo"}},
		{method: "map", index: 0, comments: []string{"Bar"}},
		{method: "take", index: 0, comments: []string{"3"}},
		{method: "skip", index: 0, comments: []string{"1"}},
		{method: "map", index: 1, comments: []string{"baz"}},
		{method: "filter", index: 1, comments: []string{"boo"}},
		{method: "take", index: 1, comments: []string{"3"}},
	}

	if len(iter.orders) != len(expected) {
		t.Error("Order len mismatch")
	}

	for idx, val := range expected {
		if iter.orders[idx].method != val.method {
			t.Errorf("Order adapter mismatch.\nExpected: [%v] Got: [%v]\n", val.method, iter.orders[idx].method)
		}
		if iter.orders[idx].index != val.index {
			t.Errorf("Order index mismatch.\nExpected: [%v] Got: [%v]\n", val.index, iter.orders[idx].index)
		}
		if iter.orders[idx].comments[0] != val.comments[0] {
			t.Errorf("Order comment mismatch.\nExpected: [%v] Got: [%v]\n", val.comments, iter.orders[idx].comments[0])
		}
	}
}

func TestFilter(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	iter.Filter(func(value int) bool {
		return value%2 == 0 // return evens
	})

	expected := []int{2, 4, 6, 8, 10}
	gotten := iter.Apply(numbers)

	if len(expected) != len(gotten) {
		t.Error("Filter len mismatch")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Filter value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	iter.Map(func(value int) int {
		return value * value // square the numbers
	})

	expected := []int{1, 4, 9, 16, 25, 36, 49, 64, 81, 100}
	gotten := iter.Apply(numbers)

	if len(expected) != len(gotten) {
		t.Error("Map len mismatch")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Map value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestTake(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 19}
	var iter Dei[int]

	iter.Take(5)

	expected := []int{1, 2, 3, 4, 5}
	gotten := iter.Apply(numbers)

	if len(expected) != len(gotten) {
		t.Error("Take len mismatch")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Take value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestSkip(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	iter.Skip(5)

	expected := []int{6, 7, 8, 9, 10}
	gotten := iter.Apply(numbers)

	if len(expected) != len(gotten) {
		t.Error("Skip len mismatch")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Skip value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestForeach(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	expected := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	var gotten []string

	iter.Foreach(func(value int) {
		gotten = append(gotten, strconv.Itoa(value))
	})

	iter.Apply(numbers)

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Foreach value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestForeachFast(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	expected := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	var gotten []string

	var mu sync.Mutex

	iter.Foreach(func(value int) { // Don't do this in production, kids.
		mu.Lock()
		gotten = append(gotten, strconv.Itoa(value))
		mu.Unlock()
	}, "con")

	iter.Apply(numbers)

	slices.SortFunc(gotten, func(a, b string) int {
		ai, _ := strconv.Atoi(a)
		bi, _ := strconv.Atoi(b)
		return ai - bi
	})

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Foreach value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestDeepClone(t *testing.T) {
	type person struct {
		name string
		tags []string
		meta map[string]int
	}

	p1 := person{
		name: "kyle",
		tags: []string{"x", "y", "z"},
		meta: map[string]int{
			"one": 1,
		},
	}

	people := []person{p1}

	var iter Dei[person]

	iter.WithDeepClone(func(value person) person {
		out := value
		// strings are value types, no need enforce deep clone for name
		out.tags = slices.Clone(value.tags)
		out.meta = maps.Clone(value.meta)

		return out
	})

	iter.Map(func(value person) person {
		value.tags[0] = "CHANGED"
		value.meta["one"] = 99
		return value
	})

	out := iter.Apply(people)

	if out[0].tags[0] != "CHANGED" {
		t.Fatalf("Deep Clone mutation error, no change.\nExpected: [\"CHANGED\"] Got: [%v]\n", out[0].tags[0])
	}

	if people[0].meta["one"] != 1 {
		t.Fatalf("Deep Clone mutation error, original data mutated.\nExpected: [1] Got: [%v]\n", out[0].meta["one"])
	}
}
