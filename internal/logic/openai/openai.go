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

func (s *sOpenAI) Text(ctx context.Context, senderId, receiverId, talkType int, text, model string, isOpenContext int, mentions ...string) (string, error) {

	if talkType == 2 {
		content := gstr.Split(text, " ")
		if len(content) > 1 {
			text = content[1]
		} else {
			content = gstr.Split(text, " ")
			if len(content) > 1 {
				text = content[1]
			}
		}
	}

	if len(text) == 0 {
		return "", nil
	}

	messages := make([]openai.ChatCompletionMessage, 0)

	// 开启上下文
	if isOpenContext == 0 {

		reply, err := redis.LRange(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), 0, -1)
		if err != nil {
			logger.Error(ctx, err)
			return "", err
		}

		messagesStr := reply.Strings()
		if len(messagesStr) == 0 {
			b, err := gjson.Marshal(sdk.ChatMessageRoleSystem)
			if err != nil {
				logger.Error(ctx, err)
				return "", err
			}
			_, err = redis.RPush(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), b)
			if err != nil {
				logger.Error(ctx, err)
				return "", err
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
					return "", err
				}
				_, err = redis.LPush(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), b)
				if err != nil {
					logger.Error(ctx, err)
					return "", err
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
		Content: text,
	}

	b, err := gjson.Marshal(chatCompletionMessage)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	logger.Infof(ctx, "chatCompletionMessage: %s", string(b))

	messages = append(messages, chatCompletionMessage)

	response, err := sdk.ChatGPTChatCompletion(ctx, model, messages)

	if err != nil {
		logger.Error(ctx, err)

		if gstr.Contains(err.Error(), "Please reduce the length of the messages") {
			start := int64(len(messages) / 2)
			if start > 1 {
				err = redis.LTrim(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), start, -1)
				if err != nil {
					logger.Error(ctx, err)
					return "", err
				} else {
					return s.Text(ctx, senderId, receiverId, talkType, text, model, isOpenContext, mentions...)
				}
			}
		}

		return "", err
	}

	_, err = redis.RPush(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), b)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	content := response.Choices[0].Message.Content

	chatCompletionMessage = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	}

	b, err = json.Marshal(chatCompletionMessage)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	_, err = redis.RPush(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), b)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	if talkType == 2 {
		for i, mention := range mentions {
			if i == 0 {
				content += "\n"
			} else {
				content += " "
			}
			content += "@" + mention
		}
	}

	return content, err
}

func (s *sOpenAI) Image(ctx context.Context, senderId, receiverId, talkType int, text string, mentions ...string) (*model.Image, error) {

	if talkType == 2 {
		content := gstr.Split(text, " ")
		if len(content) > 1 {
			text = content[1]
		}
	}

	if len(text) == 0 {
		return nil, nil
	}

	logger.Infof(ctx, "Image text: %s", text)

	imgBase64, err := sdk.GenImageBase64(ctx, text)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	imgBytes, err := base64.StdEncoding.DecodeString(imgBase64)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	imageInfo, err := service.File().SaveImage(ctx, imgBytes, ".png")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	domain, err := config.Get(ctx, "filesystem.local.domain")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	url := domain.String() + "/" + imageInfo.FilePath

	return &model.Image{
		Url:    url,
		Width:  imageInfo.Width,
		Height: imageInfo.Height,
		Size:   imageInfo.Size,
	}, nil
}
