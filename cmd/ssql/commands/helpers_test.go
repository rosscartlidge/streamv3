package commands

import (
	"testing"

	"github.com/rosscartlidge/ssql/v2"
)

func TestIsExpression(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		// Expressions (should return true)
		{name: "math addition", value: "price + tax", want: true},
		{name: "math multiplication", value: "price * quantity", want: true},
		{name: "math division", value: "total / count", want: true},
		{name: "comparison", value: "age > 18", want: true},
		{name: "equality", value: "status == \"active\"", want: true},
		{name: "ternary", value: "x > 10 ? \"high\" : \"low\"", want: true},
		{name: "function call", value: "upper(name)", want: true},
		{name: "logical and", value: "age >= 18 && active", want: true},
		{name: "logical or", value: "premium || vip", want: true},

		// Literals (should return false)
		{name: "plain string", value: "active", want: false},
		{name: "number string", value: "42", want: false},
		{name: "float string", value: "3.14", want: false},
		{name: "boolean string", value: "true", want: false},
		{name: "string with dash", value: "foo-bar", want: false},
		{name: "string with underscore", value: "foo_bar", want: false},
		{name: "email", value: "user@example.com", want: false},  // @ is not an operator we check
		{name: "url", value: "https://example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExpression(tt.value)
			if got != tt.want {
				t.Errorf("isExpression(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestEvaluateExpression_Math(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		record     ssql.Record
		want       any
		wantErr    bool
	}{
		{
			name:       "simple addition",
			expression: "2 + 2",
			record:     ssql.MakeMutableRecord().Freeze(),
			want:       4,
			wantErr:    false,
		},
		{
			name:       "multiplication",
			expression: "price * quantity",
			record: ssql.MakeMutableRecord().
				Float("price", 10.5).
				Int("quantity", int64(3)).
				Freeze(),
			want:    31.5,
			wantErr: false,
		},
		{
			name:       "complex math",
			expression: "(price * quantity) * 1.1",
			record: ssql.MakeMutableRecord().
				Float("price", 100.0).
				Int("quantity", int64(5)).
				Freeze(),
			want:    550.0,
			wantErr: false,
		},
		{
			name:       "division",
			expression: "total / count",
			record: ssql.MakeMutableRecord().
				Float("total", 100.0).
				Float("count", 4.0).
				Freeze(),
			want:    25.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateExpression(tt.expression, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("evaluateExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateExpression_Conditionals(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		record     ssql.Record
		want       any
		wantErr    bool
	}{
		{
			name:       "simple ternary",
			expression: "x > 10 ? \"high\" : \"low\"",
			record: ssql.MakeMutableRecord().
				Int("x", int64(15)).
				Freeze(),
			want:    "high",
			wantErr: false,
		},
		{
			name:       "nested ternary",
			expression: "sales > 10000 ? \"gold\" : sales > 5000 ? \"silver\" : \"bronze\"",
			record: ssql.MakeMutableRecord().
				Float("sales", 7500.0).
				Freeze(),
			want:    "silver",
			wantErr: false,
		},
		{
			name:       "ternary with math",
			expression: "quantity > 100 ? price * 0.8 : price",
			record: ssql.MakeMutableRecord().
				Float("price", 100.0).
				Int("quantity", int64(150)).
				Freeze(),
			want:    80.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateExpression(tt.expression, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("evaluateExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateExpression_Strings(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		record     ssql.Record
		want       any
		wantErr    bool
	}{
		{
			name:       "string concatenation",
			expression: "first + \" \" + last",
			record: ssql.MakeMutableRecord().
				String("first", "John").
				String("last", "Doe").
				Freeze(),
			want:    "John Doe",
			wantErr: false,
		},
		{
			name:       "string comparison",
			expression: "status == \"active\"",
			record: ssql.MakeMutableRecord().
				String("status", "active").
				Freeze(),
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateExpression(tt.expression, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("evaluateExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateExpression_Errors(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		record     ssql.Record
		wantErr    bool
	}{
		{
			name:       "invalid syntax",
			expression: "price * *",
			record:     ssql.MakeMutableRecord().Freeze(),
			wantErr:    true,
		},
		{
			name:       "undefined variable",
			expression: "nonexistent + 10",
			record:     ssql.MakeMutableRecord().Freeze(),
			wantErr:    true,
		},
		{
			name:       "type mismatch",
			expression: "name * 2",
			record: ssql.MakeMutableRecord().
				String("name", "Alice").
				Freeze(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluateExpression(tt.expression, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEvaluateExpression_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		record     ssql.Record
		want       any
		wantErr    bool
	}{
		{
			name:       "discount calculation",
			expression: "tier == \"gold\" ? price * 0.8 : tier == \"silver\" ? price * 0.9 : price",
			record: ssql.MakeMutableRecord().
				String("tier", "silver").
				Float("price", 100.0).
				Freeze(),
			want:    90.0,
			wantErr: false,
		},
		{
			name:       "total with tax",
			expression: "(price * quantity) * (1 + tax_rate)",
			record: ssql.MakeMutableRecord().
				Float("price", 50.0).
				Int("quantity", int64(2)).
				Float("tax_rate", 0.2).
				Freeze(),
			want:    120.0,
			wantErr: false,
		},
		{
			name:       "boolean logic",
			expression: "age >= 18 && status == \"active\"",
			record: ssql.MakeMutableRecord().
				Int("age", int64(25)).
				String("status", "active").
				Freeze(),
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateExpression(tt.expression, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("evaluateExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}
