package chat

import (
	"context"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/iimeta/iim-sdk/internal/consts"
	"github.com/iimeta/iim-sdk/internal/errors"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/sdk"
	"github.com/sashabaranov/go-openai"
	"time"
)

type sChat struct{}

func init() {
	service.RegisterChat(New())
}

func New() service.IChat {
	return &sChat{}
}

func (s *sChat) Chat(ctx context.Context, chat *model.Chat, retry ...int) (response openai.ChatCompletionResponse, err error) {

	// todo
	if len(retry) == 5 {
		logger.Infof(ctx, "Chat model: %s, retry: %d", chat.Model, len(retry))
		chat.Model = openai.GPT3Dot5Turbo16K
	}

	defer func() {
		if err != nil {

			e := &openai.APIError{}
			if errors.As(err, &e) {

				if len(retry) == 10 {
					response = openai.ChatCompletionResponse{
						ID:      "error",
						Object:  "chat.completion",
						Created: time.Now().Unix(),
						Model:   chat.Model,
						Choices: []openai.ChatCompletionChoice{{
							FinishReason: "stop",
							Message: openai.ChatCompletionMessage{
								Role:    openai.ChatMessageRoleAssistant,
								Content: err.Error(),
							},
						}},
					}
					return
				}

				switch e.HTTPStatusCode {
				case 400:
					if gstr.Contains(err.Error(), "Please reduce the length of the messages") {
						response = openai.ChatCompletionResponse{
							ID:      "error",
							Object:  "chat.completion",
							Created: time.Now().Unix(),
							Model:   chat.Model,
							Choices: []openai.ChatCompletionChoice{{
								FinishReason: "stop",
								Message: openai.ChatCompletionMessage{
									Role:    openai.ChatMessageRoleAssistant,
									Content: err.Error(),
								},
							}},
						}
						return
					}
					response, err = s.Chat(ctx, chat, append(retry, 1)...)
				case 429:
					response, err = s.Chat(ctx, chat, append(retry, 1)...)
				default:
					response, err = s.Chat(ctx, chat, append(retry, 1)...)
				}
			}
		}
	}()

	switch chat.Corp {
	case consts.CORP_OPENAI:
		response, err = sdk.ChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    chat.Model,
			Messages: chat.Messages,
		}, retry...)
	case consts.CORP_BAIDU:
	case consts.CORP_XFYUN:
	case consts.CORP_ALIYUN:
	default:
		return response, errors.New("Unknown Corp: " + chat.Corp)
	}

	if err != nil {
		logger.Errorf(ctx, "Chat model: %s, error: %v", chat.Model, err)
		return response, err
	}

	return response, nil
}

func (s *sChat) ChatStream(ctx context.Context, chat *model.Chat, retry ...int) (responseChan chan model.ChatCompletionStreamResponse, err error) {

	// todo
	if len(retry) == 5 {
		logger.Infof(ctx, "ChatStream model: %s, retry: %d", chat.Model, len(retry))
		chat.Model = openai.GPT3Dot5Turbo16K
	}

	responseChan = make(chan model.ChatCompletionStreamResponse)

	defer func() {
		if err != nil {

			e := &openai.APIError{}
			if errors.As(err, &e) {

				if len(retry) == 10 {
					responseChan <- model.ChatCompletionStreamResponse{
						ID:      "error",
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   chat.Model,
						Choices: []openai.ChatCompletionStreamChoice{{
							FinishReason: "stop",
							Delta: openai.ChatCompletionStreamChoiceDelta{
								Content: err.Error(),
							},
						}},
					}
					return
				}

				switch e.HTTPStatusCode {
				case 400:
					if gstr.Contains(err.Error(), "Please reduce the length of the messages") {
						responseChan <- model.ChatCompletionStreamResponse{
							ID:      "error",
							Object:  "chat.completion.chunk",
							Created: time.Now().Unix(),
							Model:   chat.Model,
							Choices: []openai.ChatCompletionStreamChoice{{
								FinishReason: "stop",
								Delta: openai.ChatCompletionStreamChoiceDelta{
									Content: err.Error(),
								},
							}},
						}
						return
					}
					responseChan, err = s.ChatStream(ctx, chat, append(retry, 1)...)
				case 429:
					responseChan, err = s.ChatStream(ctx, chat, append(retry, 1)...)
				default:
					responseChan, err = s.ChatStream(ctx, chat, append(retry, 1)...)
				}
			}
		}
	}()

	var streamResponse chan openai.ChatCompletionStreamResponse

	switch chat.Corp {
	case consts.CORP_OPENAI:
		streamResponse, err = sdk.ChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    chat.Model,
			Messages: chat.Messages,
			Stream:   true,
		}, retry...)
	case consts.CORP_BAIDU:
	case consts.CORP_XFYUN:
	case consts.CORP_ALIYUN:
	}

	if err != nil {
		logger.Error(ctx, err)
		return responseChan, err
	}

	if err = grpool.AddWithRecover(ctx, func(ctx context.Context) {

		defer close(streamResponse)
		completionTokens := 0

		promptTokens, err := sdk.NumTokensFromMessages(chat.Messages, chat.Model)
		if err != nil {
			logger.Error(ctx, err)
			return
		}

		for {
			select {
			case streamResponse := <-streamResponse:

				response := model.ChatCompletionStreamResponse{
					ID:                streamResponse.ID,
					Object:            streamResponse.Object,
					Created:           streamResponse.Created,
					Model:             streamResponse.Model,
					Choices:           streamResponse.Choices,
					PromptAnnotations: streamResponse.PromptAnnotations,
				}

				if response.Usage.CompletionTokens, err = sdk.NumTokensFromString(response.Choices[0].Delta.Content, chat.Model); err != nil {
					logger.Error(ctx, err)
					return
				}

				completionTokens += response.Usage.CompletionTokens
				response.Usage.PromptTokens = promptTokens
				response.Usage.CompletionTokens = completionTokens
				response.Usage.TotalTokens = response.Usage.PromptTokens + response.Usage.CompletionTokens

				responseChan <- response

			default:
				if err != nil {
					logger.Errorf(ctx, "ChatStream model: %s, error: %v", chat.Model, err)
					return
				}
			}
		}
	}, nil); err != nil {
		logger.Error(ctx, err)
		return responseChan, err
	}

	return responseChan, nil
}
