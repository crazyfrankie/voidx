package providers

import (
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/entities"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/dalle"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/duckduckgo"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/gaode"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/google"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/pptx"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/time"
	"github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers/wikipedia"
)

// init registers all tool functions
func init() {
	// Register time tools
	entities.RegisterTool("time", "current_time", entities.NewFuncTool("current_time", "一个用于获取当前时间的工具", time.CurrentTime))

	// Register google tools
	entities.RegisterTool("google", "google_serper", entities.NewFuncTool("google_serper", "这是一个低成本的谷歌搜索API。当你需要搜索时事的时候，可以使用该工具，该工具的输入是一个查询语句", google.GoogleSerper))

	// Register duckduckgo tools
	entities.RegisterTool("duckduckgo", "duckduckgo_search", entities.NewFuncTool("duckduckgo_search", "一个注重隐私的搜索引擎", duckduckgo.DuckduckgoSearch))

	// Register dalle tools
	entities.RegisterTool("dalle", "dalle3", entities.NewFuncTool("dalle3", "DALLE-3是一个将文本转换成图片的绘图工具", dalle.Dalle3))

	// Register gaode tools
	entities.RegisterTool("gaode", "gaode_weather", entities.NewFuncTool("gaode_weather", "根据传递的城市查询该城市的天气预报信息", gaode.GaodeWeather))

	// Register wikipedia tools
	entities.RegisterTool("wikipedia", "wikipedia_search", entities.NewFuncTool("wikipedia_search", "一个用于执行维基百科搜索并提取片段和网页的工具", wikipedia.WikipediaSearch))

	// Register pptx tools
	entities.RegisterTool("pptx", "markdown_to_pptx", entities.NewFuncTool("markdown_to_pptx", "这是一款能将Markdown文档内容转换成本地PPT文件的工具", pptx.MarkdownToPptx))
}
