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
)

type aliyun struct{}

var Aliyun *aliyun

func init() {
	Aliyun = &aliyun{}
}

func (o *aliyun) Chat(ctx context.Context, senderId, receiverId, talkType int, text, model string, mentions ...string) (string, error) {

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

	messages := make([]sdk.QwenChatCompletionMessage, 0)

	reply, err := g.Redis().LRange(ctx, fmt.Sprintf(consts.CHAT_MESSAGES_PREFIX_KEY, receiverId, senderId), 0, -1)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
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
		User: text,
	}

	b, err := json.Marshal(qwenChatCompletionMessage)
	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	logger.Infof(ctx, "qwenChatCompletionMessage: %s", string(b))

	messages = append(messages, qwenChatCompletionMessage)

	response, err := sdk.QwenChatCompletion(ctx, model, messages)

	if err != nil {
		logger.Error(ctx, err)
		return "", err
	}

	content := response.Output.Text

	qwenChatCompletionMessage.Bot = content

	b, err = json.Marshal(qwenChatCompletionMessage)
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
