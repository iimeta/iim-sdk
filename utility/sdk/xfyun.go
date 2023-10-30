package sdk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/encoding/gurl"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gorilla/websocket"
	"github.com/iimeta/iim-sdk/internal/config"
	"github.com/iimeta/iim-sdk/internal/errors"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/util"
	"net/url"
	"time"
)

var sparkRoundRobin = new(util.RoundRobin)

type Header struct {
	// req
	AppId string `json:"app_id"`
	Uid   string `json:"uid"`
	// res
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Sid     string `json:"sid,omitempty"`
	Status  int    `json:"status,omitempty"`
}

type Parameter struct {
	// req
	Chat *Chat `json:"chat"`
}

type Chat struct {
	// req
	Domain          string `json:"domain"`
	RandomThreshold int    `json:"random_threshold"`
	MaxTokens       int    `json:"max_tokens"`
}

type Payload struct {
	// req
	Message *Message `json:"message"`
	// res
	Choices *Choices `json:"choices,omitempty"`
	Usage   *Usage   `json:"usage,omitempty"`
}

type Message struct {
	// req
	Text []Text `json:"text"`
}

type Text struct {
	// req res
	Role    string `json:"role"`
	Content string `json:"content"`

	// Choices
	Index int `json:"index,omitempty"`

	// Usage
	QuestionTokens   int `json:"question_tokens,omitempty"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type Choices struct {
	// res
	Status int    `json:"status,omitempty"`
	Seq    int    `json:"seq,omitempty"`
	Text   []Text `json:"text,omitempty"`
}

type Usage struct {
	// res
	Text *Text `json:"text,omitempty"`
}

type SparkReq struct {
	Header    Header    `json:"header"`
	Parameter Parameter `json:"parameter"`
	Payload   Payload   `json:"payload"`
}
type SparkRes struct {
	Content string  `json:"content"`
	Header  Header  `json:"header"`
	Payload Payload `json:"payload"`
}

func Spark(ctx context.Context, model, uid string, text []Text, retry ...int) (res *SparkRes, err error) {

	if len(retry) > 5 {
		return nil, errors.New("响应超时, 请重试...")
	}

	logger.Infof(ctx, "Spark model: %s", model)

	now := gtime.Now().UnixMilli()

	apps := config.Cfg.Sdk.Xfyun.Models[model].Apps
	app := apps[sparkRoundRobin.Index(len(apps))]

	defer func() {
		logger.Infof(ctx, "Spark model: %s, appid: %s, totalTime: %d ms", model, app.Id, gtime.Now().UnixMilli()-now)
	}()

	sparkReq := SparkReq{
		Header: Header{
			AppId: app.Id,
			Uid:   uid,
		},
		Parameter: Parameter{
			Chat: &Chat{
				Domain:          config.Cfg.Sdk.Xfyun.Models[model].Domain,
				RandomThreshold: 0,
				MaxTokens:       config.Cfg.Sdk.Xfyun.Models[model].MaxTokens,
			},
		},
		Payload: Payload{
			Message: &Message{
				Text: text,
			},
		},
	}

	data, err := gjson.Marshal(sparkReq)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	url := getAuthorizationUrl(ctx, config.Cfg.Sdk.Xfyun.Models[model], app)

	logger.Infof(ctx, "Spark model: %s, appid: %s, getAuthorizationUrl: %s", model, app.Id, url)

	result := make(chan []byte)
	var conn *websocket.Conn

	_ = grpool.AddWithRecover(ctx, func(ctx context.Context) {
		conn, err = util.WebSocketClient(ctx, url, websocket.TextMessage, data, result)
		if err != nil {
			logger.Error(ctx, err)
		}
	}, nil)

	defer func() {
		err = conn.Close()
		if err != nil {
			logger.Error(ctx, err)
		}
	}()

	responseContent := ""
	for {
		select {
		case message := <-result:

			sparkRes := new(SparkRes)
			err := gjson.Unmarshal(message, &sparkRes)
			if err != nil {
				logger.Error(ctx, err)
				time.Sleep(3 * time.Second)
				return Spark(ctx, model, uid, text, append(retry, 1)...)
			}

			if sparkRes.Header.Code != 0 {
				if len(retry) < 5 {
					time.Sleep(3 * time.Second)
					return Spark(ctx, model, uid, text, append(retry, 1)...)
				}
				return nil, errors.New(gjson.MustEncodeString(sparkRes) + ", 发生错误, 请联系作者处理...")
			}

			responseContent += sparkRes.Payload.Choices.Text[0].Content

			if sparkRes.Header.Status == 2 {
				sparkRes.Content = responseContent
				return sparkRes, nil
			}
		}
	}
}

func SparkStream(ctx context.Context, model, uid string, text []Text, responseContent chan Payload, retry ...int) {

	if len(retry) > 5 {
		responseContent <- Payload{
			// todo
		}
		return
	}

	logger.Infof(ctx, "SparkStream model: %s", model)

	now := gtime.Now().UnixMilli()

	apps := config.Cfg.Sdk.Xfyun.Models[model].Apps
	app := apps[sparkRoundRobin.Index(len(apps))]

	defer func() {
		logger.Infof(ctx, "SparkStream model: %s, appid: %s, totalTime: %d ms", model, app.Id, gtime.Now().UnixMilli()-now)
	}()

	sparkReq := SparkReq{
		Header: Header{
			AppId: app.Id,
			Uid:   uid,
		},
		Parameter: Parameter{
			Chat: &Chat{
				Domain:          config.Cfg.Sdk.Xfyun.Models[model].Domain,
				RandomThreshold: 0,
				MaxTokens:       config.Cfg.Sdk.Xfyun.Models[model].MaxTokens,
			},
		},
		Payload: Payload{
			Message: &Message{
				Text: text,
			},
		},
	}

	data, err := gjson.Marshal(sparkReq)
	if err != nil {
		logger.Error(ctx, err)
		return
	}

	url := getAuthorizationUrl(ctx, config.Cfg.Sdk.Xfyun.Models[model], app)

	logger.Infof(ctx, "SparkStream model: %s, appid: %s, getAuthorizationUrl: %s", model, app.Id, url)

	result := make(chan []byte)
	var conn *websocket.Conn

	_ = grpool.AddWithRecover(ctx, func(ctx context.Context) {
		conn, err = util.WebSocketClient(ctx, url, websocket.TextMessage, data, result)
		if err != nil {
			logger.Error(ctx, err)
		}
	}, nil)

	defer func() {
		err = conn.Close()
		if err != nil {
			logger.Error(ctx, err)
		}
	}()

	for {
		select {
		case message := <-result:

			sparkRes := new(SparkRes)
			err := gjson.Unmarshal(message, &sparkRes)
			if err != nil {
				logger.Error(ctx, err)
				return
			}

			responseContent <- sparkRes.Payload

			if sparkRes.Header.Status == 2 {
				return
			}
		}
	}
}

func getAuthorizationUrl(ctx context.Context, model *config.Model, app *config.App) string {

	parse, err := url.Parse(config.Cfg.Sdk.Xfyun.OriginalUrl + model.Path)
	if err != nil {
		logger.Error(ctx, err)
		return ""
	}

	now := gtime.Now()
	loc, _ := time.LoadLocation("GMT")
	zone, _ := now.ToZone(loc.String())
	date := zone.Layout("Mon, 02 Jan 2006 15:04:05 GMT")

	tmp := "host: " + parse.Host + "\n"
	tmp += "date: " + date + "\n"
	tmp += "GET " + parse.Path + " HTTP/1.1"

	hash := hmac.New(sha256.New, []byte(app.Secret))

	_, err = hash.Write([]byte(tmp))
	if err != nil {
		logger.Error(ctx, err)
		return ""
	}

	signature := gbase64.EncodeToString(hash.Sum(nil))

	authorizationOrigin := gbase64.EncodeToString([]byte(fmt.Sprintf("api_key=\"%s\",algorithm=\"%s\",headers=\"%s\",signature=\"%s\"", app.Key, "hmac-sha256", "host date request-line", signature)))

	wsURL := gstr.Replace(gstr.Replace(model.BaseUrl+model.Path, "https://", "wss://"), "http://", "ws://")

	return fmt.Sprintf("%s?authorization=%s&date=%s&host=%s", wsURL, authorizationOrigin, gurl.RawEncode(date), parse.Host)
}
