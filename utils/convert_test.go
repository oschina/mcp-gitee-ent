package utils

import (
	"testing"
)

func TestSafelyConvertToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "string类型",
			input:    "hello",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "int类型",
			input:    123,
			expected: "123",
			wantErr:  false,
		},
		{
			name:     "int32类型",
			input:    int32(123),
			expected: "123",
			wantErr:  false,
		},
		{
			name:     "int64类型",
			input:    int64(123),
			expected: "123",
			wantErr:  false,
		},
		{
			name:     "float32类型-整数",
			input:    float32(123.0),
			expected: "123",
			wantErr:  false,
		},
		{
			name:     "float32类型-小数",
			input:    float32(123.45),
			expected: "123.45",
			wantErr:  false,
		},
		{
			name:     "float64类型-整数",
			input:    float64(456.0),
			expected: "456",
			wantErr:  false,
		},
		{
			name:     "float64类型-小数",
			input:    123.45,
			expected: "123.45",
			wantErr:  false,
		},
		{
			name:     "bool类型-true",
			input:    true,
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "bool类型-false",
			input:    false,
			expected: "false",
			wantErr:  false,
		},
		{
			name:     "不支持的类型",
			input:    []int{1, 2, 3},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafelyConvertToString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafelyConvertToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("SafelyConvertToString() = %v, want %v", got, tt.expected)
			}
		})
	}
}
