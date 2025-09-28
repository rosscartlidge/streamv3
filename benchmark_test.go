package streamv3

import (
	"testing"
)

// Benchmark data setup
var (
	benchmarkInts = func() []int {
		result := make([]int, 10000)
		for i := range result {
			result[i] = i
		}
		return result
	}()

	benchmarkRecords = func() []Record {
		result := make([]Record, 1000)
		for i := range result {
			result[i] = NewRecord().
				String("name", "person_"+string(rune(i%26+'A'))).
				Int("id", int64(i)).
				Int("score", int64(i%100)).
				Build()
		}
		return result
	}()
)

// =============================================================================
// STREAMV3 PERFORMANCE BENCHMARKS
// =============================================================================

func BenchmarkStreamV3_FilterAndCollect(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := From(benchmarkInts)
		filtered := stream.Where(func(x int) bool { return x%2 == 0 })
		_ = filtered.Collect()
	}
}

func BenchmarkStreamV3_MapAndCollect(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := From(benchmarkInts)
		mapped := MapTo(stream, func(x int) string { return string(rune(x%26 + 'A')) })
		_ = mapped.Collect()
	}
}

func BenchmarkStreamV3_Sort(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := From(benchmarkInts)
		sorted := SortOrdered(stream)
		_ = sorted.Collect()
	}
}

func BenchmarkStreamV3_RecordProcessing(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := From(benchmarkRecords)
		filtered := stream.Where(func(r Record) bool {
			if score, exists := r["score"]; exists {
				if scoreInt, ok := score.(int64); ok {
					return scoreInt > 50
				}
			}
			return false
		})
		_ = filtered.Collect()
	}
}

func BenchmarkStreamV3_GroupByAndAggregate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := From(benchmarkRecords)
		grouped := GroupRecordsByFields(stream, "name")
		aggregated := AggregateGroups(grouped, map[string]AggregateFunc{
			"count": Count(),
			"total": Sum("score"),
		})
		_ = aggregated.Collect()
	}
}

func BenchmarkStreamV3_ComplexPipeline(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := From(benchmarkInts)
		result := MapTo(
			stream.Where(func(x int) bool { return x%3 == 0 }).Limit(100),
			func(x int) int { return x * 2 },
		).Collect()
		_ = result
	}
}

func BenchmarkStreamV3_RawIteration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := From(benchmarkInts)
		count := 0
		for range stream.Iter() {
			count++
		}
		_ = count
	}
}

func BenchmarkStreamV3_MemoryAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := From(benchmarkInts[:1000]) // Smaller dataset for mem test
		filtered := stream.Where(func(x int) bool { return x%2 == 0 })
		_ = filtered.Collect()
	}
}

// =============================================================================
// ERROR HANDLING BENCHMARKS
// =============================================================================

func BenchmarkStreamV3_ErrorAware(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := FromSafe(benchmarkInts)
		filtered := stream.WhereSafe(func(x int) (bool, error) {
			return x%2 == 0, nil
		})
		result, err := filtered.Collect()
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkStreamV3_ErrorIgnoring(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream := FromSafe(benchmarkInts)
		simple := stream.IgnoreErrors()
		filtered := simple.Where(func(x int) bool { return x%2 == 0 })
		_ = filtered.Collect()
	}
}

// =============================================================================
// I/O OPERATION BENCHMARKS
// =============================================================================

func BenchmarkStreamV3_RecordCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := NewRecord().
			String("name", "test").
			Int("id", int64(i)).
			Float("score", 95.5).
			Bool("active", true).
			Build()
		_ = record
	}
}

func BenchmarkStreamV3_StandardMapCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := Record{
			"name":   "test",
			"id":     int64(i),
			"score":  95.5,
			"active": true,
		}
		_ = record
	}
}

// =============================================================================
// COMPARISON WITH STANDARD LIBRARY
// =============================================================================

func BenchmarkStandardLibrary_FilterAndCollect(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []int
		for _, x := range benchmarkInts {
			if x%2 == 0 {
				result = append(result, x)
			}
		}
		_ = result
	}
}

func BenchmarkStandardLibrary_Sort(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy slice to avoid modifying original
		data := make([]int, len(benchmarkInts))
		copy(data, benchmarkInts)

		// Standard library sort
		for i := 0; i < len(data)-1; i++ {
			for j := 0; j < len(data)-i-1; j++ {
				if data[j] > data[j+1] {
					data[j], data[j+1] = data[j+1], data[j]
				}
			}
		}
		_ = data
	}
}

// =============================================================================
// BENCHMARK SUMMARY
// =============================================================================

// Run with: go test -bench=. -benchmem
//
// Example commands:
// go test -bench=BenchmarkStreamV3_FilterAndCollect
// go test -bench=BenchmarkStreamV3_Sort
// go test -bench=BenchmarkStreamV3_.*MemoryAllocation -benchmem
// go test -bench=. -benchtime=5s
//
// Expected performance characteristics:
// 1. StreamV3 should be comparable to hand-written loops for simple operations
// 2. Sort operations should be very fast (uses standard library)
// 3. Error-aware operations have overhead but provide safety
// 4. Record creation with builder should be slightly slower than direct map creation
// 5. Complex pipelines should show good composability without major overhead
//
// Performance tips:
// - Use simple API for maximum performance when errors aren't critical
// - Filter early in pipelines to reduce subsequent processing
// - Consider materializing with Collect() vs iterating multiple times
// - Use standard library integration (Sort, etc.) for optimal performance