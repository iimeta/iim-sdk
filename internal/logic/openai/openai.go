package openai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/iimeta/iim-sdk/internal/config"
	"github.com/iimeta/iim-sdk/internal/consts"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/redis"
	"github.com/iimeta/iim-sdk/utility/sdk"
	"github.com/sashabaranov/go-openai"
)

type sOpenAI struct{}

func init() {
	service.RegisterOpenAI(New())
}

func New() service.IOpenAI {
	return &sOpenAI{}
}

func (s *sOpenAI) Text(ctx context.Context, userId int, message *model.Message) (*model.Text, error) {

	messages := make([]openai.ChatCompletionMessage, 0)

	if message.IsWithContext {

		reply, err := redis.LRange(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), 0, -1)
		if err != nil {
			logger.Error(ctx, err)
			return nil, err
		}

		messagesStr := reply.Strings()
		if len(messagesStr) == 0 {
			b, err := gjson.Marshal(sdk.ChatMessageRoleSystem)
			if err != nil {
				logger.Error(ctx, err)
				return nil, err
			}
			_, err = redis.RPush(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), b)
			if err != nil {
				logger.Error(ctx, err)
				return nil, err
			}
			messages = append(messages, sdk.ChatMessageRoleSystem)
		}

		for i, str := range messagesStr {

			chatCompletionMessage := openai.ChatCompletionMessage{}
			if err := gjson.Unmarshal([]byte(str), &chatCompletionMessage); err != nil {
				logger.Error(ctx, err)
				continue
			}

			if i == 0 && chatCompletionMessage.Role != openai.ChatMessageRoleSystem {
				b, err := gjson.Marshal(sdk.ChatMessageRoleSystem)
				if err != nil {
					logger.Error(ctx, err)
					return nil, err
				}
				_, err = redis.LPush(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), b)
				if err != nil {
					logger.Error(ctx, err)
					return nil, err
				}
				messages = append(messages, sdk.ChatMessageRoleSystem)
			}

			messages = append(messages, chatCompletionMessage)
		}
	} else {
		messages = append(messages, sdk.ChatMessageRoleSystem)
	}

	chatCompletionMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message.Prompt,
	}

	b, err := gjson.Marshal(chatCompletionMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	logger.Infof(ctx, "chatCompletionMessage: %s", string(b))

	messages = append(messages, chatCompletionMessage)

	response, err := sdk.ChatGPTChatCompletion(ctx, message.Model, messages)

	if err != nil {
		logger.Error(ctx, err)

		if gstr.Contains(err.Error(), "Please reduce the length of the messages") {
			start := int64(len(messages) / 2)
			if start > 1 {
				err = redis.LTrim(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), start, -1)
				if err != nil {
					logger.Error(ctx, err)
					return nil, err
				} else {
					return s.Text(ctx, userId, message)
				}
			}
		}

		return nil, err
	}

	_, err = redis.RPush(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), b)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	content := response.Choices[0].Message.Content

	chatCompletionMessage = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	}

	b, err = json.Marshal(chatCompletionMessage)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	_, err = redis.RPush(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_PREFIX_KEY, message.Corp, message.ModelType, userId), b)
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

func (s *sOpenAI) Image(ctx context.Context, userId int, message *model.Message) (imageInfo *model.Image, err error) {

	imgBase64, err := sdk.GenImageBase64(ctx, message.Prompt)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	imgBytes, err := base64.StdEncoding.DecodeString(imgBase64)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	imageInfo, err = service.File().SaveImage(ctx, imgBytes, ".png")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	domain, err := config.Get(ctx, "filesystem.local.domain")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	imageInfo.Url = domain.String() + "/" + imageInfo.FilePath

	return imageInfo, nil
}
