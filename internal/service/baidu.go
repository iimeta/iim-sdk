// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
)

type (
	IBaidu interface {
		Text(ctx context.Context, senderId, receiverId, talkType int, text, model string, mentions ...string) (string, error)
	}
)

var (
	localBaidu IBaidu
)

func Baidu() IBaidu {
	if localBaidu == nil {
		panic("implement not found for interface IBaidu, forgot register?")
	}
	return localBaidu
}

func RegisterBaidu(i IBaidu) {
	localBaidu = i
}
