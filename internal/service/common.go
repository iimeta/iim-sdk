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
	ICommon interface {
		GetMessageContext(ctx context.Context, robot *model.Robot, message *model.Message) []string
		SaveMessageContext(ctx context.Context, robot *model.Robot, message *model.Message, value any) error
		ClearMessageContext(ctx context.Context, robot *model.Robot, message *model.Message) (int64, error)
		TrimMessageContext(ctx context.Context, robot *model.Robot, message *model.Message, start, stop int64) error
	}
)

var (
	localCommon ICommon
)

func Common() ICommon {
	if localCommon == nil {
		panic("implement not found for interface ICommon, forgot register?")
	}
	return localCommon
}

func RegisterCommon(i ICommon) {
	localCommon = i
}
