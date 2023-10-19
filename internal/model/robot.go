package model

type Robot struct {
	UserId    int    `bson:"user_id,omitempty"`    // 绑定的用户ID
	IsTalk    int    `bson:"is_talk,omitempty"`    // 是否可发送消息[0:否;1:是;]
	Type      int    `bson:"type,omitempty"`       // 机器人类型
	Status    int    `bson:"status,omitempty"`     // 状态[-1:已删除;0:正常;1:已禁用;]
	Corp      string `bson:"corp,omitempty"`       // 公司
	Model     string `bson:"model,omitempty"`      // 模型
	ModelType string `bson:"model_type,omitempty"` // 模型类型, 文生文: text, 画图: image
	Role      string `bson:"role,omitempty"`       // 角色
	Prompt    string `bson:"prompt,omitempty"`     // 提示
	Proxy     string `bson:"proxy,omitempty"`      // 代理
	CreatedAt int64  `bson:"created_at,omitempty"` // 创建时间
	UpdatedAt int64  `bson:"updated_at,omitempty"` // 更新时间
}

type Text struct {
	Content string `json:"content"`
	Usage   *Usage `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Image struct {
	Url      string `json:"url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Size     int    `json:"size"`
	Md5Sum   string `json:"md5sum"`
	FilePath string `json:"file_path"`
	TaskId   string `json:"task_id"`
}

type Message struct {
	Corp          string `json:"corp"`            // 公司
	Model         string `json:"model"`           // 模型
	ModelType     string `json:"model_type"`      // 模型类型, 文生文: text, 画图: image
	Prompt        string `json:"prompt"`          // 提示
	Key           string `json:"key"`             // 密钥
	Proxy         string `json:"proxy"`           // 代理
	IsWithContext bool   `json:"is_with_context"` // 是否带上下文
	IsSave        bool   `json:"is_save"`         // 是否保存
}
