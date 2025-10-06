package main

import (
	"fmt"
	"os"
	"strings"
	"github.com/rosscartlidge/streamv3"
)

func main() {
	fmt.Println("🔧 Unix Pipes Demo: io.Reader/io.Writer Integration")
	fmt.Println("===================================================\n")

	// Demo 1: Read from stdin, process, write to stdout
	if len(os.Args) > 1 && os.Args[1] == "stdin-demo" {
		fmt.Fprintln(os.Stderr, "📥 Reading CSV from stdin, processing, writing JSON to stdout...")

		// Read CSV from stdin
		csvStream := streamv3.ReadCSVFromReader(os.Stdin)

		// Process the stream - add calculated fields
		processedStream := streamv3.Map(func(record streamv3.Record) streamv3.Record {
			// Add processed timestamp
			record["processed_at"] = "2024-01-01T10:00:00Z"

			// Convert string field to uppercase if it exists
			if name, ok := streamv3.Get[string](record, "name"); ok {
				record["name_upper"] = strings.ToUpper(name)
			}

			return record
		})(csvStream)

		// Write JSON to stdout
		stream := streamv3.From([]streamv3.Record{})
		for record := range processedStream {
			// We need to collect records first for the Stream wrapper
			records := []streamv3.Record{record}
			for r := range processedStream {
				records = append(records, r)
				break // Just get one more for demo
			}
			stream = streamv3.From(records)
			break
		}

		err := streamv3.WriteJSONToWriter(stream, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Demo 2: Generate sample data and demonstrate pipe workflows
	fmt.Println("🔧 Demo 1: Generating sample CSV data")

	// Create sample data
	sampleData := []streamv3.Record{
		streamv3.NewRecord().String("id", "1").String("name", "Alice").Int("age", 30).Float("score", 95.5).Build(),
		streamv3.NewRecord().String("id", "2").String("name", "Bob").Int("age", 25).Float("score", 87.2).Build(),
		streamv3.NewRecord().String("id", "3").String("name", "Carol").Int("age", 35).Float("score", 92.8).Build(),
	}

	// Write CSV to a file
	csvFile := "/tmp/demo_data.csv"
	fields := []string{"id", "name", "age", "score"}

	err := streamv3.WriteCSV(streamv3.From(sampleData), csvFile, fields)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Written sample data to %s\n", csvFile)

	fmt.Println("\n🔧 Demo 2: Unix pipe workflow simulation")
	fmt.Println("In a real Unix environment, you could now run:")
	fmt.Println()
	fmt.Println("# Read CSV, process, output JSON:")
	fmt.Println("cat /tmp/demo_data.csv | go run unix_pipes_demo.go stdin-demo")
	fmt.Println()
	fmt.Println("# Chain multiple processing steps:")
	fmt.Println("cat /tmp/demo_data.csv | \\")
	fmt.Println("  go run process1.go | \\")
	fmt.Println("  go run process2.go | \\")
	fmt.Println("  go run process3.go > final_output.json")
	fmt.Println()

	fmt.Println("🔧 Demo 3: In-memory pipe simulation")

	// Simulate pipe workflow using strings.Reader and strings.Builder
	csvData := `id,name,age,score
1,Alice,30,95.5
2,Bob,25,87.2
3,Carol,35,92.8`

	// Step 1: CSV → Stream (simulating: cat data.csv | program1)
	csvReader := strings.NewReader(csvData)
	stream1 := streamv3.ReadCSVFromReader(csvReader)

	// Step 2: Process Stream (simulating: program1 | program2)
	var processedRecords []streamv3.Record
	for record := range stream1 {
		// Add processing
		if age, ok := streamv3.Get[int64](record, "age"); ok && age >= 30 {
			record["category"] = "senior"
		} else {
			record["category"] = "junior"
		}
		processedRecords = append(processedRecords, record)
	}

	// Step 3: Stream → JSON (simulating: program2 | program3 > output.json)
	var jsonOutput strings.Builder
	err = streamv3.WriteJSONToWriter(streamv3.From(processedRecords), &jsonOutput)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("📤 Processed output:")
	fmt.Println(jsonOutput.String())

	fmt.Println("✅ Benefits of io.Reader/io.Writer approach:")
	fmt.Println("   🔗 Unix pipe compatibility: stdin/stdout integration")
	fmt.Println("   🧪 Easy testing: strings.Reader/strings.Builder")
	fmt.Println("   🌐 Network streams: http.Request.Body, http.ResponseWriter")
	fmt.Println("   💾 Any io source: files, buffers, networks, compression")
	fmt.Println("   🔧 Composable: mix and match different I/O sources")

	fmt.Printf("\n💡 To test real Unix pipes, run:\n")
	fmt.Printf("   echo '%s' | go run unix_pipes_demo.go stdin-demo\n", strings.ReplaceAll(csvData, "\n", "\\n"))
}