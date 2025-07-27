package entities

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// VariableType 变量的类型枚举
type VariableType string

const (
	VariableTypeString      VariableType = "string"
	VariableTypeInt         VariableType = "int"
	VariableTypeFloat       VariableType = "float"
	VariableTypeBoolean     VariableType = "boolean"
	VariableTypeListString  VariableType = "list[string]"
	VariableTypeListInt     VariableType = "list[int]"
	VariableTypeListFloat   VariableType = "list[float]"
	VariableTypeListBoolean VariableType = "list[boolean]"
)

// VariableValueType 变量内置值类型枚举
type VariableValueType string

const (
	VariableValueTypeRef       VariableValueType = "ref"       // 引用类型
	VariableValueTypeLiteral   VariableValueType = "literal"   // 字面数据/直接输入
	VariableValueTypeGenerated VariableValueType = "generated" // 生成的值，一般用在开始节点或者output中
)

// 变量名字正则匹配规则
var VariableNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// 描述最大长度
const VariableDescriptionMaxLength = 1024

// VariableTypeMap 变量类型与Go类型的映射
var VariableTypeMap = map[VariableType]any{
	VariableTypeString:      "",
	VariableTypeInt:         0,
	VariableTypeFloat:       0.0,
	VariableTypeBoolean:     false,
	VariableTypeListString:  []string{},
	VariableTypeListInt:     []int{},
	VariableTypeListFloat:   []float64{},
	VariableTypeListBoolean: []bool{},
}

// VariableTypeDefaultValueMap 变量类型默认值映射
var VariableTypeDefaultValueMap = map[VariableType]any{
	VariableTypeString:      "",
	VariableTypeInt:         0,
	VariableTypeFloat:       0.0,
	VariableTypeBoolean:     false,
	VariableTypeListString:  []string{},
	VariableTypeListInt:     []int{},
	VariableTypeListFloat:   []float64{},
	VariableTypeListBoolean: []bool{},
}

// VariableContent 变量内容实体信息，如果类型为引用，则使用content记录引用节点id+引用节点的变量名
type VariableContent struct {
	RefNodeID  *uuid.UUID `json:"ref_node_id,omitempty"`
	RefVarName string     `json:"ref_var_name"`
}

// VariableValue 变量的实体值信息
type VariableValue struct {
	Type    VariableValueType `json:"type"`
	Content any               `json:"content"`
}

// VariableEntity 变量实体信息
type VariableEntity struct {
	Name        string         `json:"name"`        // 变量的名字
	Description string         `json:"description"` // 变量的描述信息
	Required    bool           `json:"required"`    // 变量是否必填
	Type        VariableType   `json:"type"`        // 变量的类型
	Value       VariableValue  `json:"value"`       // 变量对应的值
	Meta        map[string]any `json:"meta"`        // 变量元数据，存储一些额外的信息
}

// NewVariableEntity 创建新的变量实体
func NewVariableEntity() *VariableEntity {
	return &VariableEntity{
		Required: true,
		Type:     VariableTypeString,
		Value: VariableValue{
			Type:    VariableValueTypeLiteral,
			Content: "",
		},
		Meta: make(map[string]any),
	}
}

// ValidateName 校验变量名字
func (v *VariableEntity) ValidateName() error {
	if !VariableNamePattern.MatchString(v.Name) {
		return fmt.Errorf("变量名字仅支持字母、数字和下划线，且以字母/下划线为开头")
	}
	return nil
}

// ValidateDescription 校验描述信息，截取前1024个字符
func (v *VariableEntity) ValidateDescription() {
	if len(v.Description) > VariableDescriptionMaxLength {
		v.Description = v.Description[:VariableDescriptionMaxLength]
	}
}

// Validate 校验变量实体
func (v *VariableEntity) Validate() error {
	if err := v.ValidateName(); err != nil {
		return err
	}
	v.ValidateDescription()
	return nil
}
