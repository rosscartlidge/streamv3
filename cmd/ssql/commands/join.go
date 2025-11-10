package commands

import (
	"fmt"
	"iter"
	"os"
	"strings"

	cf "github.com/rosscartlidge/autocli/v3"
	"github.com/rosscartlidge/ssql/v2"
	"github.com/rosscartlidge/ssql/v2/cmd/ssql/lib"
)

// RegisterJoin registers the join subcommand
func RegisterJoin(cmd *cf.CommandBuilder) *cf.CommandBuilder {
	cmd.Subcommand("join").
		Description("Join records from two data sources (SQL JOIN)").
		Example("ssql read-csv users.csv | ssql join -right orders.csv -on user_id", "Inner join users and orders on user_id").
		Example("ssql read-csv employees.csv | ssql join -type left -right departments.csv -on dept_id", "Left join employees with departments").
		Flag("-generate", "-g").
			Bool().
			Global().
			Help("Generate Go code instead of executing").
		Done().
		Flag("-type", "-t").
			String().
			Completer(&cf.StaticCompleter{Options: []string{"inner", "left", "right", "full"}}).
			Global().
			Default("inner").
			Help("Join type: inner, left, right, full (default: inner)").
		Done().
		Flag("-right", "-r").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.{csv,jsonl}"}).
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
		Flag("FILE").
			String().
			Completer(&cf.FileCompleter{Pattern: "*.jsonl"}).
			Global().
			Default("").
			Help("Left-side input JSONL file (or stdin if not specified)").
		Done().
		Handler(func(ctx *cf.Context) error {
			var inputFile, rightFile, joinType string
			var generate bool

			if fileVal, ok := ctx.GlobalFlags["FILE"]; ok {
				inputFile = fileVal.(string)
			}
			if rightVal, ok := ctx.GlobalFlags["-right"]; ok {
				rightFile = rightVal.(string)
			}
			if typeVal, ok := ctx.GlobalFlags["-type"]; ok {
				joinType = typeVal.(string)
			} else {
				joinType = "inner" // default
			}
			if genVal, ok := ctx.GlobalFlags["-generate"]; ok {
				generate = genVal.(bool)
			}

			// Validate required flags
			if rightFile == "" {
				return fmt.Errorf("right-side file required (use -right)")
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

			// Check if generation mode is enabled
			if shouldGenerate(generate) {
				return generateJoinCode(rightFile, joinType, onFields, leftField, rightField)
			}

			// Read left-side input (stdin or file)
			leftInput, err := lib.OpenInput(inputFile)
			if err != nil {
				return fmt.Errorf("opening left input: %w", err)
			}
			defer leftInput.Close()

			leftRecords := lib.ReadJSONL(leftInput)

			// Read right-side file
			var rightSeq iter.Seq[ssql.Record]
			if strings.HasSuffix(rightFile, ".csv") {
				csvRecords, err := ssql.ReadCSV(rightFile)
				if err != nil {
					return fmt.Errorf("reading right CSV: %w", err)
				}
				rightSeq = csvRecords
			} else {
				rightInput, err := os.Open(rightFile)
				if err != nil {
					return fmt.Errorf("opening right file: %w", err)
				}
				defer rightInput.Close()
				rightSeq = lib.ReadJSONL(rightInput)
			}

			// Build join predicate
			var predicate ssql.JoinPredicate
			if len(onFields) > 0 {
				predicate = ssql.OnFields(onFields...)
			} else {
				// Use different field names
				predicate = ssql.OnCondition(func(left, right ssql.Record) bool {
					leftVal, leftOk := ssql.Get[any](left, leftField)
					rightVal, rightOk := ssql.Get[any](right, rightField)
					if !leftOk || !rightOk {
						return false
					}
					return fmt.Sprintf("%v", leftVal) == fmt.Sprintf("%v", rightVal)
				})
			}

			// Apply appropriate join
			var joinFilter ssql.Filter[ssql.Record, ssql.Record]
			switch joinType {
			case "inner":
				joinFilter = ssql.InnerJoin(rightSeq, predicate)
			case "left":
				joinFilter = ssql.LeftJoin(rightSeq, predicate)
			case "right":
				joinFilter = ssql.RightJoin(rightSeq, predicate)
			case "full":
				joinFilter = ssql.FullJoin(rightSeq, predicate)
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
		Done()
	return cmd
}
