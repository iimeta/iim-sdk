package sdk

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/iimeta/iim-sdk/internal/config"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/util"
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

func ErnieBot(ctx context.Context, model string, messages []ErnieBotMessage, retry ...int) (*ErnieBotRes, error) {

	logger.Infof(ctx, "ErnieBot model: %s", model)

	now := gtime.Now().Unix()

	defer func() {
		logger.Infof(ctx, "ErnieBot model: %s, 总耗时: %d", model, gtime.Now().Unix()-now)
	}()

	req := ErnieBotReq{
		Messages: messages,
	}

	ernieBotRes := new(ErnieBotRes)
	err := util.HttpPostJson(ctx, fmt.Sprintf("%s?access_token=%s", config.Cfg.Sdk.Baidu.Models[model].BaseUrl+config.Cfg.Sdk.Baidu.Models[model].Path, GetAccessToken(ctx, model)), nil, req, &ernieBotRes, config.Cfg.Sdk.Baidu.Models[model].ProxyUrl)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	logger.Infof(ctx, "ErnieBot model: %s, ernieBotRes: %s", model, gjson.MustEncodeString(ernieBotRes))

	if ernieBotRes.ErrorCode != 0 {
		return nil, gerror.Newf("ErnieBot model: %s, ErrorCode: %d, ErrorMsg: %s", model, ernieBotRes.ErrorCode, ernieBotRes.ErrorMsg)
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

func GetAccessToken(ctx context.Context, model string) string {

	apps := config.Cfg.Sdk.Baidu.Models[model].Apps
	app := apps[ernieBotRoundRobin.Index(len(apps))]

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
