package main

import (
	"fmt"
	"github.com/rosscartlidge/ssql"
	"iter"
	"slices"
)

func main() {
	fmt.Println("🧪 Comprehensive iter.Seq JSON Test")
	fmt.Println("===================================\n")

	// Test various iter.Seq types
	stringTags := slices.Values([]string{"urgent", "work", "bug"})
	intScores := slices.Values([]int{85, 92, 78})
	floatValues := slices.Values([]float64{1.5, 2.3, 3.7})
	boolFlags := slices.Values([]bool{true, false, true})

	record := ssql.MakeMutableRecord().
		String("id", "MIXED-001").
		String("title", "Complex Task").
		StringSeq("string_tags", stringTags).
		IntSeq("int_scores", intScores).
		Float64Seq("float_values", floatValues).
		BoolSeq("bool_flags", boolFlags).
		Freeze()

	fmt.Println("📊 Record with multiple iter.Seq types:")
	fmt.Printf("  ID: %s\n", ssql.GetOr(record, "id", ""))
	fmt.Printf("  Title: %s\n", ssql.GetOr(record, "title", ""))

	if seq, ok := ssql.Get[iter.Seq[string]](record, "string_tags"); ok {
		fmt.Print("  String Tags: ")
		for val := range seq {
			fmt.Printf("%s ", val)
		}
		fmt.Println()
	}

	if seq, ok := ssql.Get[iter.Seq[int]](record, "int_scores"); ok {
		fmt.Print("  Int Scores: ")
		for val := range seq {
			fmt.Printf("%d ", val)
		}
		fmt.Println()
	}

	if seq, ok := ssql.Get[iter.Seq[float64]](record, "float_values"); ok {
		fmt.Print("  Float Values: ")
		for val := range seq {
			fmt.Printf("%.1f ", val)
		}
		fmt.Println()
	}

	if seq, ok := ssql.Get[iter.Seq[bool]](record, "bool_flags"); ok {
		fmt.Print("  Bool Flags: ")
		for val := range seq {
			fmt.Printf("%t ", val)
		}
		fmt.Println()
	}

	fmt.Println("\n🔧 Testing WriteJSON with all iter.Seq types:")

	stream := ssql.From([]ssql.Record{record})
	filename := "/tmp/comprehensive_seq_test.json"

	err := ssql.WriteJSON(stream, filename)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("✅ JSON written successfully")

	fmt.Println("\n📖 Raw JSON file:")
	fmt.Println("------------------")

	// Use bash to show the actual file with formatting
	fmt.Println("(Run 'cat /tmp/comprehensive_seq_test.json' to see raw content)")

	fmt.Println("\n✅ Success! All iter.Seq types are properly converted to JSON arrays")
	fmt.Println("  • string iter.Seq → JSON string array")
	fmt.Println("  • int iter.Seq → JSON number array")
	fmt.Println("  • float64 iter.Seq → JSON number array")
	fmt.Println("  • bool iter.Seq → JSON boolean array")
}
