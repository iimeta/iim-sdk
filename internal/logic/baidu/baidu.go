package baidu

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

type sBaidu struct{}

func init() {
	service.RegisterBaidu(New())
}

func New() service.IBaidu {
	return &sBaidu{}
}

func (s *sBaidu) Text(ctx context.Context, userId int, message *model.Message) (*model.Text, error) {

	messages := make([]sdk.ErnieBotMessage, 0)

	if message.IsWithContext {
		reply, err := g.Redis().LRange(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), 0, -1)
		if err != nil {
			logger.Error(ctx, err)
			return nil, err
		}

		messagesStr := reply.Strings()

		for _, str := range messagesStr {
			ernieBotMessage := sdk.ErnieBotMessage{}
			if err := json.Unmarshal([]byte(str), &ernieBotMessage); err != nil {
				logger.Error(ctx, err)
				continue
			}
			if ernieBotMessage.Role != openai.ChatMessageRoleSystem {
				messages = append(messages, ernieBotMessage)
			}
		}
	}

	ernieBotMessage := sdk.ErnieBotMessage{
		Role:    sdk.ErnieBotMessageRoleUser,
		Content: message.Prompt,
	}

	b, err := json.Marshal(ernieBotMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	logger.Infof(ctx, "ernieBotMessage: %s", string(b))

	messages = append(messages, ernieBotMessage)

	response, err := sdk.ErnieBot(ctx, message.Model, messages)

	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	_, err = g.Redis().RPush(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), b)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	content := response.Result

	ernieBotMessage = sdk.ErnieBotMessage{
		Role:    sdk.ErnieBotMessageRoleAssistant,
		Content: content,
	}

	b, err = json.Marshal(ernieBotMessage)
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
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
	}, nil
}
