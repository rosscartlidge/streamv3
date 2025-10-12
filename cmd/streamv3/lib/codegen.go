package lib

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// CodeGenerator generates Go code from a Pipeline
type CodeGenerator struct {
	buf    *bytes.Buffer
	indent int
}

// GenerateGoCode converts a pipeline to Go code
func GenerateGoCode(pipeline *Pipeline) (string, error) {
	g := &CodeGenerator{buf: new(bytes.Buffer)}

	// Generate package and imports
	g.writePackage()
	g.writeImports(pipeline)
	g.writeln("")
	g.writeMainFunc(pipeline)

	return g.buf.String(), nil
}

func (g *CodeGenerator) writePackage() {
	g.writeln("package main")
	g.writeln("")
}

func (g *CodeGenerator) writeImports(p *Pipeline) {
	needsLog := false
	needsStrings := false

	// Check what imports we need
	for _, cmd := range p.Commands {
		if cmd.Name == "write-csv" || cmd.Name == "where" {
			needsLog = true
		}
		if cmd.Name == "where" {
			if op, ok := cmd.Args["op"]; ok && (op == "contains" || op == "startswith" || op == "endswith") {
				needsStrings = true
			}
		}
	}

	g.writeln("import (")
	if needsLog {
		g.writeln("\t\"log\"")
	}
	if needsStrings {
		g.writeln("\t\"strings\"")
	}
	g.writeln("\t\"github.com/rosscartlidge/streamv3\"")
	g.writeln(")")
}

func (g *CodeGenerator) writeMainFunc(p *Pipeline) {
	g.writeln("func main() {")
	g.indent++

	varName := "records"
	varCounter := 0

	for _, cmd := range p.Commands {
		switch cmd.Name {
		case "read-csv":
			g.writeReadCSV(varName, cmd)

		case "where":
			newVar := fmt.Sprintf("filtered%d", varCounter)
			varCounter++
			g.writeWhere(varName, newVar, cmd)
			varName = newVar

		case "select":
			newVar := fmt.Sprintf("selected%d", varCounter)
			varCounter++
			g.writeSelect(varName, newVar, cmd)
			varName = newVar

		case "limit":
			newVar := fmt.Sprintf("limited%d", varCounter)
			varCounter++
			g.writeLimit(varName, newVar, cmd)
			varName = newVar

		case "sort":
			newVar := fmt.Sprintf("sorted%d", varCounter)
			varCounter++
			g.writeSort(varName, newVar, cmd)
			varName = newVar

		case "write-csv":
			g.writeWriteCSV(varName, cmd)

		default:
			g.writeComment(fmt.Sprintf("Unsupported command: %s", cmd.Name))
		}
	}

	g.indent--
	g.writeln("}")
}

func (g *CodeGenerator) writeReadCSV(varName string, cmd Command) {
	inputFile := cmd.GetInputFile()
	if inputFile == "" {
		inputFile = "stdin"
		g.writeComment("TODO: Replace 'stdin' with actual filename or stdin handling")
	}

	g.writeln(fmt.Sprintf("// Read CSV from %s", inputFile))
	g.writeln(fmt.Sprintf("%s := streamv3.ReadCSV(%q)", varName, inputFile))
	g.writeln("")
}

func (g *CodeGenerator) writeWhere(input, output string, cmd Command) {
	field := cmd.Args["field"]
	op := cmd.Args["op"]
	value := cmd.Args["value"]

	g.writeln(fmt.Sprintf("// Filter: %s %s %s", field, op, value))
	g.writeln(fmt.Sprintf("%s := streamv3.Where(func(r streamv3.Record) bool {", output))
	g.indent++

	// Generate comparison based on operator
	switch op {
	case "eq":
		// Try to detect type from value
		if isNumeric(value) {
			g.writeln(fmt.Sprintf("if val, ok := r[%q].(float64); ok {", field))
			g.indent++
			g.writeln(fmt.Sprintf("return val == %s", value))
			g.indent--
			g.writeln("}")
		} else {
			g.writeln(fmt.Sprintf("return r[%q] == %q", field, value))
		}

	case "ne":
		if isNumeric(value) {
			g.writeln(fmt.Sprintf("if val, ok := r[%q].(float64); ok {", field))
			g.indent++
			g.writeln(fmt.Sprintf("return val != %s", value))
			g.indent--
			g.writeln("}")
		} else {
			g.writeln(fmt.Sprintf("return r[%q] != %q", field, value))
		}

	case "gt":
		g.writeln(fmt.Sprintf("if val, ok := r[%q].(float64); ok {", field))
		g.indent++
		g.writeln(fmt.Sprintf("return val > %s", value))
		g.indent--
		g.writeln("}")

	case "ge":
		g.writeln(fmt.Sprintf("if val, ok := r[%q].(float64); ok {", field))
		g.indent++
		g.writeln(fmt.Sprintf("return val >= %s", value))
		g.indent--
		g.writeln("}")

	case "lt":
		g.writeln(fmt.Sprintf("if val, ok := r[%q].(float64); ok {", field))
		g.indent++
		g.writeln(fmt.Sprintf("return val < %s", value))
		g.indent--
		g.writeln("}")

	case "le":
		g.writeln(fmt.Sprintf("if val, ok := r[%q].(float64); ok {", field))
		g.indent++
		g.writeln(fmt.Sprintf("return val <= %s", value))
		g.indent--
		g.writeln("}")

	case "contains":
		g.writeln(fmt.Sprintf("if val, ok := r[%q].(string); ok {", field))
		g.indent++
		g.writeln(fmt.Sprintf("return strings.Contains(val, %q)", value))
		g.indent--
		g.writeln("}")

	case "startswith":
		g.writeln(fmt.Sprintf("if val, ok := r[%q].(string); ok {", field))
		g.indent++
		g.writeln(fmt.Sprintf("return strings.HasPrefix(val, %q)", value))
		g.indent--
		g.writeln("}")

	case "endswith":
		g.writeln(fmt.Sprintf("if val, ok := r[%q].(string); ok {", field))
		g.indent++
		g.writeln(fmt.Sprintf("return strings.HasSuffix(val, %q)", value))
		g.indent--
		g.writeln("}")

	default:
		g.writeComment(fmt.Sprintf("Unsupported operator: %s", op))
		g.writeln("return false")
	}

	g.writeln("return false")
	g.indent--
	g.writeln(fmt.Sprintf("})(%s)", input))
	g.writeln("")
}

func (g *CodeGenerator) writeSelect(input, output string, cmd Command) {
	// Parse multiple -field flags
	fields := []string{}
	renames := make(map[string]string)

	// Note: Our parser currently only supports single values per flag
	// For MVP, we'll handle the simple case of a single field
	if field, ok := cmd.Args["field"]; ok {
		fields = append(fields, field)
		if as, ok := cmd.Args["as"]; ok {
			renames[field] = as
		}
	}

	if len(fields) == 0 {
		g.writeComment("Select: no fields specified")
		g.writeln(fmt.Sprintf("%s := %s", output, input))
		g.writeln("")
		return
	}

	g.writeln(fmt.Sprintf("// Select fields: %s", strings.Join(fields, ", ")))
	g.writeln(fmt.Sprintf("%s := streamv3.Select(func(r streamv3.Record) streamv3.Record {", output))
	g.indent++
	g.writeln("return streamv3.Record{")
	g.indent++

	for _, field := range fields {
		outputName := field
		if renamed, ok := renames[field]; ok {
			outputName = renamed
		}
		g.writeln(fmt.Sprintf("%q: r[%q],", outputName, field))
	}

	g.indent--
	g.writeln("}")
	g.indent--
	g.writeln(fmt.Sprintf("})(%s)", input))
	g.writeln("")
}

func (g *CodeGenerator) writeLimit(input, output string, cmd Command) {
	n := cmd.Args["n"]
	if n == "" {
		n = "100"
	}

	g.writeln(fmt.Sprintf("// Limit to first %s records", n))
	g.writeln(fmt.Sprintf("%s := streamv3.Limit[streamv3.Record](%s)(%s)", output, n, input))
	g.writeln("")
}

func (g *CodeGenerator) writeSort(input, output string, cmd Command) {
	field := cmd.Args["field"]
	desc := cmd.Args["desc"] == "true"

	descStr := ""
	if desc {
		descStr = " (descending)"
	}

	g.writeln(fmt.Sprintf("// Sort by %s%s", field, descStr))
	g.writeln(fmt.Sprintf("%s := streamv3.SortBy(func(r streamv3.Record) float64 {", output))
	g.indent++
	g.writeln(fmt.Sprintf("if val, ok := r[%q].(float64); ok {", field))
	g.indent++
	if desc {
		g.writeln("return -val // Descending")
	} else {
		g.writeln("return val")
	}
	g.indent--
	g.writeln("}")
	g.writeln("return 0")
	g.indent--
	g.writeln(fmt.Sprintf("})(%s)", input))
	g.writeln("")
}

func (g *CodeGenerator) writeWriteCSV(input string, cmd Command) {
	outputFile := cmd.GetOutputFile()
	if outputFile == "" {
		outputFile = "output.csv"
		g.writeComment("TODO: Replace 'output.csv' with actual filename or stdout handling")
	}

	g.writeln(fmt.Sprintf("// Write CSV to %s", outputFile))
	g.writeln(fmt.Sprintf("if err := streamv3.WriteCSV(%s, %q); err != nil {", input, outputFile))
	g.indent++
	g.writeln("log.Fatalf(\"Error writing CSV: %v\", err)")
	g.indent--
	g.writeln("}")
}

func (g *CodeGenerator) writeln(s string) {
	if s == "" {
		g.buf.WriteByte('\n')
		return
	}

	for i := 0; i < g.indent; i++ {
		g.buf.WriteByte('\t')
	}
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
}

func (g *CodeGenerator) writeComment(s string) {
	g.writeln(fmt.Sprintf("// %s", s))
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
