package entities

import (
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"github.com/crazyfrankie/voidx/pkg/consts"
)

// AgentConfig represents the configuration for an agent
type AgentConfig struct {
	// UserID represents the unique identifier of the user
	UserID uuid.UUID `json:"user_id"`

	// InvokeFrom represents the source of invocation
	InvokeFrom consts.InvokeFrom `json:"invoke_from"`

	// MaxIterationCount represents the maximum number of iterations the agent can perform
	MaxIterationCount int `json:"max_iteration_count"`

	// SystemPrompt represents the system prompt template for the agent
	SystemPrompt string `json:"system_prompt"`

	// PresetPrompt represents the preset prompt for specific tasks
	PresetPrompt string `json:"preset_prompt"`

	// EnableLongTermMemory determines if long-term memory is enabled
	EnableLongTermMemory bool `json:"enable_long_term_memory"`

	// Tools represents the list of tools available to the agent
	Tools []tools.Tool `json:"tools"`

	// ReviewConfig represents the configuration for content review
	ReviewConfig ReviewConfig `json:"review_config"`
}

// ReviewConfig represents the configuration for content review
type ReviewConfig struct {
	Enable        bool                `json:"enable"`
	Keywords      []string            `json:"keywords"`
	InputsConfig  ReviewInputsConfig  `json:"inputs_config"`
	OutputsConfig ReviewOutputsConfig `json:"outputs_config"`
}

// ReviewInputsConfig represents the configuration for input review
type ReviewInputsConfig struct {
	Enable         bool   `json:"enable"`
	PresetResponse string `json:"preset_response"`
}

// ReviewOutputsConfig represents the configuration for output review
type ReviewOutputsConfig struct {
	Enable bool `json:"enable"`
}

// AgentState represents the current state of an agent
type AgentState struct {
	// TaskID represents the unique identifier for the current task
	TaskID uuid.UUID `json:"task_id"`

	// IterationCount represents the current iteration count
	IterationCount int `json:"iteration_count"`

	// History represents the conversation history
	History []llms.ChatMessage `json:"history"`

	// LongTermMemory represents the long-term memory content
	LongTermMemory string `json:"long_term_memory"`

	// Messages represents the current messages in processing
	Messages []llms.ChatMessage `json:"messages"`
}

// Constants for agent configuration
const (
	// DefaultMaxIterationCount represents the default maximum iteration count
	DefaultMaxIterationCount = 5

	// DatasetRetrievalToolName represents the name of the dataset retrieval tool
	DatasetRetrievalToolName = "dataset_retrieval"

	// MaxIterationResponse represents the response when max iterations are reached
	MaxIterationResponse = "当前Agent迭代次数已超过限制，请重试"
)

// System prompt templates
const (
	AgentSystemPromptTemplate = `你是一个高度定制的智能体应用，旨在为用户提供准确、专业的内容生成和问题解答，请严格遵守以下规则：

1.**预设任务执行**
  - 你需要基于用户提供的预设提示(PRESET-PROMPT)，按照要求生成特定内容，确保输出符合用户的预期和指引；

2.**工具调用和参数生成**
  - 当任务需要时，你可以调用绑定的外部工具(如知识库检索、计算工具等)，并生成符合任务需求的调用参数，确保工具使用的准确性和高效性；

3.**历史对话和长期记忆**
  - 你可以参考历史对话记录，结合经过摘要提取的长期记忆，以提供更加个性化和上下文相关的回复，这将有助于在连续对话中保持一致性，并提供更加精确的反馈；

4.**外部知识库检索**
  - 如果用户的问题超出当前的知识范围或需要额外补充，你可以调用recall_dataset(知识库检索工具)以获取外部信息，确保答案的完整性和正确性；

5.**高效性和简洁性**
  - 保持对用户需求的精准理解和高效响应，提供简洁且有效的答案，避免冗长或无关信息；
  
<预设提示>
{preset_prompt}
</预设提示>

<长期记忆>
{long_term_memory}
</长期记忆>`

	ReactAgentSystemPromptTemplate = `你是一个高度定制的智能体应用，旨在为用户提供准确、专业的内容生成和问题解答，请严格遵守以下规则：

1.**预设任务执行**
  - 你需要基于用户提供的预设提示(PRESET-PROMPT)，按照要求生成特定内容，确保输出符合用户的预期和指引；

2.**工具调用和参数生成**
  - 当任务需要时，你可以调用绑定的外部工具(如知识库检索、计算工具等)，并生成符合任务需求的调用参数，确保工具使用的准确性和高效性，如果不需要调用工具的时候，请不要返回任何工具调用相关的json信息，如果用户传递了多条消息，请不要在最终答案里重复生成工具调用参数；

3.**历史对话和长期记忆**
  - 你可以参考` + "`历史对话`" + `记录，结合经过摘要提取的` + "`长期记忆`" + `，以提供更加个性化和上下文相关的回复，这将有助于在连续对话中保持一致性，并提供更加精确的反馈；

4.**外部知识库检索**
  - 如果用户的问题超出当前的知识范围或需要额外补充，你可以调用` + "`recall_dataset(知识库检索工具)`" + `以获取外部信息，确保答案的完整性和正确性；

5.**高效性和简洁性**
  - 保持对用户需求的精准理解和高效响应，提供简洁且有效的答案，避免冗长或无关信息；

6.**工具调用**
  - Agent智能体应用还提供了工具调用，具体信息可以参考<工具描述>里的工具信息，工具调用参数请参考` + "`args`" + `中的信息描述。
  - 工具描述说明:
    - 示例: google_serper - 这是一个低成本的谷歌搜索API。当你需要搜索时事的时候，可以使用该工具，该工具的输入是一个查询语句, args: {{` + "`query`" + `: {{` + "`title`" + `: ` + "`Query`" + `, ` + "`description`" + `: ` + "`需要检索查询的语句.`" + `, ` + "`type`" + `: ` + "`string`" + `}}}}
    - 格式: 工具名称 - 工具描述, args: 工具参数信息字典
  - LLM生成的工具调用参数说明:
    - 示例: ` + "```json\\n{{\"name\": \"google_serper\", \"args\": {{\"query\": \"慕课网 AI课程\"}}}}\\n```" + `
    - 格式: ` + "```json\\n{{\"name\": 需要调用的工具名称, \"args\": 调用该工具的输入参数字典}}\\n```" + `
    - 要求:
      - 生成的内容必须是符合规范的json字符串，并且仅包含两个字段` + "`name`" + `和` + "`args`" + `，其中` + "`name`" + `代表工具的名称，` + "`args`" + `代表调用该工具传递的参数，如果没有参数则传递空字典` + "`{{}}`" + `。
      - 生成的内容必须以"` + "```json" + `"为开头，以"` + "```" + `"为结尾，前面和后面不要添加任何内容，避免代码解析出错。
      - 注意` + "`工具描述参数args`" + `和最终生成的` + "`工具调用参数args`" + `的区别，不要错误生成。
      - 如果不需要工具调用，则正常生成即可，程序会自动检测内容开头是否为"` + "```json" + `"进行判断
    - 正确示例:
      - ` + "```json\\n{{\"name\": \"google_serper\", \"args\": {{\"query\": \"慕课网 AI课程\"}}}}\\n```" + `
      - ` + "```json\\n{{\"name\": \"current_time\", \"args\": {{}}}}\\n```" + `
      - ` + "```json\\n{{\"name\": \"dalle\", \"args\": {{\"query\": \"一幅老爷爷爬山的图片\", \"size\": \"1024x1024\"}}}}\\n```" + `
    - 错误示例:
      - 错误原因(在最前的` + "```json" + `前生成了内容): 好的，我将调用工具进行搜索。\\n` + "```json\\n{{\"name\": \"google_serper\", \"args\": {{\"query\": \"慕课网 AI课程\"}}}}\\n```" + `
      - 错误原因(在最后的` + "```" + `后生成了内容): ` + "```json\\n{{\"name\": \"google_serper\", \"args\": {{\"query\": \"慕课网 AI课程\"}}}}\\n```" + `，我将准备调用工具，请稍等。
      - 错误原因(生成了json，但是不包含在"` + "```json" + `"和"` + "```" + `"内): {{` + "\"name\": \"current_time\", \"args\": {{}}" + `}}
      - 错误原因(将描述参数的内容填充到生成参数中): ` + "```json\\n{{\"name\": \"google_serper\", \"args\": {{\"query\": {{\"title\": \"Query\", \"description\": \"需要检索查询的语句.\", \"type\": \"string\"}}}}\n```" + `

<预设提示>
{preset_prompt}
</预设提示>

<长期记忆>
{long_term_memory}
</长期记忆>

<工具描述>
{tool_description}
</工具描述>`
)
