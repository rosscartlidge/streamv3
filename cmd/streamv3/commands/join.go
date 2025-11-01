package commands

import (
	"context"
	"fmt"
	"iter"
	"os"
	"strings"

	cf "github.com/rosscartlidge/completionflags"
	"github.com/rosscartlidge/streamv3"
	"github.com/rosscartlidge/streamv3/cmd/streamv3/lib"
)

// joinCommand implements the join command
type joinCommand struct {
	cmd *cf.Command
}

func init() {
	RegisterCommand(newJoinCommand())
}

func newJoinCommand() *joinCommand {
	var joinType, rightFile, inputFile string
	var generate bool

	joinTypes := []string{"inner", "left", "right", "full"}

	cmd := cf.NewCommand("join").
		Description("Join records from two data sources (SQL JOIN)").
		Flag("-type", "-t").
			String().
			Completer(&cf.StaticCompleter{Options: joinTypes}).
			Bind(&joinType).
			Global().
			Default("inner").
			Help("Join type: inner, left, right, full (default: inner)").
			Done().
		Flag("-right", "-r").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.{csv,jsonl}"}).
			Bind(&rightFile).
			Global().
			Help("Right-side file to join with (CSV or JSONL)").
			Done().
		Flag("-on").
			String().
			Completer(cf.NoCompleter{Hint: "<field-name>"}).
			Accumulate().
			Local().
			Help("Field name for equality join (same name in both sides)").
			Done().
		Flag("-left-field").
			String().
			Completer(cf.NoCompleter{Hint: "<left-field>"}).
			Local().
			Help("Field name from left side").
			Done().
		Flag("-right-field").
			String().
			Completer(cf.NoCompleter{Hint: "<right-field>"}).
			Local().
			Help("Field name from right side").
			Done().
		Flag("-generate", "-g").
			Bool().
			Bind(&generate).
			Global().
			Help("Generate Go code instead of executing").
			Done().
		Flag("-file", "-f").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
			Bind(&inputFile).
			Global().
			Default("").
			Help("Left-side input JSONL file (or stdin if not specified)").
			Done().
		Handler(func(ctx *cf.Context) error {
			// If -generate flag is set, generate Go code instead of executing
			if shouldGenerate(generate) {
				return generateJoinCode(ctx, joinType, rightFile, inputFile)
			}

			// Validate required flags
			if rightFile == "" {
				return fmt.Errorf("right-side file required (use -right)")
			}

			// Validate join type
			validType := false
			for _, t := range joinTypes {
				if joinType == t {
					validType = true
					break
				}
			}
			if !validType {
				return fmt.Errorf("invalid join type %q, must be one of: %v", joinType, joinTypes)
			}

			// Parse join condition from first clause
			var onFields []string
			var leftField, rightField string

			if len(ctx.Clauses) > 0 {
				clause := ctx.Clauses[0]

				// Get -on fields (simple equality on same field name)
				if onRaw, ok := clause.Flags["-on"]; ok {
					if onSlice, ok := onRaw.([]any); ok {
						for _, v := range onSlice {
							if field, ok := v.(string); ok && field != "" {
								onFields = append(onFields, field)
							}
						}
					}
				}

				// Get -left-field and -right-field
				if lf, ok := clause.Flags["-left-field"].(string); ok {
					leftField = lf
				}
				if rf, ok := clause.Flags["-right-field"].(string); ok {
					rightField = rf
				}
			}

			// Validate join conditions
			if len(onFields) == 0 && (leftField == "" || rightField == "") {
				return fmt.Errorf("join condition required: use -on <field> OR (-left-field <field> -right-field <field>)")
			}
			if len(onFields) > 0 && (leftField != "" || rightField != "") {
				return fmt.Errorf("cannot use both -on and -left-field/-right-field")
			}

			// Read left-side input (stdin or file)
			leftInput, err := lib.OpenInput(inputFile)
			if err != nil {
				return fmt.Errorf("opening left input: %w", err)
			}
			defer leftInput.Close()

			leftRecords := lib.ReadJSONL(leftInput)

			// Read right-side file into an iter.Seq[Record]
			var rightSeq iter.Seq[streamv3.Record]
			if strings.HasSuffix(rightFile, ".csv") {
				// Read CSV
				csvRecords, err := streamv3.ReadCSV(rightFile)
				if err != nil {
					return fmt.Errorf("reading right CSV: %w", err)
				}
				rightSeq = csvRecords
			} else {
				// Read JSONL
				rightInput, err := os.Open(rightFile)
				if err != nil {
					return fmt.Errorf("opening right file: %w", err)
				}
				defer rightInput.Close()
				rightSeq = lib.ReadJSONL(rightInput)
			}

			// Build join predicate
			var predicate streamv3.JoinPredicate
			if len(onFields) > 0 {
				predicate = streamv3.OnFields(onFields...)
			} else {
				// Use different field names
				predicate = streamv3.OnCondition(func(left, right streamv3.Record) bool {
					leftVal, leftOk := left[leftField]
					rightVal, rightOk := right[rightField]
					if !leftOk || !rightOk {
						return false
					}
					return fmt.Sprintf("%v", leftVal) == fmt.Sprintf("%v", rightVal)
				})
			}

			// Apply appropriate join
			// JOIN operations expect iter.Seq[Record] not Filter
			var joinFilter streamv3.Filter[streamv3.Record, streamv3.Record]
			switch joinType {
			case "inner":
				joinFilter = streamv3.InnerJoin(rightSeq, predicate)
			case "left":
				joinFilter = streamv3.LeftJoin(rightSeq, predicate)
			case "right":
				joinFilter = streamv3.RightJoin(rightSeq, predicate)
			case "full":
				joinFilter = streamv3.FullJoin(rightSeq, predicate)
			default:
				return fmt.Errorf("unsupported join type: %s", joinType)
			}

			joined := joinFilter(leftRecords)

			// Write output as JSONL
			if err := lib.WriteJSONL(os.Stdout, joined); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		}).
		Build()

	return &joinCommand{cmd: cmd}
}

func (c *joinCommand) Name() string {
	return "join"
}

func (c *joinCommand) Description() string {
	return "Join records from two data sources (SQL JOIN)"
}

func (c *joinCommand) GetCFCommand() *cf.Command {
	return c.cmd
}

func (c *joinCommand) Execute(ctx context.Context, args []string) error {
	// Handle -help flag before completionflags framework takes over
	if len(args) > 0 && (args[0] == "-help" || args[0] == "--help") {
		fmt.Println("join - Join records from two data sources (SQL JOIN)")
		fmt.Println()
		fmt.Println("Usage: streamv3 join -type <type> -right <file> [-on <field>]...")
		fmt.Println("       streamv3 join -type <type> -right <file> -left-field <field> -right-field <field>")
		fmt.Println()
		fmt.Println("Join Types:")
		fmt.Println("  inner  Inner join (default) - only matching records")
		fmt.Println("  left   Left join - all left records, matched right records")
		fmt.Println("  right  Right join - all right records, matched left records")
		fmt.Println("  full   Full outer join - all records from both sides")
		fmt.Println()
		fmt.Println("Join Conditions:")
		fmt.Println("  -on <field>                    Field with same name in both sides")
		fmt.Println("  -on <field1> -on <field2>      Multiple fields (composite key)")
		fmt.Println("  -left-field / -right-field     Different field names in each side")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Inner join on single field")
		fmt.Println("  streamv3 read-csv employees.csv | \\")
		fmt.Println("    streamv3 join -type inner -right departments.csv -on dept_id")
		fmt.Println()
		fmt.Println("  # Left join with different field names")
		fmt.Println("  streamv3 read-csv employees.csv | \\")
		fmt.Println("    streamv3 join -type left -right depts.csv \\")
		fmt.Println("      -left-field department_id -right-field id")
		fmt.Println()
		fmt.Println("  # Join on multiple fields")
		fmt.Println("  streamv3 read-csv sales.csv | \\")
		fmt.Println("    streamv3 join -right products.csv \\")
		fmt.Println("      -on product_id -on region")
		fmt.Println()
		fmt.Println("  # SQL equivalent:")
		fmt.Println("  # SELECT * FROM employees e")
		fmt.Println("  # INNER JOIN departments d ON e.dept_id = d.dept_id")
		fmt.Println("  streamv3 read-csv employees.csv | \\")
		fmt.Println("    streamv3 join -type inner -right departments.csv -on dept_id")
		return nil
	}

	return c.cmd.Execute(args)
}

// generateJoinCode generates Go code for the join command
func generateJoinCode(ctx *cf.Context, joinType, rightFile, inputFile string) error {
	// Read all previous code fragments from stdin (if any)
	fragments, err := lib.ReadAllCodeFragments()
	if err != nil {
		return fmt.Errorf("reading code fragments: %w", err)
	}

	// Pass through all previous fragments
	for _, frag := range fragments {
		if err := lib.WriteCodeFragment(frag); err != nil {
			return fmt.Errorf("writing previous fragment: %w", err)
		}
	}

	// Get input variable from last fragment
	var inputVar string
	if len(fragments) > 0 {
		inputVar = fragments[len(fragments)-1].Var
	} else {
		inputVar = "records"
	}

	// Parse join condition
	var onFields []string
	var leftField, rightField string

	if len(ctx.Clauses) > 0 {
		clause := ctx.Clauses[0]

		if onRaw, ok := clause.Flags["-on"]; ok {
			if onSlice, ok := onRaw.([]any); ok {
				for _, v := range onSlice {
					if field, ok := v.(string); ok && field != "" {
						onFields = append(onFields, field)
					}
				}
			}
		}

		if lf, ok := clause.Flags["-left-field"].(string); ok {
			leftField = lf
		}
		if rf, ok := clause.Flags["-right-field"].(string); ok {
			rightField = rf
		}
	}

	// Generate code to read right-side data
	var rightVarCode string
	if strings.HasSuffix(rightFile, ".csv") {
		rightVarCode = fmt.Sprintf(`rightRecords, err := streamv3.ReadCSV(%q)
	if err != nil {
		return fmt.Errorf("reading right CSV: %%w", err)
	}`, rightFile)
	} else {
		rightVarCode = fmt.Sprintf(`rightFile, err := os.Open(%q)
	if err != nil {
		return fmt.Errorf("opening right file: %%w", err)
	}
	defer rightFile.Close()
	rightRecords := /* read JSONL from rightFile */`, rightFile)
	}

	// Generate predicate code
	var predicateCode string
	if len(onFields) > 0 {
		var fieldsList []string
		for _, f := range onFields {
			fieldsList = append(fieldsList, fmt.Sprintf("%q", f))
		}
		predicateCode = fmt.Sprintf("streamv3.OnFields(%s)", strings.Join(fieldsList, ", "))
	} else {
		predicateCode = fmt.Sprintf(`streamv3.OnCondition(func(left, right streamv3.Record) bool {
		leftVal, leftOk := left[%q]
		rightVal, rightOk := right[%q]
		if !leftOk || !rightOk {
			return false
		}
		return fmt.Sprintf("%%v", leftVal) == fmt.Sprintf("%%v", rightVal)
	})`, leftField, rightField)
	}

	// Generate join call
	var joinFunc string
	switch joinType {
	case "inner":
		joinFunc = "InnerJoin"
	case "left":
		joinFunc = "LeftJoin"
	case "right":
		joinFunc = "RightJoin"
	case "full":
		joinFunc = "FullJoin"
	default:
		joinFunc = "InnerJoin"
	}

	code := fmt.Sprintf(`%s
	joined := streamv3.%s(rightRecords, %s)(%s)`, rightVarCode, joinFunc, predicateCode, inputVar)

	// Create code fragment
	imports := []string{"fmt", "os"}
	frag := lib.NewStmtFragment("joined", inputVar, code, imports, getCommandString())

	// Write to stdout
	return lib.WriteCodeFragment(frag)
}
