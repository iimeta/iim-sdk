package robot

import (
	"context"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/iimeta/iim-sdk/internal/dao"
	"github.com/iimeta/iim-sdk/internal/errors"
	"github.com/iimeta/iim-sdk/internal/logic/aliyun"
	"github.com/iimeta/iim-sdk/internal/logic/baidu"
	"github.com/iimeta/iim-sdk/internal/logic/midjourney"
	"github.com/iimeta/iim-sdk/internal/logic/openai"
	"github.com/iimeta/iim-sdk/internal/logic/xfyun"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
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
		RobotName: robot.RobotName,
		Describe:  robot.Describe,
		Logo:      robot.Logo,
		IsTalk:    robot.IsTalk,
		Status:    robot.Status,
		Type:      robot.Type,
		Company:   robot.Company,
		Model:     robot.Model,
		ModelType: robot.ModelType,
		Role:      robot.Role,
		Prompt:    robot.Prompt,
		MsgType:   robot.MsgType,
		Proxy:     robot.Proxy,
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

	robots := make([]*model.Robot, len(robotList))
	for _, robot := range robotList {
		robots = append(robots, &model.Robot{
			UserId:    robot.UserId,
			RobotName: robot.RobotName,
			Describe:  robot.Describe,
			Logo:      robot.Logo,
			IsTalk:    robot.IsTalk,
			Status:    robot.Status,
			Type:      robot.Type,
			Company:   robot.Company,
			Model:     robot.Model,
			ModelType: robot.ModelType,
			Role:      robot.Role,
			Prompt:    robot.Prompt,
			MsgType:   robot.MsgType,
			Proxy:     robot.Proxy,
			CreatedAt: robot.CreatedAt,
			UpdatedAt: robot.UpdatedAt,
		})
	}

	return robots, nil
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

func RobotReply(ctx context.Context, robotInfo *model.Robot, text string, isOpenContext int, mentions ...string) {

	logger.Info(ctx, gjson.MustEncodeString(robotInfo))

	text = strings.TrimSpace(text)

	switch robotInfo.Company {
	case "OpenAI":
		switch robotInfo.ModelType {
		case "chat":
			openai.OpenAI.Chat(ctx, senderId, receiverId, talkType, text, robotInfo.Model, isOpenContext, mentions...)
		case "image":
			openai.OpenAI.Image(ctx, senderId, receiverId, talkType, text, mentions...)
		}
	case "Baidu":
		switch robotInfo.ModelType {
		case "chat":
			baidu.ErnieBot.Chat(ctx, senderId, receiverId, talkType, text, robotInfo.Model, mentions...)
		}
	case "Xfyun":
		switch robotInfo.ModelType {
		case "chat":
			xfyun.Spark.Chat(ctx, senderId, receiverId, talkType, text, robotInfo.Model, mentions...)
		}
	case "Aliyun":
		switch robotInfo.ModelType {
		case "chat":
			aliyun.Aliyun.Chat(ctx, senderId, receiverId, talkType, text, robotInfo.Model, mentions...)
		}
	case "Midjourney":
		switch robotInfo.ModelType {
		case "image":
			midjourney.Midjourney.Image(ctx, senderId, receiverId, talkType, text, robotInfo.Proxy)
		}
	}
}

func (s *sRobot) Text(ctx context.Context, robotInfo *model.Robot, prompt string, isWithContext bool) (string, error) {
	return "", nil
}

func (s *sRobot) Image(ctx context.Context, robotInfo *model.Robot, prompt string, isSaveImage bool) (*model.Image, error) {
	return nil, nil
}
