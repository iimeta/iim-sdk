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
		Text(ctx context.Context, robotInfo *model.Robot, prompt string, isWithContext bool) (string, error)
		Image(ctx context.Context, robotInfo *model.Robot, prompt string, isSaveImage bool) (*model.Image, error)
	}
)

var (
	localRobot IRobot
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
