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
	IAliyun interface {
		Text(ctx context.Context, userId int, message *model.Message) (*model.Text, error)
	}
)

var (
	localAliyun IAliyun
)

func Aliyun() IAliyun {
	if localAliyun == nil {
		panic("implement not found for interface IAliyun, forgot register?")
	}
	return localAliyun
}

func RegisterAliyun(i IAliyun) {
	localAliyun = i
}
