package aliyun

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
)

type sAliyun struct{}

func init() {
	service.RegisterAliyun(New())
}

func New() service.IAliyun {
	return &sAliyun{}
}

func (s *sAliyun) Text(ctx context.Context, userId int, message *model.Message) (*model.Text, error) {

	messages := make([]sdk.QwenChatCompletionMessage, 0)

	if message.IsWithContext {
		reply, err := g.Redis().LRange(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), 0, -1)
		if err != nil {
			logger.Error(ctx, err)
			return nil, err
		}

		messagesStr := reply.Strings()

		for _, str := range messagesStr {
			qwenChatCompletionMessage := sdk.QwenChatCompletionMessage{}
			if err := json.Unmarshal([]byte(str), &qwenChatCompletionMessage); err != nil {
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

	response, err := sdk.QwenChatCompletion(ctx, message.Model, messages)

	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	content := response.Output.Text

	qwenChatCompletionMessage.Bot = content

	b, err = json.Marshal(qwenChatCompletionMessage)
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
			PromptTokens:     response.Usage.InputTokens,
			CompletionTokens: response.Usage.OutputTokens,
			TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
		},
	}, nil
}
