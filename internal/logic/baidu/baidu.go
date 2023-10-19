package baidu

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/iimeta/iim-sdk/internal/consts"
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

func (s *sBaidu) Text(ctx context.Context, senderId, receiverId, talkType int, text, model string, mentions ...string) (string, error) {

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

	messages := make([]sdk.ErnieBotMessage, 0)

	reply, err := g.Redis().LRange(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), 0, -1)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
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

	ernieBotMessage := sdk.ErnieBotMessage{
		Role:    sdk.ErnieBotMessageRoleUser,
		Content: text,
	}

	b, err := json.Marshal(ernieBotMessage)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	logger.Infof(ctx, "ernieBotMessage: %s", string(b))

	messages = append(messages, ernieBotMessage)

	response, err := sdk.ErnieBot(ctx, model, messages)

	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	_, err = g.Redis().RPush(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), b)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	content := response.Result

	ernieBotMessage = sdk.ErnieBotMessage{
		Role:    sdk.ErnieBotMessageRoleAssistant,
		Content: content,
	}

	b, err = json.Marshal(ernieBotMessage)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	_, err = g.Redis().RPush(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), b)
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
