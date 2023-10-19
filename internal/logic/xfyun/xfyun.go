package xfyun

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/iimeta/iim-sdk/internal/consts"
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

func (s *sXfyun) Text(ctx context.Context, userId int, message *model.Message) (*model.Text, error) {

	messages := make([]sdk.Text, 0)

	if message.IsWithContext {
		reply, err := g.Redis().LRange(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), 0, -1)
		if err != nil {
			logger.Error(ctx, err)
			return nil, err
		}

		messagesStr := reply.Strings()

		for _, str := range messagesStr {
			textMessage := sdk.Text{}
			if err := json.Unmarshal([]byte(str), &textMessage); err != nil {
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

	b, err := json.Marshal(textMessage)
	if err != nil {
		logger.Error(ctx, err)
	}

	logger.Infof(ctx, "textMessage: %s", string(b))

	messages = append(messages, textMessage)

	response, err := sdk.SparkChat(ctx, message.Model, fmt.Sprintf("%d", userId), messages)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	_, err = g.Redis().RPush(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), b)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	content := response.Content

	textMessage = sdk.Text{
		Role:    sdk.SparkMessageRoleAssistant,
		Content: content,
	}

	b, err = json.Marshal(textMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	_, err = g.Redis().RPush(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), b)
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
