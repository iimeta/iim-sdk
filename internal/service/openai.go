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
	IOpenAI interface {
		Text(ctx context.Context, robot *model.Robot, message *model.Message, retry ...int) (*model.Text, error)
		Image(ctx context.Context, robot *model.Robot, message *model.Message) (imageInfo *model.Image, err error)
	}
)

var (
	localOpenAI IOpenAI
)

func OpenAI() IOpenAI {
	if localOpenAI == nil {
		panic("implement not found for interface IOpenAI, forgot register?")
	}
	return localOpenAI
}

func RegisterOpenAI(i IOpenAI) {
	localOpenAI = i
}
