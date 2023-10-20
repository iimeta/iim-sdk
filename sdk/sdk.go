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
)

var Robot = service.Robot()

func NewMessage() *model.Message {
	return new(model.Message)
}
