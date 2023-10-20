package xfyun

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/sdk"
	"github.com/sashabaranov/go-openai"
)

type sXfyun struct{}

func init() {
	service.RegisterXfyun(New())
}

func New() service.IXfyun {
	return &sXfyun{}
}

func (s *sXfyun) Text(ctx context.Context, robot *model.Robot, message *model.Message) (*model.Text, error) {

	messages := make([]sdk.Text, 0)

	if message.IsWithContext {

		contexts := service.Common().GetMessageContext(ctx, robot, message)

		for _, context := range contexts {
			textMessage := sdk.Text{}
			if err := json.Unmarshal([]byte(context), &textMessage); err != nil {
				logger.Error(ctx, err)
				continue
			}
			if textMessage.Role != openai.ChatMessageRoleSystem {
				messages = append(messages, textMessage)
			}
		}
	}

	textMessage := sdk.Text{
		Role:    sdk.SparkMessageRoleUser,
		Content: message.Prompt,
	}

	messages = append(messages, textMessage)

	response, err := sdk.SparkChat(ctx, robot.Model, fmt.Sprintf("%v", message.Sid), messages)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	err = service.Common().SaveMessageContext(ctx, robot, message, textMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	content := response.Content

	textMessage = sdk.Text{
		Role:    sdk.SparkMessageRoleAssistant,
		Content: content,
	}

	err = service.Common().SaveMessageContext(ctx, robot, message, textMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	return &model.Text{
		Content: content,
		Usage: &model.Usage{
			PromptTokens:     response.Payload.Usage.Text.PromptTokens,
			CompletionTokens: response.Payload.Usage.Text.CompletionTokens,
			TotalTokens:      response.Payload.Usage.Text.TotalTokens,
		},
	}, nil
}
