package derp

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	clone "github.com/huandu/go-clone/generic"
)

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

	var pipe Derp[person]

	pipe.Map(func(value person) person {
		value.tags[0] = "CHANGED"
		value.meta["one"] = 99
		return value
	})

	out, err := pipe.Apply(people)
	if err != nil {
		t.Fatalf("TestDeepClone(); error from Apply(): %v", err)
	}

	if out[0].tags[0] != "CHANGED" {
		t.Fatalf("TestDeepClone(); mutation error, no change.\nExpected: [\"CHANGED\"] Got: [%v]\n", out[0].tags[0])
	}

	if people[0].meta["one"] != 1 {
		t.Fatalf("TestDeepClone(); mutation error, original data mutated.\nExpected: [1] Got: [%v]\n", out[0].meta["one"])
	}
}

func TestDeepClonePointerCycle(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	type node struct {
		value int
		prev  *node
		next  *node
	}

	makeDLL := func(values []int) *node {
		if len(values) == 0 {
			return nil
		}
		head := &node{value: values[0]}
		prev := head
		for _, v := range values[1:] {
			n := &node{value: v, prev: prev}
			prev.next = n
			prev = n
		}
		return head
	}

	printList := func(head *node) string {
		var sb strings.Builder
		for n := head; n != nil; n = n.next {
			prev := "nil"
			if n.prev != nil {
				prev = fmt.Sprintf("%d", n.prev.value)
			}
			next := "nil"
			if n.next != nil {
				next = fmt.Sprintf("%d", n.next.value)
			}
			fmt.Fprintf(&sb, "Node %d: prev=%s, next=%s\n", n.value, prev, next)
		}
		return sb.String()
	}

	expected := `Node 1: prev=nil, next=2
Node 2: prev=1, next=3
Node 3: prev=2, next=4
Node 4: prev=3, next=5
Node 5: prev=4, next=6
Node 6: prev=5, next=7
Node 7: prev=6, next=8
Node 8: prev=7, next=9
Node 9: prev=8, next=10
Node 10: prev=9, next=nil`
	head := makeDLL(numbers)
	cloned := clone.Slowly(head)
	gotten := printList(cloned)

	if strings.Trim(expected, "\n") != strings.Trim(gotten, "\n") {
		t.Errorf("TestDeepClonePointerCycle(); failed deep-cloning pointer cycle")
	}
}

func TestFilter(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var pipe Derp[int]

	pipe.Filter(func(value int) bool {
		return value%2 == 0 // return evens
	})

	expected := []int{2, 4, 6, 8, 10}
	gotten, err := pipe.Apply(numbers)

	if err != nil {
		t.Errorf("TestFilter() error from Apply(): %v", err)
	}

	if len(expected) != len(gotten) {
		t.Error("TestFilter(); length inequality error")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("TestFilter(); value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

// Testing is the only reason for writing code like this.
func TestForeach(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var pipe Derp[int]

	expected := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	var gotten []string

	pipe.Foreach(func(value int) {
		gotten = append(gotten, strconv.Itoa(value))
	})

	pipe.Apply(numbers)

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("TestForeach(); value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestForeachFast(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var pipe Derp[int]

	expected := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	var gotten []string

	var mu sync.Mutex

	pipe.Foreach(func(value int) { // Again, don't do this in production.
		mu.Lock()
		gotten = append(gotten, strconv.Itoa(value))
		mu.Unlock()
	})

	_, err := pipe.Apply(numbers, CfeOpt)
	if err != nil {
		t.Errorf("TestForeachFast(); error from Apply(): %v", err)
	}

	slices.SortFunc(gotten, func(a, b string) int {
		ai, _ := strconv.Atoi(a)
		bi, _ := strconv.Atoi(b)
		return ai - bi
	})

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("TestForeachFast(); value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestForeachMut(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var pipe Derp[int]

	pipe.Foreach(func(value int) {
		value = value * 2
	})

	out, err := pipe.Apply(numbers)
	if err != nil {
		t.Errorf("TestForeachMut(); error from Apply(): %v", err)
	}

	if !slices.Equal(numbers, out) {
		t.Errorf("TestForeachMut(); output mutated. Expected equality with input.")
	}
}

func TestMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var pipe Derp[int]

	pipe.Map(func(value int) int {
		return value * value // square the numbers
	})

	expected := []int{1, 4, 9, 16, 25, 36, 49, 64, 81, 100}
	gotten, err := pipe.Apply(numbers)

	if err != nil {
		t.Errorf("TestMap(); error from Apply(): %v", err)
	}

	if len(expected) != len(gotten) {
		t.Error("TestMap(); length inequality error")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("TestMap(); value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestOrder(t *testing.T) {
	var pipe Derp[int]

	pipe.Filter(func(value int) bool {
		return value%2 == 0
	}, "Foo")

	pipe.Map(func(value int) int {
		return value * 2
	}, "Bar")

	pipe.Take(3)

	pipe.Skip(1)

	pipe.Map(func(value int) int {
		return value + 1
	}, "baz")

	pipe.Filter(func(value int) bool {
		return value%2 != 0
	}, "boo")

	pipe.Take(3)

	expected := []order{
		{method: "filter", index: 0, comments: []string{"Foo"}},
		{method: "map", index: 0, comments: []string{"Bar"}},
		{method: "take", index: 0, comments: []string{"take(3)"}},
		{method: "skip", index: 0, comments: []string{"skip(1)"}},
		{method: "map", index: 1, comments: []string{"baz"}},
		{method: "filter", index: 1, comments: []string{"boo"}},
		{method: "take", index: 1, comments: []string{"take(3)"}},
	}

	if len(pipe.orders) != len(expected) {
		t.Error("TestOrder(); length inequality error")
	}

	for idx, val := range expected {
		if pipe.orders[idx].method != val.method {
			t.Errorf("TestOrder(); order adapter mismatch.\nExpected: [%v] Got: [%v]\n", val.method, pipe.orders[idx].method)
		}
		if pipe.orders[idx].index != val.index {
			t.Errorf("TestOrder(); order index mismatch.\nExpected: [%v] Got: [%v]\n", val.index, pipe.orders[idx].index)
		}
		if pipe.orders[idx].comments[0] != val.comments[0] {
			t.Errorf("TestOrder(); order comment mismatch.\nExpected: [%v] Got: [%v]\n", val.comments, pipe.orders[idx].comments[0])
		}
	}
}

func TestReduce(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var pipe Derp[int]

	pipe.Reduce(func(acc, value int) int {
		return acc + value
	})

	out, err := pipe.Apply(numbers)
	if err != nil {
		t.Errorf("TestReduce(); error from Apply(): %v", err)
	}

	if out[0] != 55 {
		t.Errorf("TestReduce(); value inequality.\nExpected [55] Got: [%v]\n", out)
	}
}

func TestReduceConcurrent(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var pipe Derp[int]

	pipe.Reduce(func(acc, value int) int {
		return acc + value
	})

	out, err := pipe.Apply(numbers)
	if err != nil {
		t.Errorf("TestReduce(); error from Apply(): %v", err)
	}

	if out[0] != 55 {
		t.Errorf("TestReduce(); value inequality.\nExpected [55] Got: [%v]\n", out)
	}
}

func TestSkip(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var halfPipe Derp[int]

	if err := halfPipe.Skip(5); err != nil {
		log.Println(err)
	}

	expected := []int{6, 7, 8, 9, 10}
	gotten, err := halfPipe.Apply(numbers)

	if err != nil {
		t.Errorf("TestSkip() error from Apply(): %v", err)
	}

	if len(expected) != len(gotten) {
		t.Error("TestSkip(); length inequality error")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("TestSkip(); value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestTake(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var halfPipe Derp[int]

	if err := halfPipe.Take(5); err != nil {
		log.Println(err)
	}

	expected := []int{1, 2, 3, 4, 5}
	gotten, err := halfPipe.Apply(numbers)

	if err != nil {
		t.Errorf("TestTake() error from Apply(): %v", err)
	}

	if len(expected) != len(gotten) {
		t.Error("TestTake(); length inequality error")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("TestTake(); value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}
