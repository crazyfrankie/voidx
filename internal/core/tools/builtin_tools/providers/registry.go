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
	entities.RegisterTool("time", "current_time", time.CurrentTime)

	// Register google tools
	entities.RegisterTool("google", "google_serper", google.GoogleSerper)

	// Register duckduckgo tools
	entities.RegisterTool("duckduckgo", "duckduckgo_search", duckduckgo.DuckduckgoSearch)

	// Register dalle tools
	entities.RegisterTool("dalle", "dalle3", dalle.Dalle3)

	// Register gaode tools
	entities.RegisterTool("gaode", "gaode_weather", gaode.GaodeWeather)

	// Register wikipedia tools
	entities.RegisterTool("wikipedia", "wikipedia_search", wikipedia.WikipediaSearch)

	// Register pptx tools
	entities.RegisterTool("pptx", "markdown_to_pptx", pptx.MarkdownToPptx)
}
