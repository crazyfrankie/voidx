package entity

// VariableType 变量类型枚举
type VariableType string

const (
	VariableTypeString  VariableType = "string"  // 字符串类型
	VariableTypeNumber  VariableType = "number"  // 数字类型
	VariableTypeBoolean VariableType = "boolean" // 布尔类型
	VariableTypeArray   VariableType = "array"   // 数组类型
	VariableTypeObject  VariableType = "object"  // 对象类型
	VariableTypeAny     VariableType = "any"     // 任意类型
)

// IsValid 检查变量类型是否有效
func (t VariableType) IsValid() bool {
	switch t {
	case VariableTypeString, VariableTypeNumber, VariableTypeBoolean,
		VariableTypeArray, VariableTypeObject, VariableTypeAny:
		return true
	default:
		return false
	}
}

// String 返回变量类型的字符串表示
func (t VariableType) String() string {
	return string(t)
}

// GetDefaultValue 获取变量类型的默认值
func (t VariableType) GetDefaultValue() any {
	switch t {
	case VariableTypeString:
		return ""
	case VariableTypeNumber:
		return 0
	case VariableTypeBoolean:
		return false
	case VariableTypeArray:
		return []any{}
	case VariableTypeObject:
		return map[string]any{}
	case VariableTypeAny:
		return nil
	default:
		return nil
	}
}
