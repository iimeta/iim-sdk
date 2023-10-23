package baidu

import (
	"context"
	"encoding/json"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/sdk"
	"github.com/sashabaranov/go-openai"
)

type sBaidu struct{}

func init() {
	service.RegisterBaidu(New())
}

func New() service.IBaidu {
	return &sBaidu{}
}

func (s *sBaidu) Text(ctx context.Context, robot *model.Robot, message *model.Message) (*model.Text, error) {

	messages := make([]sdk.ErnieBotMessage, 0)

	if message.IsWithContext {

		contexts := service.Common().GetMessageContext(ctx, robot, message)

		for _, context := range contexts {
			ernieBotMessage := sdk.ErnieBotMessage{}
			if err := json.Unmarshal([]byte(context), &ernieBotMessage); err != nil {
				logger.Error(ctx, err)
				continue
			}
			if ernieBotMessage.Role != openai.ChatMessageRoleSystem {
				messages = append(messages, ernieBotMessage)
			}
		}
	}

	ernieBotMessage := sdk.ErnieBotMessage{
		Role:    sdk.RoleUser,
		Content: message.Prompt,
	}

	messages = append(messages, ernieBotMessage)

	response, err := sdk.ErnieBot(ctx, robot.Model, messages)

	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	err = service.Common().SaveMessageContext(ctx, robot, message, ernieBotMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	content := response.Result

	ernieBotMessage = sdk.ErnieBotMessage{
		Role:    sdk.RoleAssistant,
		Content: content,
	}

	err = service.Common().SaveMessageContext(ctx, robot, message, ernieBotMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	return &model.Text{
		Content: content,
		Usage: &model.Usage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}, nil
}
