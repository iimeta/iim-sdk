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
	IMidjourney interface {
		Image(ctx context.Context, senderId, receiverId, talkType int, text string, proxy string) (*model.Image, string, error)
	}
)

var (
	localMidjourney IMidjourney
)

func Midjourney() IMidjourney {
	if localMidjourney == nil {
		panic("implement not found for interface IMidjourney, forgot register?")
	}
	return localMidjourney
}

func RegisterMidjourney(i IMidjourney) {
	localMidjourney = i
}
