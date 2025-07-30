package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	"github.com/crazyfrankie/voidx/internal/audio/repository"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type AudioService struct {
	repo *repository.AudioRepo
}

func NewAudioService(repo *repository.AudioRepo) *AudioService {
	return &AudioService{repo: repo}
}

// AudioToText 将传递的语音转换成文本
func (s *AudioService) AudioToText(ctx context.Context, userID uuid.UUID, audioData []byte, filename string) (string, error) {
	// 1. 创建OpenAI客户端
	client := s.getOpenAIClient()

	// 2. 创建音频请求
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filename,
		Reader:   bytes.NewReader(audioData),
	}

	// 3. 调用Whisper服务转换语音
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		return "", errno.ErrInternalServer.AppendBizMessage(fmt.Errorf("语音转文本失败, %w", err))
	}

	return resp.Text, nil
}

// MessageToAudio 将消息转换成流式事件输出语音
func (s *AudioService) MessageToAudio(ctx context.Context, userID, messageID uuid.UUID) (<-chan resp.TTSEvent, error) {
	// 1. 根据消息ID获取消息并校验权限
	message, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("消息不存在"))
	}

	if message.CreatedBy != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该消息"))
	}

	if message.Answer == "" {
		return nil, errno.ErrValidate.AppendBizMessage(errors.New("消息内容为空"))
	}

	// 2. 获取会话信息
	conversation, err := s.repo.GetConversationByID(ctx, message.ConversationID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该会话"))
	}

	// 3. 检查TTS配置
	ttsConfig, err := s.getTTSConfig(ctx, message, conversation)
	if err != nil {
		return nil, err
	}

	if !ttsConfig.Enable {
		return nil, errno.ErrValidate.AppendBizMessage(errors.New("该应用未开启文字转语音功能"))
	}

	// 4. 创建事件通道
	eventChan := make(chan resp.TTSEvent, 100)

	// 5. 启动异步TTS处理
	go s.processTTS(ctx, message, conversation, ttsConfig, eventChan)

	return eventChan, nil
}

// TTSConfig TTS配置
type TTSConfig struct {
	Enable bool   `json:"enable"`
	Voice  string `json:"voice"`
}

// getTTSConfig 获取TTS配置
func (s *AudioService) getTTSConfig(ctx context.Context, message *entity.Message, conversation *entity.Conversation) (*TTSConfig, error) {
	// 默认配置
	config := &TTSConfig{
		Enable: true,
		Voice:  "echo",
	}

	// 根据调用来源获取不同的配置
	switch message.InvokeFrom {
	case "web_app", "debugger":
		app, err := s.repo.GetAppByID(ctx, conversation.AppID)
		if err != nil {
			return nil, errno.ErrNotFound.AppendBizMessage(errors.New("应用不存在"))
		}

		// 权限校验
		if message.InvokeFrom == "debugger" && app.AccountID != message.CreatedBy {
			return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该应用"))
		}

		if message.InvokeFrom == "web_app" && app.Status != "published" {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("应用未发布"))
		}

		// 获取应用配置
		appConfig, err := s.repo.GetAppConfig(ctx, app.ID, message.InvokeFrom == "debugger")
		if err != nil {
			return nil, err
		}

		// 解析TTS配置
		if ttsData, ok := appConfig["text_to_speech"].(map[string]any); ok {
			if enable, ok := ttsData["enable"].(bool); ok {
				config.Enable = enable
			}
			if voice, ok := ttsData["voice"].(string); ok && voice != "" {
				config.Voice = voice
			}
		}

	case "service_api":
		return nil, errno.ErrValidate.AppendBizMessage(errors.New("开放API消息不支持文本转语音服务"))
	}

	return config, nil
}

// processTTS 处理TTS转换
func (s *AudioService) processTTS(ctx context.Context, message *entity.Message, conversation *entity.Conversation, config *TTSConfig, eventChan chan<- resp.TTSEvent) {
	defer close(eventChan)

	// 1. 创建OpenAI客户端
	client := s.getOpenAIClient()

	// 2. 创建TTS请求
	req := openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Input:          message.Answer,
		Voice:          openai.SpeechVoice(config.Voice),
		ResponseFormat: openai.SpeechResponseFormatMp3,
	}

	// 3. 调用TTS服务
	response, err := client.CreateSpeech(ctx, req)
	if err != nil {
		// 发送错误事件
		eventChan <- resp.TTSEvent{
			ConversationID: conversation.ID.String(),
			MessageID:      message.ID.String(),
			Audio:          "", // 空音频表示错误
		}
		return
	}
	defer response.Close()

	// 4. 流式读取音频数据并发送事件
	buffer := make([]byte, 1024)
	for {
		n, err := response.Read(buffer)
		if n > 0 {
			// 将音频数据编码为base64并发送
			audioData := base64.StdEncoding.EncodeToString(buffer[:n])
			select {
			case eventChan <- resp.TTSEvent{
				ConversationID: conversation.ID.String(),
				MessageID:      message.ID.String(),
				Audio:          audioData,
			}:
			case <-ctx.Done():
				return
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
	}

	// 5. 发送结束事件
	eventChan <- resp.TTSEvent{
		ConversationID: conversation.ID.String(),
		MessageID:      message.ID.String(),
		Audio:          "", // 空音频表示结束
	}
}

// getOpenAIClient 获取OpenAI客户端
func (s *AudioService) getOpenAIClient() *openai.Client {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	if baseURL := os.Getenv("OPENAI_API_BASE"); baseURL != "" {
		config.BaseURL = baseURL
	}
	return openai.NewClientWithConfig(config)
}
