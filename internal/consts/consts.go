package consts

const (
	// Corp:ModelType:UserId
	MESSAGE_CONTEXT_PREFIX_KEY = "message:context:%s:%s:%d"
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
