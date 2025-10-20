package main

import (
	"fmt"
	"iter"
	"slices"
	"strings"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔄 DotFlatten vs CrossFlatten Comparison")
	fmt.Println("=======================================\n")

	// Example 1: Same input data for both operations
	fmt.Println("📊 Example 1: Product Configuration Expansion")
	fmt.Println("--------------------------------------------")

	colors := slices.Values([]string{"red", "blue"})
	sizes := slices.Values([]string{"small", "large"})

	product := streamv3.MakeMutableRecord().
		String("name", "T-Shirt").
		Float("base_price", 19.99).
		StringSeq("colors", colors).
		StringSeq("sizes", sizes).
		Freeze()

	fmt.Printf("Input: %s (base price: $%.2f)\n",
		streamv3.GetOr(product, "name", ""),
		streamv3.GetOr(product, "base_price", 0.0))

	// Show input sequences
	if colorsSeq, ok := streamv3.Get[iter.Seq[string]](product, "colors"); ok {
		fmt.Print("  Colors: ")
		for color := range colorsSeq { fmt.Printf("%s ", color) }
		fmt.Println()
	}

	if sizesSeq, ok := streamv3.Get[iter.Seq[string]](product, "sizes"); ok {
		fmt.Print("  Sizes: ")
		for size := range sizesSeq { fmt.Printf("%s ", size) }
		fmt.Println()
	}

	fmt.Println("\n🔹 DotFlatten Result (Dot Product - Pairs corresponding elements):")
	dotResults := streamv3.DotFlatten(".", "colors", "sizes")(slices.Values([]streamv3.Record{product}))

	count := 1
	for result := range dotResults {
		name := streamv3.GetOr(result, "name", "")
		color := streamv3.GetOr(result, "colors", "")
		size := streamv3.GetOr(result, "sizes", "")
		price := streamv3.GetOr(result, "base_price", 0.0)
		fmt.Printf("  %d. %s %s %s - $%.2f\n", count, color, size, name, price)
		count++
	}

	fmt.Println("\n🔸 CrossFlatten Result (Cartesian Product - All combinations):")
	crossResults := streamv3.CrossFlatten(".", "colors", "sizes")(slices.Values([]streamv3.Record{product}))

	count = 1
	for result := range crossResults {
		name := streamv3.GetOr(result, "name", "")
		color := streamv3.GetOr(result, "colors", "")
		size := streamv3.GetOr(result, "sizes", "")
		price := streamv3.GetOr(result, "base_price", 0.0)
		fmt.Printf("  %d. %s %s %s - $%.2f\n", count, color, size, name, price)
		count++
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// Example 2: Different sequence lengths to show dot product behavior
	fmt.Println("\n📊 Example 2: Different Sequence Lengths")
	fmt.Println("----------------------------------------")

	shortTags := slices.Values([]string{"urgent", "work"})
	longScores := slices.Values([]int{85, 92, 78, 95}) // Longer sequence

	task := streamv3.MakeMutableRecord().
		String("id", "TASK-456").
		String("assignee", "Alice").
		StringSeq("tags", shortTags).      // 2 elements
		IntSeq("scores", longScores).      // 4 elements
		Freeze()

	fmt.Printf("Input: %s (assignee: %s)\n",
		streamv3.GetOr(task, "id", ""),
		streamv3.GetOr(task, "assignee", ""))

	// Show input sequences with lengths
	if tagsSeq, ok := streamv3.Get[iter.Seq[string]](task, "tags"); ok {
		fmt.Print("  Tags (2 items): ")
		for tag := range tagsSeq { fmt.Printf("%s ", tag) }
		fmt.Println()
	}

	if scoresSeq, ok := streamv3.Get[iter.Seq[int]](task, "scores"); ok {
		fmt.Print("  Scores (4 items): ")
		for score := range scoresSeq { fmt.Printf("%d ", score) }
		fmt.Println()
	}

	fmt.Println("\n🔹 DotFlatten Result (Uses minimum length - discards excess):")
	dotResults2 := streamv3.DotFlatten(".", "tags", "scores")(slices.Values([]streamv3.Record{task}))

	count = 1
	for result := range dotResults2 {
		id := streamv3.GetOr(result, "id", "")
		tag := streamv3.GetOr(result, "tags", "")
		score := streamv3.GetOr(result, "scores", 0)
		assignee := streamv3.GetOr(result, "assignee", "")
		fmt.Printf("  %d. %s - Tag: %s, Score: %d (assignee: %s)\n", count, id, tag, score, assignee)
		count++
	}

	fmt.Println("\n🔸 CrossFlatten Result (All combinations - creates 2×4=8 records):")
	crossResults2 := streamv3.CrossFlatten(".", "tags", "scores")(slices.Values([]streamv3.Record{task}))

	count = 1
	for result := range crossResults2 {
		id := streamv3.GetOr(result, "id", "")
		tag := streamv3.GetOr(result, "tags", "")
		score := streamv3.GetOr(result, "scores", 0)
		assignee := streamv3.GetOr(result, "assignee", "")
		fmt.Printf("  %d. %s - Tag: %s, Score: %d (assignee: %s)\n", count, id, tag, score, assignee)
		count++
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// Example 3: Nested records with sequences
	fmt.Println("\n📊 Example 3: Nested Records + Sequences")
	fmt.Println("----------------------------------------")

	userInfo := streamv3.MakeMutableRecord().
		String("name", "Bob").
		String("department", "Engineering").
		Freeze()

	permissions := slices.Values([]string{"read", "write", "admin"})

	userRecord := streamv3.MakeMutableRecord().
		String("user_id", "USR-789").
		Nested("profile", userInfo).          // Nested record
		StringSeq("permissions", permissions). // Sequence
		Freeze()

	fmt.Printf("Input: %s\n", streamv3.GetOr(userRecord, "user_id", ""))

	if profile, ok := streamv3.Get[streamv3.Record](userRecord, "profile"); ok {
		name := streamv3.GetOr(profile, "name", "")
		dept := streamv3.GetOr(profile, "department", "")
		fmt.Printf("  Profile: %s (%s)\n", name, dept)
	}

	if permSeq, ok := streamv3.Get[iter.Seq[string]](userRecord, "permissions"); ok {
		fmt.Print("  Permissions: ")
		for perm := range permSeq { fmt.Printf("%s ", perm) }
		fmt.Println()
	}

	fmt.Println("\n🔹 DotFlatten Result (Flattens nested record + expands sequence):")
	dotResults3 := streamv3.DotFlatten(".")(slices.Values([]streamv3.Record{userRecord}))

	count = 1
	for result := range dotResults3 {
		fmt.Printf("  Record %d fields:\n", count)
		for key, value := range result {
			fmt.Printf("    %s: %v\n", key, value)
		}
		count++
		if count > 3 { break } // Limit output
	}

	fmt.Println("\n💡 Key Differences:")
	fmt.Println("  🔹 DotFlatten: Linear expansion, pairs corresponding elements")
	fmt.Println("  🔸 CrossFlatten: Exponential expansion, all possible combinations")
	fmt.Println("  📏 DotFlatten uses minimum sequence length (discards excess)")
	fmt.Println("  📈 CrossFlatten creates N×M×... records for sequences of lengths N, M, ...")
	fmt.Println("  🎯 Both support field-specific expansion and nested record flattening")
}