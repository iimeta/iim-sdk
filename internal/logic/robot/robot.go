package robot

import (
	"context"
	"github.com/iimeta/iim-sdk/internal/consts"
	"github.com/iimeta/iim-sdk/internal/dao"
	"github.com/iimeta/iim-sdk/internal/errors"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type sRobot struct{}

func init() {
	service.RegisterRobot(New())
}

func New() service.IRobot {
	return &sRobot{}
}

func (s *sRobot) GetRobotByUserId(ctx context.Context, userId int) (*model.Robot, error) {

	robot, err := dao.Robot.GetRobotByUserId(ctx, userId)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Error(ctx, err)
		return nil, err
	}

	if robot == nil {
		return nil, nil
	}

	return &model.Robot{
		UserId:    robot.UserId,
		IsTalk:    robot.IsTalk,
		Status:    robot.Status,
		Type:      robot.Type,
		Corp:      robot.Corp,
		Model:     robot.Model,
		ModelType: robot.ModelType,
		Role:      robot.Role,
		Prompt:    robot.Prompt,
		Proxy:     robot.Proxy,
		Key:       robot.Key,
		CreatedAt: robot.CreatedAt,
		UpdatedAt: robot.UpdatedAt,
	}, nil
}

func (s *sRobot) GetRobotsByUserIds(ctx context.Context, userId ...int) ([]*model.Robot, error) {

	robotList, err := dao.Robot.GetRobotList(ctx, userId...)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Error(ctx, err)
		return nil, err
	}

	if robotList == nil || len(robotList) == 0 {
		return nil, nil
	}

	robots := make([]*model.Robot, 0)
	for _, robot := range robotList {
		robots = append(robots, &model.Robot{
			UserId:    robot.UserId,
			IsTalk:    robot.IsTalk,
			Status:    robot.Status,
			Type:      robot.Type,
			Corp:      robot.Corp,
			Model:     robot.Model,
			ModelType: robot.ModelType,
			Role:      robot.Role,
			Prompt:    robot.Prompt,
			Proxy:     robot.Proxy,
			Key:       robot.Key,
			CreatedAt: robot.CreatedAt,
			UpdatedAt: robot.UpdatedAt,
		})
	}

	return robots, nil
}

func (s *sRobot) ClearMessageContext(ctx context.Context, robot *model.Robot, message *model.Message) (int64, error) {
	return service.Common().ClearMessageContext(ctx, robot, message)
}

func (s *sRobot) IsNeedRobotReply(ctx context.Context, userId ...int) ([]*model.Robot, bool) {

	// todo 需要改成查缓存
	robots, err := s.GetRobotsByUserIds(ctx, userId...)
	if err != nil {
		logger.Error(ctx, err)
		return nil, false
	}

	if robots == nil || len(robots) == 0 {
		return nil, false
	}

	return robots, true
}

func (s *sRobot) Text(ctx context.Context, robot *model.Robot, message *model.Message) (*model.Text, error) {

	switch robot.Corp {
	case consts.CORP_OPENAI:
		return service.OpenAI().Text(ctx, robot, message)
	case consts.CORP_BAIDU:
		return service.Baidu().Text(ctx, robot, message)
	case consts.CORP_XFYUN:
		return service.Xfyun().Text(ctx, robot, message)
	case consts.CORP_ALIYUN:
		return service.Aliyun().Text(ctx, robot, message)
	default:
		return nil, errors.New("Unknown Corp: " + robot.Corp)
	}
}

func (s *sRobot) Image(ctx context.Context, robot *model.Robot, message *model.Message) (*model.Image, error) {

	switch robot.Corp {
	case consts.CORP_OPENAI:
		return service.OpenAI().Image(ctx, robot, message)
	case consts.CORP_MIDJOURNEY:
		return service.Midjourney().Image(ctx, robot, message)
	default:
		return nil, errors.New("Unknown Corp: " + robot.Corp)
	}
}
