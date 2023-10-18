package robot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/iimeta/iim-sdk/internal/consts"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/sdk"
	"github.com/sashabaranov/go-openai"
)

type spark struct{}

var Spark *spark

func init() {
	Spark = &spark{}
}

func (o *spark) Chat(ctx context.Context, senderId, receiverId, talkType int, text, model string, mentions ...string) (string, error) {

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

	messages := make([]sdk.Text, 0)

	reply, err := g.Redis().LRange(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), 0, -1)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
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

	textMessage := sdk.Text{
		Role:    sdk.SparkMessageRoleUser,
		Content: text,
	}

	b, err := json.Marshal(textMessage)
	if err != nil {
		logger.Error(ctx, err)
	}

	logger.Infof(ctx, "textMessage: %s", string(b))

	messages = append(messages, textMessage)

	response, err := sdk.SparkChat(ctx, model, fmt.Sprintf("%d", receiverId), messages)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	_, err = g.Redis().RPush(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), b)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	content := response

	textMessage = sdk.Text{
		Role:    sdk.SparkMessageRoleAssistant,
		Content: content,
	}

	b, err = json.Marshal(textMessage)
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

	return content, nil
}
