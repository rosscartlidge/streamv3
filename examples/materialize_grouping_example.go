package main

import (
	"fmt"
	"iter"
	"slices"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🎯 Materialize + GroupBy Content Example")
	fmt.Println("========================================\n")

	// Create records where some have the same tag sequences (different instances)
	urgentWork1 := slices.Values([]string{"urgent", "work"})
	urgentWork2 := slices.Values([]string{"urgent", "work"}) // Same content, different sequence
	feature := slices.Values([]string{"feature", "enhancement"})
	bugFix := slices.Values([]string{"urgent", "work"}) // Same content as urgentWork1

	tasks := []streamv3.Record{
		streamv3.NewRecord().String("id", "TASK-001").String("team", "Backend").StringSeq("tags", urgentWork1).Build(),
		streamv3.NewRecord().String("id", "TASK-002").String("team", "Frontend").StringSeq("tags", urgentWork2).Build(), // Same tags content
		streamv3.NewRecord().String("id", "TASK-003").String("team", "Backend").StringSeq("tags", feature).Build(),
		streamv3.NewRecord().String("id", "TASK-004").String("team", "QA").StringSeq("tags", bugFix).Build(), // Same tags content again
	}

	fmt.Println("📊 Tasks with iter.Seq tag fields:")
	for i, task := range tasks {
		id := streamv3.GetOr(task, "id", "")
		team := streamv3.GetOr(task, "team", "")
		fmt.Printf("  %d. %s (%s team)\n", i+1, id, team)

		if tagsSeq, ok := streamv3.Get[iter.Seq[string]](task, "tags"); ok {
			fmt.Print("     Tags: ")
			for tag := range tagsSeq {
				fmt.Printf("%s ", tag)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n🔧 Step 1: Materialize tag sequences to enable content-based grouping")
	fmt.Println("--------------------------------------------------------------------")

	// Materialize sequences to string representations
	results := streamv3.Chain(
		streamv3.Materialize("tags", "tags_key", ","),
		streamv3.GroupByFields("group_data", "tags_key"), // Group by materialized content
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"count":      streamv3.Count(),
			"teams":      streamv3.Collect("team"),
			"task_ids":   streamv3.Collect("id"),
		}),
	)(slices.Values(tasks))

	fmt.Println("Groups by tag content (not sequence instance):")
	for result := range results {
		tagsKey := streamv3.GetOr(result, "tags_key", "")
		count := streamv3.GetOr(result, "count", int64(0))
		fmt.Printf("\n  Tag combination: \"%s\" (%d tasks)\n", tagsKey, count)

		// Get teams and task IDs as any type, then handle
		if teams, ok := streamv3.Get[any](result, "teams"); ok {
			fmt.Print("    Teams: ")
			if teamsSlice, ok := teams.([]any); ok {
				for _, team := range teamsSlice {
					fmt.Printf("%v ", team)
				}
			} else {
				fmt.Printf("%v", teams)
			}
			fmt.Println()
		}

		if taskIds, ok := streamv3.Get[any](result, "task_ids"); ok {
			fmt.Print("    Task IDs: ")
			if idsSlice, ok := taskIds.([]any); ok {
				for _, id := range idsSlice {
					fmt.Printf("%v ", id)
				}
			} else {
				fmt.Printf("%v", taskIds)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n✅ Key Success: Content-based grouping works!")
	fmt.Println("  • TASK-001, TASK-002, TASK-004 all grouped together")
	fmt.Println("  • They have different iter.Seq instances but same content")
	fmt.Println("  • Without Materialize(), they would be in separate groups")

	fmt.Println("\n🆚 Comparison with different separators:")
	fmt.Println("----------------------------------------")

	// Try different separator to show control
	pipeSeparated := streamv3.Chain(
		streamv3.Materialize("tags", "tags_pipe", "|"), // Different separator
		streamv3.GroupByFields("group_data", "tags_pipe"),
		streamv3.Aggregate("group_data", map[string]streamv3.AggregateFunc{
			"count": streamv3.Count(),
		}),
	)(slices.Values(tasks))

	fmt.Println("Same grouping with pipe separator:")
	for result := range pipeSeparated {
		tagsPipe := streamv3.GetOr(result, "tags_pipe", "")
		count := streamv3.GetOr(result, "count", int64(0))
		fmt.Printf("  \"%s\": %d tasks\n", tagsPipe, count)
	}

	fmt.Println("\n💡 Benefits demonstrated:")
	fmt.Println("  🚀 Efficient: No record explosion (4 tasks → 4 tasks)")
	fmt.Println("  🎯 Semantic: Groups by actual content, not object identity")
	fmt.Println("  🔧 Flexible: User controls materialization format")
	fmt.Println("  🔒 Safe: GroupBy validation prevents accidental complex field usage")
}