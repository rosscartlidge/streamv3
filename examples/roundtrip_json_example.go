package main

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strings"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔄 Round-trip Data Preservation Test: Stream → JSON → Stream")
	fmt.Println("============================================================\n")

	// Create complex original data with all supported types
	tags := slices.Values([]string{"urgent", "security", "critical"})
	scores := slices.Values([]int{95, 88, 92, 76})
	flags := slices.Values([]bool{true, false, true})
	weights := slices.Values([]float64{1.5, 2.3, 0.8})

	metadata := streamv3.MakeMutableRecord().
		String("priority", "high").
		Int("version", 2).
		Float("confidence", 0.95).
		Bool("verified", true).
		Freeze()

	// Create JSONString field
	configJSON, _ := streamv3.NewJSONString(map[string]any{
		"timeout": 30,
		"retries": 3,
		"enabled": true,
	})

	originalRecords := []streamv3.Record{
		streamv3.MakeMutableRecord().
			String("id", "TASK-001").
			String("title", "Security Update").
			Int("priority_num", 1).
			Float("score", 95.5).
			Bool("completed", false).
			StringSeq("tags", tags).
			IntSeq("scores", scores).
			BoolSeq("flags", flags).
			Float64Seq("weights", weights).
			Nested("metadata", metadata).
			JSONString("config", configJSON).
			Freeze(),
		streamv3.MakeMutableRecord().
			String("id", "TASK-002").
			String("title", "Feature Request").
			Int("priority_num", 2).
			Float("score", 87.2).
			Bool("completed", true).
			Freeze(),
	}

	fmt.Println("📊 Original data:")
	for i, record := range originalRecords {
		fmt.Printf("  Record %d:\n", i+1)
		printRecord(record, "    ")
	}

	fmt.Println("\n🔧 Step 1: Stream → JSON file")
	originalStream := streamv3.From(originalRecords)
	jsonFile := "/tmp/roundtrip_test.json"

	err := streamv3.WriteJSON(originalStream, jsonFile)
	if err != nil {
		fmt.Printf("❌ Error writing JSON: %v\n", err)
		return
	}
	fmt.Printf("✅ Written to %s\n", jsonFile)

	fmt.Println("\n🔧 Step 2: JSON file → Stream")
	reconstructedStream, err := streamv3.ReadJSON(jsonFile)
	if err != nil {
		fmt.Printf("❌ Error reading JSON: %v\n", err)
		return
	}

	var reconstructedRecords []streamv3.Record
	for record := range reconstructedStream {
		reconstructedRecords = append(reconstructedRecords, record)
	}

	fmt.Printf("✅ Read %d records from JSON\n", len(reconstructedRecords))

	fmt.Println("\n📊 Reconstructed data:")
	for i, record := range reconstructedRecords {
		fmt.Printf("  Record %d:\n", i+1)
		printRecord(record, "    ")
	}

	fmt.Println("\n🧪 Data Preservation Analysis:")
	fmt.Println("==============================")

	if len(originalRecords) != len(reconstructedRecords) {
		fmt.Printf("❌ Record count mismatch: %d → %d\n", len(originalRecords), len(reconstructedRecords))
		return
	}

	allMatch := true
	for i := range originalRecords {
		original := originalRecords[i]
		reconstructed := reconstructedRecords[i]

		fmt.Printf("\nRecord %d comparison:\n", i+1)
		matches := compareRecords(original, reconstructed)
		if !matches {
			allMatch = false
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	if allMatch {
		fmt.Println("✅ SUCCESS: All data preserved perfectly in round-trip!")
		fmt.Println("   • Simple types: Preserved exactly")
		fmt.Println("   • iter.Seq fields: Converted to arrays, data intact")
		fmt.Println("   • Record fields: Converted to maps, structure preserved")
		fmt.Println("   • JSONString fields: Parsed to structures, no double-encoding")
		fmt.Println("   • Metadata fields: ReadJSON adds _line_number (expected)")
	} else {
		fmt.Println("❌ FAILURE: Some data was lost or corrupted in round-trip")
	}

	fmt.Println("\n🔍 Key Insights:")
	fmt.Println("  📝 JSON format naturally preserves: strings, numbers, booleans, arrays, objects")
	fmt.Println("  🔄 iter.Seq → arrays: Content preserved, type changed (expected)")
	fmt.Println("  🏗️ Record → map[string]any: Structure preserved, type changed (expected)")
	fmt.Println("  📦 JSONString → parsed: Avoids double-encoding, preserves original structure")
	fmt.Println("  ⚠️  Type information: Lost (JSON limitation), but data integrity maintained")
}

func printRecord(record streamv3.Record, indent string) {
	for key, value := range record {
		switch v := value.(type) {
		case streamv3.JSONString:
			fmt.Printf("%s%s: %s (JSONString)\n", indent, key, v)
		case streamv3.Record:
			fmt.Printf("%s%s: {Record with %d fields}\n", indent, key, len(v))
		case iter.Seq[string]:
			fmt.Printf("%s%s: [", indent, key)
			first := true
			for item := range v {
				if !first { fmt.Print(", ") }
				fmt.Printf("%q", item)
				first = false
			}
			fmt.Println("] (iter.Seq[string])")
		case iter.Seq[int]:
			fmt.Printf("%s%s: [", indent, key)
			first := true
			for item := range v {
				if !first { fmt.Print(", ") }
				fmt.Printf("%d", item)
				first = false
			}
			fmt.Println("] (iter.Seq[int])")
		case iter.Seq[bool]:
			fmt.Printf("%s%s: [", indent, key)
			first := true
			for item := range v {
				if !first { fmt.Print(", ") }
				fmt.Printf("%t", item)
				first = false
			}
			fmt.Println("] (iter.Seq[bool])")
		case iter.Seq[float64]:
			fmt.Printf("%s%s: [", indent, key)
			first := true
			for item := range v {
				if !first { fmt.Print(", ") }
				fmt.Printf("%.1f", item)
				first = false
			}
			fmt.Println("] (iter.Seq[float64])")
		default:
			fmt.Printf("%s%s: %v (%T)\n", indent, key, value, value)
		}
	}
}

func compareRecords(original, reconstructed streamv3.Record) bool {
	matches := true

	// Check each field in original
	for key, originalValue := range original {
		reconstructedValue, exists := reconstructed[key]
		if !exists {
			fmt.Printf("  ❌ Missing field: %s\n", key)
			matches = false
			continue
		}

		fieldMatches := compareValues(key, originalValue, reconstructedValue)
		if fieldMatches {
			fmt.Printf("  ✅ %s: Preserved\n", key)
		} else {
			matches = false
		}
	}

	// Check for extra fields in reconstructed (ignore ReadJSON metadata)
	for key := range reconstructed {
		if _, exists := original[key]; !exists {
			if key == "_line_number" {
				fmt.Printf("  ℹ️  %s: Added by ReadJSON (expected)\n", key)
				// Don't count metadata fields as failures
			} else {
				fmt.Printf("  ⚠️  Extra field: %s\n", key)
				matches = false
			}
		}
	}

	return matches
}

func compareValues(fieldName string, original, reconstructed any) bool {
	// Handle JSONString specially
	if originalJSON, ok := original.(streamv3.JSONString); ok {
		// JSONString should be parsed back to its original structure
		originalParsed, err := originalJSON.Parse()
		if err != nil {
			fmt.Printf("  ❌ %s: Failed to parse original JSONString: %v\n", fieldName, err)
			return false
		}
		return reflect.DeepEqual(originalParsed, reconstructed)
	}

	// Handle iter.Seq types - they should become arrays
	if isIterSeq(original) {
		// Convert original sequence to slice for comparison
		originalSlice := materializeSequenceForComparison(original)
		return reflect.DeepEqual(originalSlice, reconstructed)
	}

	// Handle Record types - they should become map[string]any
	if originalRecord, ok := original.(streamv3.Record); ok {
		reconstructedMap, ok := reconstructed.(map[string]any)
		if !ok {
			fmt.Printf("  ❌ %s: Record not converted to map[string]any, got %T\n", fieldName, reconstructed)
			return false
		}

		// Convert Record to map[string]any for comparison
		originalMap := make(map[string]any)
		for k, v := range originalRecord {
			originalMap[k] = v
		}
		return reflect.DeepEqual(originalMap, reconstructedMap)
	}

	// Handle numeric types - JSON converts all numbers to float64
	if originalInt, ok := original.(int64); ok {
		if reconstructedFloat, ok := reconstructed.(float64); ok {
			return float64(originalInt) == reconstructedFloat
		}
	}

	// For simple types, direct comparison
	equal := reflect.DeepEqual(original, reconstructed)
	if !equal {
		fmt.Printf("  ❌ %s: Value mismatch - original: %v (%T), reconstructed: %v (%T)\n",
			fieldName, original, original, reconstructed, reconstructed)
	}
	return equal
}

func isIterSeq(value any) bool {
	switch value.(type) {
	case iter.Seq[string], iter.Seq[int], iter.Seq[bool], iter.Seq[float64]:
		return true
	default:
		return false
	}
}

func materializeSequenceForComparison(value any) []any {
	var result []any
	switch seq := value.(type) {
	case iter.Seq[string]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[int]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[bool]:
		for v := range seq { result = append(result, v) }
	case iter.Seq[float64]:
		for v := range seq { result = append(result, v) }
	}
	return result
}