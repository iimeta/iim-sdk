package sdk

import (
	_ "github.com/iimeta/iim-sdk/internal/core"
	"github.com/iimeta/iim-sdk/internal/service"

	_ "github.com/iimeta/iim-sdk/internal/packed"

	_ "github.com/iimeta/iim-sdk/internal/logic"
	"github.com/iimeta/iim-sdk/internal/model"
)

var Robot = service.Robot()

func NewMessage() *model.Message {
	return new(model.Message)
}

type Message struct {
	*model.Message
}
