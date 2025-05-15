package utils

import (
	"fmt"
	"math"
	"reflect"
)

// SafelyConvertToInt 尝试将各种类型安全地转换为 int
// 支持 int, int32, int64, float32, float64, string 类型的转换
// 对于浮点数，要求必须是整数值（如 1.0），不接受带小数部分的值
// 对于字符串，尝试解析为整数或浮点数（浮点数必须是整数值）
func SafelyConvertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		if float32(int(v)) == v {
			return int(v), nil
		}
		return 0, NewParamError("number", "must be an integer value")
	case float64:
		if float64(int(v)) == v {
			return int(v), nil
		}
		return 0, NewParamError("number", "must be an integer value")
	case string:
		var intValue int
		if _, err := fmt.Sscanf(v, "%d", &intValue); err == nil {
			return intValue, nil
		}
		var floatValue float64
		if _, err := fmt.Sscanf(v, "%f", &floatValue); err == nil {
			if math.Floor(floatValue) == floatValue {
				return int(floatValue), nil
			}
		}
		return 0, NewParamError("number", "must be a valid integer")
	default:
		return 0, NewParamError("number", fmt.Sprintf("unsupported type: %v", reflect.TypeOf(value)))
	}
}

// SafelyConvertToString 尝试将各种类型安全地转换为 string
// 支持 int, int32, int64, float32, float64, string, bool 类型的转换
// 对于数字类型，直接转换为对应的字符串表示
// 对于布尔值，转换为 "true" 或 "false"
func SafelyConvertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int32:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case float32:
		return fmt.Sprintf("%g", v), nil
	case float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		return "", NewParamError("string", fmt.Sprintf("unsupported type: %v", reflect.TypeOf(value)))
	}
}
