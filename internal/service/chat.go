// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/sashabaranov/go-openai"
)

type (
	IChat interface {
		Chat(ctx context.Context, chat *model.Chat, retry ...int) (response openai.ChatCompletionResponse, err error)
		ChatStream(ctx context.Context, chat *model.Chat, retry ...int) (responseChan chan model.ChatCompletionStreamResponse, err error)
	}
)

var (
	localChat IChat
)

func Chat() IChat {
	if localChat == nil {
		panic("implement not found for interface IChat, forgot register?")
	}
	return localChat
}

func RegisterChat(i IChat) {
	localChat = i
}
