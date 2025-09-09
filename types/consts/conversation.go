package consts

// Conversation相关常量定义

// SummarizerTemplate 摘要汇总模板
const SummarizerTemplate = `逐步总结提供的对话内容，在之前的总结基础上继续添加并返回一个新的总结，并确保新总结的长度不要超过2000个字符，必要的时候可以删除一些信息，尽可能简洁。

EXAMPLE
当前总结:
人类询问 AI 对人工智能的看法。AI 认为人工智能是一股向善的力量。

新的会话:
Human: 为什么你认为人工智能是一股向善的力量？
AI: 因为人工智能将帮助人类发挥他们全部的潜力。

新的总结:
人类询问AI对人工智能的看法，AI认为人工智能是一股向善的力量，因为它将帮助人类发挥全部潜力。
END OF EXAMPLE

当前总结:
{summary}

新的会话:
{new_lines}

新的总结:`

// ConversationNameTemplate 会话名字提示模板
const ConversationNameTemplate = "请从用户传递的内容中提取出对应的主题"

// ConversationInfoPrompt 会话信息提取提示词
const ConversationInfoPrompt = `你需要将用户的输入分解为"主题"和"意图"，以便准确识别用户输入的类型。
注意：用户的语言可能是多样性的，可以是英文、中文、日语、法语等。
确保你的输出与用户的语言尽可能一致并简短！

示例1：
用户输入: hi, my name is LiHua.
{
    "language_type": "用户的输入是纯英文",
    "reasoning": "输出语言必须是英文",
    "subject": "Users greet me"
}

示例2:
用户输入: hello
{
    "language_type": "用户的输入是纯英文",
    "reasoning": "输出语言必须是英文",
    "subject": "Greeting myself"
}

示例3:
用户输入: www.imooc.com讲了什么
{
    "language_type": "用户输入是中英文混合",
    "reasoning": "英文部分是URL，主要意图还是使用中文表达的，所以输出语言必须是中文",
    "subject": "询问网站www.imooc.com"
}

示例4:
用户输入: why小红的年龄is老than小明?
{
    "language_type": "用户输入是中英文混合",
    "reasoning": "英文部分是口语化输入，主要意图是中文，且中文占据更大的实际意义，所以输出语言必须是中文",
    "subject": "询问小红和小明的年龄"
}

示例5:
用户输入: yo, 你今天怎么样?
{
    "language_type": "用户输入是中英文混合",
    "reasoning": "英文部分是口语化输入，主要意图是中文，所以输出语言必须是中文",
    "subject": "询问我今天的状态"
}`

// SuggestedQuestionsTemplate 建议问题提示词模板
const SuggestedQuestionsTemplate = "请根据传递的历史信息预测人类最后可能会问的三个问题"

// SuggestedQuestionsPrompt 建议问题生成提示词
const SuggestedQuestionsPrompt = `请帮我预测人类最可能会问的三个问题，并且每个问题都保持在50个字符以内。
生成的内容必须是指定模式的JSON格式数组: ["问题1", "问题2", "问题3"]`

// InvokeFrom 会话调用来源
type InvokeFrom string

const (
	InvokeFromServiceAPI     InvokeFrom = "service_api"     // 开放api服务调用
	InvokeFromWebApp         InvokeFrom = "web_app"         // web应用
	InvokeFromDebugger       InvokeFrom = "debugger"        // 调试页面
	InvokeFromAssistantAgent InvokeFrom = "assistant_agent" // 辅助Agent调用
	InvokeFromEndUser        InvokeFrom = "end_user"
)

// MessageStatus 会话状态
type MessageStatus string

const (
	MessageStatusNormal  MessageStatus = "normal"  // 正常
	MessageStatusStop    MessageStatus = "stop"    // 停止
	MessageStatusTimeout MessageStatus = "timeout" // 超时
	MessageStatusError   MessageStatus = "error"   // 出错
)

func (ms MessageStatus) String() string {
	return string(ms)
}
