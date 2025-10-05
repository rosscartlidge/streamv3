package main

import (
	"fmt"
	"iter"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("⚠️  iter.Seq Consumption Issue Demonstration")
	fmt.Println("===========================================\n")

	// Create a sequence that gets consumed when used
	numbers := func(yield func(int) bool) {
		for i := 1; i <= 5; i++ {
			if !yield(i) {
				return
			}
		}
	}

	// Store it in a record
	record := streamv3.NewRecord().
		String("id", "test").
		IntSeq("numbers", numbers).
		Build()

	fmt.Println("🔍 First access to sequence:")
	if numbersSeq, ok := streamv3.Get[iter.Seq[int]](record, "numbers"); ok {
		fmt.Print("Numbers: ")
		for num := range numbersSeq {
			fmt.Printf("%d ", num)
		}
		fmt.Println()
	}

	fmt.Println("\n🔍 Second access to same sequence:")
	if numbersSeq, ok := streamv3.Get[iter.Seq[int]](record, "numbers"); ok {
		fmt.Print("Numbers: ")
		count := 0
		for num := range numbersSeq {
			fmt.Printf("%d ", num)
			count++
		}
		if count == 0 {
			fmt.Print("(empty - sequence was consumed!)")
		}
		fmt.Println()
	}

	fmt.Println("\n❌ Problem: iter.Seq can only be consumed once!")
	fmt.Println("💡 Solution: Use SeqFactory pattern for reusable sequences")
}