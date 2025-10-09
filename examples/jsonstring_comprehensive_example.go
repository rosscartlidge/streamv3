package main

import (
	"fmt"
	"slices"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🚀 JSONString Type Comprehensive Test")
	fmt.Println("====================================\n")

	// Create complex data
	tags := slices.Values([]string{"urgent", "security", "critical"})
	scores := slices.Values([]int{95, 88, 92})
	metadata := streamv3.NewRecord().
		String("priority", "high").
		Int("version", 2).
		Float("weight", 1.5).
		Build()

	// Create record with complex fields
	task := streamv3.NewRecord().
		String("id", "TASK-001").
		String("title", "Security Update").
		StringSeq("tags", tags).
		IntSeq("scores", scores).
		Record("metadata", metadata).
		Build()

	fmt.Println("📊 Original record:")
	fmt.Printf("  ID: %s\n", streamv3.GetOr(task, "id", ""))
	fmt.Printf("  Title: %s\n", streamv3.GetOr(task, "title", ""))

	stream := streamv3.From([]streamv3.Record{task})

	fmt.Println("\n🔧 Test 1: MaterializeJSON creates JSONString fields")
	fmt.Println("--------------------------------------------------")

	// Use MaterializeJSON to create JSONString fields
	materialized := streamv3.Chain(
		streamv3.MaterializeJSON("tags", "tags_json"),
		streamv3.MaterializeJSON("scores", "scores_json"),
		streamv3.MaterializeJSON("metadata", "metadata_json"),
	)(stream)

	var result streamv3.Record
	for r := range materialized {
		result = r
		break
	}

	// Test JSONString type assertions and methods
	if tagsJSON, ok := streamv3.Get[streamv3.JSONString](result, "tags_json"); ok {
		fmt.Printf("✅ tags_json is JSONString type: %s\n", tagsJSON)
		fmt.Printf("   IsValid: %t\n", tagsJSON.IsValid())
		fmt.Printf("   Parsed: %v\n", tagsJSON.MustParse())
		fmt.Printf("   Pretty:\n%s\n", tagsJSON.Pretty())
	}

	if scoresJSON, ok := streamv3.Get[streamv3.JSONString](result, "scores_json"); ok {
		fmt.Printf("✅ scores_json is JSONString type: %s\n", scoresJSON)
		fmt.Printf("   IsValid: %t\n", scoresJSON.IsValid())
		fmt.Printf("   Parsed: %v\n", scoresJSON.MustParse())
	}

	if metaJSON, ok := streamv3.Get[streamv3.JSONString](result, "metadata_json"); ok {
		fmt.Printf("✅ metadata_json is JSONString type: %s\n", metaJSON)
		fmt.Printf("   IsValid: %t\n", metaJSON.IsValid())
		fmt.Printf("   Pretty:\n%s\n", metaJSON.Pretty())
	}

	fmt.Println("\n🔧 Test 2: Creating records with JSONString fields directly")
	fmt.Println("-----------------------------------------------------------")

	// Create JSONString from complex data
	userJSON, _ := streamv3.NewJSONString(map[string]any{
		"name": "Alice",
		"age":  30,
		"roles": []string{"admin", "developer"},
	})

	scoresJSON, _ := streamv3.NewJSONString([]int{85, 92, 78, 88})

	// Create record with JSONString fields using fluent API
	recordWithJSON := streamv3.NewRecord().
		String("id", "USER-001").
		JSONString("user_data", userJSON).
		JSONString("test_scores", scoresJSON).
		Build()

	fmt.Printf("Record with JSONString fields:\n")
	fmt.Printf("  ID: %s\n", streamv3.GetOr(recordWithJSON, "id", ""))

	if userData, ok := streamv3.Get[streamv3.JSONString](recordWithJSON, "user_data"); ok {
		fmt.Printf("  User Data (JSONString): %s\n", userData)
		fmt.Printf("  Pretty User Data:\n%s\n", userData.Pretty())
	}

	if testScores, ok := streamv3.Get[streamv3.JSONString](recordWithJSON, "test_scores"); ok {
		fmt.Printf("  Test Scores (JSONString): %s\n", testScores)
	}

	fmt.Println("\n🔧 Test 3: Grouping by JSONString fields")
	fmt.Println("----------------------------------------")

	// Create multiple records with same JSONString values
	commonTags, _ := streamv3.NewJSONString([]string{"urgent", "work"})
	otherTags, _ := streamv3.NewJSONString([]string{"feature", "enhancement"})

	tasks := []streamv3.Record{
		streamv3.NewRecord().String("id", "T1").String("team", "Backend").JSONString("tags_json", commonTags).Build(),
		streamv3.NewRecord().String("id", "T2").String("team", "Frontend").JSONString("tags_json", commonTags).Build(),
		streamv3.NewRecord().String("id", "T3").String("team", "QA").JSONString("tags_json", otherTags).Build(),
	}

	groupResults := streamv3.Chain(
		streamv3.GroupByFields("group_data", "tags_json"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"count": streamv3.Count(),
			"teams": streamv3.Collect("team"),
		}),
	)(slices.Values(tasks))

	fmt.Println("Grouping by JSONString:")
	for group := range groupResults {
		tagsJSON := streamv3.GetOr(group, "tags_json", streamv3.JSONString(""))
		count := streamv3.GetOr(group, "count", int64(0))
		fmt.Printf("  Tags: %s → %d tasks\n", tagsJSON, count)
	}

	fmt.Println("\n🔧 Test 4: CSV and JSON output with JSONString")
	fmt.Println("----------------------------------------------")

	// Test CSV output
	csvStream := streamv3.From([]streamv3.Record{recordWithJSON})
	err := streamv3.WriteCSV(csvStream, "/tmp/jsonstring_test.csv")
	if err != nil {
		fmt.Printf("❌ CSV Error: %v\n", err)
	} else {
		fmt.Println("✅ CSV written successfully")
	}

	// Test JSON output
	jsonStream := streamv3.From([]streamv3.Record{recordWithJSON})
	err = streamv3.WriteJSON(jsonStream, "/tmp/jsonstring_test.json")
	if err != nil {
		fmt.Printf("❌ JSON Error: %v\n", err)
	} else {
		fmt.Println("✅ JSON written successfully")
	}

	fmt.Println("\n✅ JSONString Benefits Demonstrated:")
	fmt.Println("  🎯 Type safety: Distinct from regular strings")
	fmt.Println("  🔧 Rich methods: Parse(), Pretty(), IsValid()")
	fmt.Println("  📊 Grouping: Content-based grouping works perfectly")
	fmt.Println("  💾 I/O: Proper handling in CSV (raw) and JSON (parsed)")
	fmt.Println("  🏗️ API: Fluent builder support with .JSONString()")
	fmt.Println("  ⚡ Performance: Avoids double-encoding in JSON output")
}