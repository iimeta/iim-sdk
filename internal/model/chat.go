package model

import "github.com/sashabaranov/go-openai"

type Chat struct {
	UserId    int                            `json:"user_id,omitempty"`    // 绑定的用户ID
	IsTalk    int                            `json:"is_talk,omitempty"`    // 是否可发送消息[0:否;1:是;]
	Type      int                            `json:"type,omitempty"`       // 机器人类型
	Status    int                            `json:"status,omitempty"`     // 状态[-1:已删除;0:正常;1:已禁用;]
	Corp      string                         `json:"corp,omitempty"`       // 公司
	Model     string                         `json:"model,omitempty"`      // 模型
	ModelType string                         `json:"model_type,omitempty"` // 模型类型, 文生文: text, 画图: image
	Role      string                         `json:"role,omitempty"`       // 角色
	Prompt    string                         `json:"prompt,omitempty"`     // 提示
	Proxy     string                         `json:"proxy,omitempty"`      // 代理
	Key       string                         `json:"key,omitempty"`        // 密钥
	CreatedAt int64                          `json:"created_at,omitempty"` // 创建时间
	UpdatedAt int64                          `json:"updated_at,omitempty"` // 更新时间
	Messages  []openai.ChatCompletionMessage `json:"messages"`
	Stream    bool                           `json:"stream,omitempty"`
}

type ChatMessage struct {
	Role         string               `json:"role"`
	Content      string               `json:"content"`
	FunctionCall *openai.FunctionCall `json:"function_call,omitempty"`
}

type ChatCompletionRequest struct {
	Model            string                         `json:"model"`
	Messages         []openai.ChatCompletionMessage `json:"messages"`
	MaxTokens        int                            `json:"max_tokens,omitempty"`
	Temperature      float32                        `json:"temperature,omitempty"`
	TopP             float32                        `json:"top_p,omitempty"`
	N                int                            `json:"n,omitempty"`
	Stream           bool                           `json:"stream,omitempty"`
	Stop             []string                       `json:"stop,omitempty"`
	PresencePenalty  float32                        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32                        `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int                 `json:"logit_bias,omitempty"`
	User             string                         `json:"user,omitempty"`
	Functions        []openai.FunctionDefinition    `json:"functions,omitempty"`
	FunctionCall     any                            `json:"function_call,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string                        `json:"id"`
	Object  string                        `json:"object"`
	Created int64                         `json:"created"`
	Model   string                        `json:"model"`
	Choices []openai.ChatCompletionChoice `json:"choices"`
	Usage   Usage                         `json:"usage"`
}
