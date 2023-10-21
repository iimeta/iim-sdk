// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/iimeta/iim-sdk/internal/model"
)

type (
	IRobot interface {
		GetRobotByUserId(ctx context.Context, userId int) (*model.Robot, error)
		GetRobotsByUserIds(ctx context.Context, userId ...int) ([]*model.Robot, error)
		IsNeedRobotReply(ctx context.Context, userId ...int) ([]*model.Robot, bool)
		Text(ctx context.Context, robot *model.Robot, message *model.Message) (*model.Text, error)
		Image(ctx context.Context, robot *model.Robot, message *model.Message) (*model.Image, error)
	}
	ICommon interface {
		ClearMessageContext(ctx context.Context, robot *model.Robot, message *model.Message) (int64, error)
	}
)

var (
	localRobot  IRobot
	localCommon ICommon
)

func Robot() IRobot {
	if localRobot == nil {
		panic("implement not found for interface IRobot, forgot register?")
	}
	return localRobot
}

func RegisterRobot(i IRobot) {
	localRobot = i
}

func Common() ICommon {
	if localCommon == nil {
		panic("implement not found for interface ICommon, forgot register?")
	}
	return localCommon
}

func RegisterCommon(i ICommon) {
	localCommon = i
}
