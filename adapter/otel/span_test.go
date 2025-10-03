package otel

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestConvertAttributeValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  attribute.Value
	}{
		{
			name:  "bool value",
			value: true,
			want:  attribute.BoolValue(true),
		},
		{
			name:  "int64 value",
			value: int64(42),
			want:  attribute.Int64Value(42),
		},
		{
			name:  "float64 value",
			value: float64(3.14),
			want:  attribute.Float64Value(3.14),
		},
		{
			name:  "string value",
			value: "test",
			want:  attribute.StringValue("test"),
		},
		{
			name:  "bool slice",
			value: []bool{true, false, true},
			want:  attribute.BoolSliceValue([]bool{true, false, true}),
		},
		{
			name:  "int64 slice",
			value: []int64{1, 2, 3},
			want:  attribute.Int64SliceValue([]int64{1, 2, 3}),
		},
		{
			name:  "float64 slice",
			value: []float64{1.1, 2.2, 3.3},
			want:  attribute.Float64SliceValue([]float64{1.1, 2.2, 3.3}),
		},
		{
			name:  "string slice",
			value: []string{"a", "b", "c"},
			want:  attribute.StringSliceValue([]string{"a", "b", "c"}),
		},
		{
			name:  "int value",
			value: 42,
			want:  attribute.IntValue(42),
		},
		{
			name:  "unsupported type - struct",
			value: struct{ Name string }{Name: "test"},
			want:  attribute.StringValue("{test}"),
		},
		{
			name:  "nil value",
			value: nil,
			want:  attribute.StringValue("<nil>"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertAttributeValue(tt.value)

			// Compare types and values
			if got.Type() != tt.want.Type() {
				t.Errorf("convertAttributeValue() type = %v, want %v", got.Type(), tt.want.Type())
				return
			}

			// For slice types, compare using AsInterface() to get the underlying slice
			gotVal := got.AsInterface()
			wantVal := tt.want.AsInterface()

			// Use string comparison as a simple way to verify equality
			if got.Type() == attribute.STRINGSLICE ||
				got.Type() == attribute.INT64SLICE ||
				got.Type() == attribute.FLOAT64SLICE ||
				got.Type() == attribute.BOOLSLICE {
				gotStr := got.Emit()
				wantStr := tt.want.Emit()
				if gotStr != wantStr {
					t.Errorf("convertAttributeValue() = %v, want %v", gotStr, wantStr)
				}
			} else if gotVal != wantVal {
				t.Errorf("convertAttributeValue() = %v, want %v", gotVal, wantVal)
			}
		})
	}
}
