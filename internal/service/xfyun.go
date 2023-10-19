// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
)

type (
	IXfyun interface {
		Text(ctx context.Context, senderId, receiverId, talkType int, text, model string, mentions ...string) (string, error)
	}
)

var (
	localXfyun IXfyun
)

func Xfyun() IXfyun {
	if localXfyun == nil {
		panic("implement not found for interface IXfyun, forgot register?")
	}
	return localXfyun
}

func RegisterXfyun(i IXfyun) {
	localXfyun = i
}
