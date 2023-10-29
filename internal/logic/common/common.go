package common

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/iimeta/iim-sdk/internal/consts"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/redis"
	"github.com/iimeta/iim-sdk/utility/sdk"
	"github.com/iimeta/iim-sdk/utility/util"
	"github.com/sashabaranov/go-openai"
)

type sCommon struct{}

func init() {
	service.RegisterCommon(New())
}

func New() service.ICommon {
	return &sCommon{}
}

func (s *sCommon) GetMessageContext(ctx context.Context, robot *model.Robot, message *model.Message) []string {

	reply, err := g.Redis().LRange(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_KEY, robot.ModelType, message.Stype, message.Sid, robot.UserId), 0, -1)
	if err != nil {
		logger.Error(ctx, err)
		return nil
	}

	return reply.Strings()
}

func (s *sCommon) SaveMessageContext(ctx context.Context, robot *model.Robot, message *model.Message, value any) error {

	b, err := json.Marshal(value)
	if err != nil {
		logger.Error(ctx, err)
		return err
	}

	_, err = g.Redis().RPush(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_KEY, robot.ModelType, message.Stype, message.Sid, robot.UserId), b)
	if err != nil {
		logger.Error(ctx, err)
		return err
	}

	return nil
}

func (s *sCommon) ClearMessageContext(ctx context.Context, robot *model.Robot, message *model.Message) (int64, error) {
	return redis.Del(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_KEY, robot.ModelType, message.Stype, message.Sid, robot.UserId))
}

func (s *sCommon) TrimMessageContext(ctx context.Context, robot *model.Robot, message *model.Message, start, stop int64) error {
	return redis.LTrim(ctx, fmt.Sprintf(consts.MESSAGE_CONTEXT_KEY, robot.ModelType, message.Stype, message.Sid, robot.UserId), start, stop)
}

func (s *sCommon) Translate(ctx context.Context, text string, retry ...int) (res string) {

	var err error
	var response openai.ChatCompletionResponse

	defer func() {
		if err != nil && len(retry) < 5 {
			res = s.Translate(ctx, text, append(retry, 1)...)
		}
	}()

	if util.HasChinese(text) {

		response, err = sdk.ChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo16K,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "把中文翻译成英文",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			}}, retry...)

		if err != nil {
			logger.Error(ctx, err)
			return text
		}

		return response.Choices[0].Message.Content
	}

	return text
}
