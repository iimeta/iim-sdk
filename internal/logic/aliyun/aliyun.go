package aliyun

import (
	"context"
	"encoding/json"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/sdk"
)

type sAliyun struct{}

func init() {
	service.RegisterAliyun(New())
}

func New() service.IAliyun {
	return &sAliyun{}
}

func (s *sAliyun) Text(ctx context.Context, robot *model.Robot, message *model.Message) (*model.Text, error) {

	messages := make([]sdk.QwenChatCompletionMessage, 0)

	if message.IsWithContext {

		contexts := service.Common().GetMessageContext(ctx, robot, message)

		for _, context := range contexts {
			qwenChatCompletionMessage := sdk.QwenChatCompletionMessage{}
			if err := json.Unmarshal([]byte(context), &qwenChatCompletionMessage); err != nil {
				logger.Error(ctx, err)
				continue
			}
			messages = append(messages, qwenChatCompletionMessage)
		}
	}

	qwenChatCompletionMessage := sdk.QwenChatCompletionMessage{
		User: message.Prompt,
	}

	b, err := json.Marshal(qwenChatCompletionMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	logger.Infof(ctx, "qwenChatCompletionMessage: %s", string(b))

	messages = append(messages, qwenChatCompletionMessage)

	response, err := sdk.QwenChatCompletion(ctx, robot.Model, messages)

	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	content := response.Output.Text

	qwenChatCompletionMessage.Bot = content

	err = service.Common().SaveMessageContext(ctx, robot, message, qwenChatCompletionMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	return &model.Text{
		Content: content,
		Usage: &model.Usage{
			PromptTokens:     response.Usage.InputTokens,
			CompletionTokens: response.Usage.OutputTokens,
			TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
		},
	}, nil
}
