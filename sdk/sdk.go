package sdk

import (
	"github.com/iimeta/iim-sdk/internal/consts"
	_ "github.com/iimeta/iim-sdk/internal/core"

	_ "github.com/iimeta/iim-sdk/internal/packed"

	_ "github.com/iimeta/iim-sdk/internal/logic"

	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
)

const (
	MODEL_TYPE_TEXT  = consts.MODEL_TYPE_TEXT
	MODEL_TYPE_IMAGE = consts.MODEL_TYPE_IMAGE

	CORP_OPENAI     = consts.CORP_OPENAI
	CORP_BAIDU      = consts.CORP_BAIDU
	CORP_XFYUN      = consts.CORP_XFYUN
	CORP_ALIYUN     = consts.CORP_ALIYUN
	CORP_MIDJOURNEY = consts.CORP_MIDJOURNEY
)

var Robot = service.Robot()

func NewMessage() *model.Message {
	return new(model.Message)
}
