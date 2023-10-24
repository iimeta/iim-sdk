package sdk

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/iimeta/iim-sdk/internal/config"
	"github.com/iimeta/iim-sdk/internal/errors"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/util"
	"time"
)

var ernieBotRoundRobin = new(util.RoundRobin)

const ACCESS_TOKEN_KEY = "sdk:baidu:access_token:%s"

type ErnieBotMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ErnieBotReq struct {
	Messages []ErnieBotMessage `json:"messages"`
}
type ErnieBotRes struct {
	Id               string `json:"id"`
	Object           string `json:"object"`
	Created          int    `json:"created"`
	Result           string `json:"result"`
	IsTruncated      bool   `json:"is_truncated"`
	NeedClearHistory bool   `json:"need_clear_history"`
	Usage            struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

func ErnieBot(ctx context.Context, model string, messages []ErnieBotMessage, retry ...int) (res *ErnieBotRes, err error) {

	if len(retry) > 5 {
		return nil, errors.New("响应超时, 请重试...")
	}

	logger.Infof(ctx, "ErnieBot model: %s", model)

	now := gtime.Now().Unix()

	apps := config.Cfg.Sdk.Baidu.Models[model].Apps
	app := apps[ernieBotRoundRobin.Index(len(apps))]

	defer func() {
		logger.Infof(ctx, "ErnieBot model: %s, appid: %s, 总耗时: %d", model, app.Id, gtime.Now().Unix()-now)
	}()

	req := ErnieBotReq{
		Messages: messages,
	}

	ernieBotRes := new(ErnieBotRes)
	err = util.HttpPostJson(ctx, fmt.Sprintf("%s?access_token=%s", config.Cfg.Sdk.Baidu.Models[model].BaseUrl+config.Cfg.Sdk.Baidu.Models[model].Path, getAccessToken(ctx, app)), nil, req, &ernieBotRes, config.Cfg.Sdk.Baidu.Models[model].ProxyUrl)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(3 * time.Second)
		return ErnieBot(ctx, model, messages, append(retry, 1)...)
	}

	if ernieBotRes.ErrorCode != 0 {
		logger.Error(ctx, gjson.MustEncodeString(ernieBotRes))
		if len(retry) < 5 {
			time.Sleep(3 * time.Second)
			return ErnieBot(ctx, model, messages, append(retry, 1)...)

		}
		return nil, gerror.Newf("ErnieBot ErrorCode: %d, ErrorMsg: %s, 发生错误, 请联系作者处理...", ernieBotRes.ErrorCode, ernieBotRes.ErrorMsg)
	}

	return ernieBotRes, nil
}

type GetAccessTokenRes struct {
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	SessionKey       string `json:"session_key"`
	AccessToken      string `json:"access_token"`
	Scope            string `json:"scope"`
	SessionSecret    string `json:"session_secret"`
	ErrorDescription string `json:"error_description"`
	Error            string `json:"error"`
}

func getAccessToken(ctx context.Context, app *config.App) string {

	reply, err := g.Redis().Get(ctx, fmt.Sprintf(ACCESS_TOKEN_KEY, app.Id))
	if err == nil && reply.String() != "" {
		return reply.String()
	}

	data := g.Map{
		"grant_type":    "client_credentials",
		"client_id":     app.Key,
		"client_secret": app.Secret,
	}

	getAccessTokenRes := new(GetAccessTokenRes)
	err = util.HttpPost(ctx, config.Cfg.Sdk.Baidu.AccessToken.BaseUrl+config.Cfg.Sdk.Baidu.AccessToken.Path, nil, data, &getAccessTokenRes, config.Cfg.Sdk.Baidu.AccessToken.ProxyUrl)
	if err != nil {
		logger.Error(ctx, err)
		return ""
	}

	if getAccessTokenRes.Error != "" {
		logger.Error(ctx, getAccessTokenRes.Error)
		return ""
	}

	_ = g.Redis().SetEX(ctx, fmt.Sprintf(ACCESS_TOKEN_KEY, app.Id), getAccessTokenRes.AccessToken, getAccessTokenRes.ExpiresIn)

	return getAccessTokenRes.AccessToken
}
