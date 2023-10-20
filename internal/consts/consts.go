package consts

const (
	// ModelType:Stype:Sid:Robot
	MESSAGE_CONTEXT_KEY = "message:context:%s:%v:%v:%d"
)

const (
	RootStatusDeleted = -1
	RootStatusNormal  = 0
	RootStatusDisable = 1
)

const (
	MODEL_TYPE_TEXT  = "text"
	MODEL_TYPE_IMAGE = "image"
)

const (
	CORP_OPENAI     = "OpenAI"
	CORP_BAIDU      = "Baidu"
	CORP_XFYUN      = "Xfyun"
	CORP_ALIYUN     = "Aliyun"
	CORP_MIDJOURNEY = "Midjourney"
)
