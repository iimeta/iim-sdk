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
