package main

import (
	"fmt"
	"iter"
	"slices"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🧪 MaterializeJSON vs Materialize Comparison")
	fmt.Println("============================================\n")

	// Create records with various complex field types
	tags1 := slices.Values([]string{"urgent", "work"})
	tags2 := slices.Values([]string{"feature", "enhancement"})
	numbers := slices.Values([]int{85, 92, 78})
	bools := slices.Values([]bool{true, false})

	// Create nested records
	user1 := streamv3.NewRecord().String("name", "Alice").Int("age", 30).Build()
	user2 := streamv3.NewRecord().String("name", "Bob").Int("age", 25).Build()

	tasks := []streamv3.Record{
		streamv3.NewRecord().
			String("id", "TASK-001").
			String("team", "Backend").
			StringSeq("tags", tags1).
			IntSeq("scores", numbers).
			BoolSeq("flags", bools).
			Record("user", user1).
			Build(),
		streamv3.NewRecord().
			String("id", "TASK-002").
			String("team", "Frontend").
			StringSeq("tags", tags2).
			Record("user", user2).
			Build(),
	}

	fmt.Println("📊 Original records:")
	for i, task := range tasks {
		id := streamv3.GetOr(task, "id", "")
		team := streamv3.GetOr(task, "team", "")
		fmt.Printf("  %d. %s (%s team)\n", i+1, id, team)

		if user, ok := streamv3.Get[streamv3.Record](task, "user"); ok {
			name := streamv3.GetOr(user, "name", "")
			age := streamv3.GetOr(user, "age", 0)
			fmt.Printf("     User: %s (age %d)\n", name, age)
		}

		if tagsSeq, ok := streamv3.Get[iter.Seq[string]](task, "tags"); ok {
			fmt.Print("     Tags: ")
			for tag := range tagsSeq {
				fmt.Printf("%s ", tag)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n🔧 Test 1: Old Materialize() with string sequences (comma-separated)")
	fmt.Println("---------------------------------------------------------------------")

	oldResults := streamv3.Chain(
		streamv3.Materialize("tags", "tags_string", ","),
		streamv3.GroupByFields("group_data", "tags_string"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"count": streamv3.Count(),
			"teams": streamv3.Collect("team"),
		}),
	)(slices.Values(tasks))

	fmt.Println("Old Materialize results:")
	for result := range oldResults {
		tagsString := streamv3.GetOr(result, "tags_string", "")
		count := streamv3.GetOr(result, "count", int64(0))
		fmt.Printf("  Tags: \"%s\" → %d tasks\n", tagsString, count)
	}

	fmt.Println("\n🆕 Test 2: New MaterializeJSON() with string sequences (JSON arrays)")
	fmt.Println("--------------------------------------------------------------------")

	newResults := streamv3.Chain(
		streamv3.MaterializeJSON("tags", "tags_json"),
		streamv3.GroupByFields("group_data", "tags_json"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"count": streamv3.Count(),
			"teams": streamv3.Collect("team"),
		}),
	)(slices.Values(tasks))

	fmt.Println("MaterializeJSON results:")
	for result := range newResults {
		tagsJSON := streamv3.GetOr(result, "tags_json", "")
		count := streamv3.GetOr(result, "count", int64(0))
		fmt.Printf("  Tags: %s → %d tasks\n", tagsJSON, count)
	}

	fmt.Println("\n🏗️ Test 3: MaterializeJSON with Record fields (impossible with old Materialize)")
	fmt.Println("------------------------------------------------------------------------------")

	userResults := streamv3.Chain(
		streamv3.MaterializeJSON("user", "user_json"),
		streamv3.GroupByFields("group_data", "user_json"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"count":    streamv3.Count(),
			"task_ids": streamv3.Collect("id"),
		}),
	)(slices.Values(tasks))

	fmt.Println("MaterializeJSON with Record fields:")
	for result := range userResults {
		userJSON := streamv3.GetOr(result, "user_json", "")
		count := streamv3.GetOr(result, "count", int64(0))
		fmt.Printf("  User: %s → %d tasks\n", userJSON, count)
	}

	fmt.Println("\n✅ Key Advantages of MaterializeJSON:")
	fmt.Println("  🎯 Type preservation: [\"urgent\",\"work\"] vs \"urgent,work\"")
	fmt.Println("  🏗️ Handles any complex type: Records, nested structures")
	fmt.Println("  📊 JSON standard: parseable, structured representation")
	fmt.Println("  🔄 Order sensitive: [\"A\",\"B\"] ≠ [\"B\",\"A\"] (correct for sequences)")
	fmt.Println("  🌐 Universal: works with any JSON-serializable data")
}