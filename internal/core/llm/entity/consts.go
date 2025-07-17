package entity

// DefaultModelParameterTemplate contains the default parameter templates
// used to reduce YAML configuration, standardized based on OpenAI's model interface
var DefaultModelParameterTemplate = map[DefaultModelParameterName]ModelParameter{
	// Temperature parameter template
	ParameterTemperature: {
		Name:      string(ParameterTemperature),
		Label:     "温度",
		Type:      ParameterTypeFloat,
		Help:      "温度控制随机性，较低的温度会导致较少的随机生成。随着温度接近零，模型将变得更确定，较高的温度会导致更多随机内容被生成",
		Required:  false,
		Default:   1.0,
		Min:       floatPtr(0.0),
		Max:       floatPtr(2.0),
		Precision: 2,
		Options:   []ModelParameterOption{},
	},
	// Top P nuclear sampling parameter
	ParameterTopP: {
		Name:      string(ParameterTopP),
		Label:     "Top P",
		Type:      ParameterTypeFloat,
		Help:      "通过核心采样控制多样性，0.5表示考虑了一半的所有可能性加权选项",
		Required:  false,
		Default:   0.0,
		Min:       floatPtr(0.0),
		Max:       floatPtr(1.0),
		Precision: 2,
		Options:   []ModelParameterOption{},
	},
	// Presence penalty parameter
	ParameterPresencePenalty: {
		Name:      string(ParameterPresencePenalty),
		Label:     "存在惩罚",
		Type:      ParameterTypeFloat,
		Help:      "对文本中已有的标记的对数概率施加惩罚",
		Required:  false,
		Default:   0.0,
		Min:       floatPtr(-2.0),
		Max:       floatPtr(2.0),
		Precision: 2,
		Options:   []ModelParameterOption{},
	},
	// Frequency penalty parameter
	ParameterFrequencyPenalty: {
		Name:      string(ParameterFrequencyPenalty),
		Label:     "频率惩罚",
		Type:      ParameterTypeFloat,
		Help:      "对文本中已有的标记的对数概率施加惩罚",
		Required:  false,
		Default:   0.0,
		Min:       floatPtr(-2.0),
		Max:       floatPtr(2.0),
		Precision: 2,
		Options:   []ModelParameterOption{},
	},
	// Max tokens parameter
	ParameterMaxTokens: {
		Name:      string(ParameterMaxTokens),
		Label:     "最大标记",
		Type:      ParameterTypeInt,
		Help:      "要生成的标记的最大数量，类型为整型",
		Required:  false,
		Default:   nil,
		Min:       floatPtr(1.0),
		Max:       floatPtr(16384.0),
		Precision: 0,
		Options:   []ModelParameterOption{},
	},
}

// floatPtr returns a pointer to a float64 value
func floatPtr(f float64) *float64 {
	return &f
}
