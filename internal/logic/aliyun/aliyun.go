package aliyun

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/iimeta/iim-sdk/internal/consts"
	m "github.com/iimeta/iim-sdk/internal/model"
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

func (s *sAliyun) Text(ctx context.Context, userId int, model, prompt string) (*m.Text, error) {

	messages := make([]sdk.QwenChatCompletionMessage, 0)

	reply, err := g.Redis().LRange(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, model, userId), 0, -1)
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

	qwenChatCompletionMessage := sdk.QwenChatCompletionMessage{
		User: prompt,
	}

	b, err := json.Marshal(qwenChatCompletionMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	logger.Infof(ctx, "qwenChatCompletionMessage: %s", string(b))

	messages = append(messages, qwenChatCompletionMessage)

	response, err := sdk.QwenChatCompletion(ctx, model, messages)

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

	_, err = g.Redis().RPush(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, model, userId), b)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	return &m.Text{
		Content: content,
	}, nil
}
