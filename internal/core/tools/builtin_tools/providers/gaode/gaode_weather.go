package gaode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// GaodeWeatherTool represents a tool for Gaode weather query
type GaodeWeatherTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	APIKey      string `json:"-"`
}

// CityResponse represents the response from city query API
type CityResponse struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Districts []struct {
		AdCode string `json:"adcode"`
		Name   string `json:"name"`
	} `json:"districts"`
}

// WeatherResponse represents the response from weather API
type WeatherResponse struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Forecasts []struct {
		City       string `json:"city"`
		AdCode     string `json:"adcode"`
		Province   string `json:"province"`
		ReportTime string `json:"reporttime"`
		Casts      []struct {
			Date         string `json:"date"`
			Week         string `json:"week"`
			DayWeather   string `json:"dayweather"`
			NightWeather string `json:"nightweather"`
			DayTemp      string `json:"daytemp"`
			NightTemp    string `json:"nighttemp"`
			DayWind      string `json:"daywind"`
			NightWind    string `json:"nightwind"`
			DayPower     string `json:"daypower"`
			NightPower   string `json:"nightpower"`
		} `json:"casts"`
	} `json:"forecasts"`
}

// NewGaodeWeatherTool creates a new GaodeWeatherTool instance
func NewGaodeWeatherTool() *GaodeWeatherTool {
	return &GaodeWeatherTool{
		Name:        "gaode_weather",
		Description: "当你想查询天气或者与天气相关的问题时可以使用的工具",
		APIKey:      os.Getenv("GAODE_API_KEY"),
	}
}

// Run executes the Gaode weather query
func (t *GaodeWeatherTool) Run(args map[string]interface{}) (interface{}, error) {
	city, ok := args["city"].(string)
	if !ok {
		return nil, fmt.Errorf("city parameter is required and must be a string")
	}

	if t.APIKey == "" {
		return "高德开放平台API未配置", nil
	}

	// Step 1: Get city adcode
	cityURL := fmt.Sprintf("https://restapi.amap.com/v3/config/district?key=%s&keywords=%s&subdistrict=0",
		t.APIKey, url.QueryEscape(city))

	resp, err := http.Get(cityURL)
	if err != nil {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}

	var cityResp CityResponse
	if err := json.Unmarshal(body, &cityResp); err != nil {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}

	if cityResp.Info != "OK" || len(cityResp.Districts) == 0 {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}

	adCode := cityResp.Districts[0].AdCode

	// Step 2: Get weather information
	weatherURL := fmt.Sprintf("https://restapi.amap.com/v3/weather/weatherInfo?key=%s&city=%s&extensions=all",
		t.APIKey, adCode)

	resp, err = http.Get(weatherURL)
	if err != nil {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}

	var weatherResp WeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}

	if weatherResp.Info != "OK" || len(weatherResp.Forecasts) == 0 {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}

	// Format the weather information
	forecast := weatherResp.Forecasts[0]
	result := map[string]interface{}{
		"city":        forecast.City,
		"province":    forecast.Province,
		"report_time": forecast.ReportTime,
		"forecasts":   forecast.Casts,
	}

	// Convert to JSON string for consistency with Python version
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf("获取%s天气预报信息失败", city), nil
	}

	return string(jsonResult), nil
}

// GaodeWeather is the exported function for dynamic loading
func GaodeWeather(args map[string]interface{}) (interface{}, error) {
	tool := NewGaodeWeatherTool()
	return tool.Run(args)
}
