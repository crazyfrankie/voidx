package gaode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GaodeWeatherRequest 高德天气查询请求参数
type GaodeWeatherRequest struct {
	City string `json:"city" jsonschema:"description=需要查询天气预报的目标城市，例如：广州"`
}

// GaodeWeatherResponse 高德天气查询响应
type GaodeWeatherResponse struct {
	Success bool   `json:"success"`
	Weather string `json:"weather,omitempty"`
	Message string `json:"message"`
}

// CityResponse 城市查询响应
type CityResponse struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Districts []struct {
		AdCode string `json:"adcode"`
		Name   string `json:"name"`
	} `json:"districts"`
}

// WeatherResponse 天气查询响应
type WeatherResponse struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Forecasts []struct {
		City     string `json:"city"`
		Province string `json:"province"`
		Casts    []struct {
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

// gaodeWeatherTool 高德天气预报工具实现
func gaodeWeatherTool(ctx context.Context, req GaodeWeatherRequest) (GaodeWeatherResponse, error) {
	apiKey := os.Getenv("GAODE_API_KEY")
	if apiKey == "" {
		return GaodeWeatherResponse{
			Success: false,
			Message: "高德开放平台API未配置",
		}, nil
	}

	if req.City == "" {
		return GaodeWeatherResponse{
			Success: false,
			Message: "城市参数不能为空",
		}, nil
	}

	weather, err := getWeather(ctx, req.City, apiKey)
	if err != nil {
		return GaodeWeatherResponse{
			Success: false,
			Message: fmt.Sprintf("获取%s天气预报信息失败: %v", req.City, err),
		}, nil
	}

	return GaodeWeatherResponse{
		Success: true,
		Weather: weather,
		Message: fmt.Sprintf("成功获取%s的天气预报", req.City),
	}, nil
}

// getWeather 获取天气信息
func getWeather(ctx context.Context, city, apiKey string) (string, error) {
	// 1. 获取城市编码
	adCode, err := getCityAdCode(ctx, city, apiKey)
	if err != nil {
		return "", err
	}

	// 2. 根据城市编码获取天气信息
	weatherURL := fmt.Sprintf("https://restapi.amap.com/v3/weather/weatherInfo?key=%s&city=%s&extensions=all",
		apiKey, adCode)

	req, err := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var weatherResp WeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return "", err
	}

	if weatherResp.Info != "OK" {
		return "", fmt.Errorf("weather API error: %s", weatherResp.Info)
	}

	// 格式化天气信息
	if len(weatherResp.Forecasts) == 0 || len(weatherResp.Forecasts[0].Casts) == 0 {
		return fmt.Sprintf("%s暂无天气预报信息", city), nil
	}

	forecast := weatherResp.Forecasts[0]
	result := fmt.Sprintf("城市: %s, %s\n\n", forecast.City, forecast.Province)

	for i, cast := range forecast.Casts {
		if i >= 3 { // 只显示前3天
			break
		}
		result += fmt.Sprintf("日期: %s (%s)\n", cast.Date, cast.Week)
		result += fmt.Sprintf("白天: %s, %s°C, %s%s级\n",
			cast.DayWeather, cast.DayTemp, cast.DayWind, cast.DayPower)
		result += fmt.Sprintf("夜间: %s, %s°C, %s%s级\n\n",
			cast.NightWeather, cast.NightTemp, cast.NightWind, cast.NightPower)
	}

	return result, nil
}

// getCityAdCode 获取城市行政区域编码
func getCityAdCode(ctx context.Context, city, apiKey string) (string, error) {
	cityURL := fmt.Sprintf("https://restapi.amap.com/v3/config/district?key=%s&keywords=%s&subdistrict=0",
		apiKey, url.QueryEscape(city))

	req, err := http.NewRequestWithContext(ctx, "GET", cityURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var cityResp CityResponse
	if err := json.Unmarshal(body, &cityResp); err != nil {
		return "", err
	}

	if cityResp.Info != "OK" || len(cityResp.Districts) == 0 {
		return "", fmt.Errorf("city not found: %s", city)
	}

	return cityResp.Districts[0].AdCode, nil
}

// NewGaodeWeatherTool 创建高德天气预报工具
func NewGaodeWeatherTool() (tool.InvokableTool, error) {
	return utils.InferTool("gaode_weather", "当你想查询天气或者与天气相关的问题时可以使用的工具", gaodeWeatherTool)
}
