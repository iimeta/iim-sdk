package model

import "github.com/sashabaranov/go-openai"

type Chat struct {
	Corp     string                         `json:"corp,omitempty"`  // 公司
	Model    string                         `json:"model,omitempty"` // 模型
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

type ChatCompletionStreamResponse struct {
	ID                string                              `json:"id"`
	Object            string                              `json:"object"`
	Created           int64                               `json:"created"`
	Model             string                              `json:"model"`
	Choices           []openai.ChatCompletionStreamChoice `json:"choices"`
	PromptAnnotations []openai.PromptAnnotation           `json:"prompt_annotations,omitempty"`
	Usage             openai.Usage                        `json:"usage"`
}
