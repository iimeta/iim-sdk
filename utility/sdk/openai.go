package sdk

import (
	"context"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/iimeta/iim-sdk/internal/config"
	"github.com/iimeta/iim-sdk/internal/errors"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/util"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var clientMap sync.Map
var openaiRoundrobin = new(util.RoundRobin)

func init() {
	ctx := gctx.New()
	for model := range config.Cfg.Sdk.OpenAI.Models {
		Init(ctx, model)
	}
}

func Init(ctx context.Context, model string) {

	baseURL := config.Cfg.Sdk.OpenAI.Models[model].BaseUrl
	proxyURL := config.Cfg.Sdk.OpenAI.Models[model].ProxyUrl
	apiKeys := config.Cfg.Sdk.OpenAI.Models[model].ApiKeys
	apiKey := openaiRoundrobin.PickKey(apiKeys)

	logger.Infof(ctx, "Init OpenAI model: %s, apiKey: %s", model, apiKey)

	config := openai.DefaultConfig(apiKey)

	if baseURL != "" {
		logger.Infof(ctx, "Init OpenAI model: %s, baseURL: %s", model, baseURL)
		config.BaseURL = baseURL
	}

	transport := &http.Transport{}

	if baseURL == "" && proxyURL != "" {
		logger.Infof(ctx, "Init OpenAI model: %s, proxyURL: %s", model, proxyURL)
		proxyUrl, err := url.Parse(proxyURL)
		if err != nil {
			panic(err)
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	config.HTTPClient = &http.Client{
		Transport: transport,
	}

	setClient(model, openai.NewClientWithConfig(config))
}

func ChatCompletion(ctx context.Context, request openai.ChatCompletionRequest, retry ...int) (openai.ChatCompletionResponse, error) {

	logger.Infof(ctx, "ChatCompletion model: %s", request.Model)

	if len(retry) > 0 {
		Init(ctx, request.Model)
	}

	now := gtime.Now().Unix()

	defer func() {
		logger.Infof(ctx, "ChatCompletion model: %s, 总耗时: %d", request.Model, gtime.Now().Unix()-now)
	}()

	response, err := getClient(request.Model).CreateChatCompletion(ctx, request)

	if err != nil {
		logger.Errorf(ctx, "ChatCompletion model: %s, error: %v", request.Model, err)
		return openai.ChatCompletionResponse{}, err
	}

	logger.Infof(ctx, "ChatCompletion model: %s, response: %s", request.Model, gjson.MustEncodeString(response))

	return response, nil
}

func ChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest, responseContent chan openai.ChatCompletionStreamResponse, retry ...int) error {

	logger.Infof(ctx, "ChatCompletionStream model: %s", request.Model)

	if len(retry) > 0 {
		Init(ctx, request.Model)
	}

	now := gtime.Now().Unix()

	defer func() {
		logger.Infof(ctx, "ChatCompletionStream model: %s, 总耗时: %d", request.Model, gtime.Now().Unix()-now)
	}()

	stream, err := getClient(request.Model).CreateChatCompletionStream(ctx, request)
	if err != nil {
		logger.Errorf(ctx, "ChatCompletionStream model: %s, error: %v", request.Model, err)
		return err
	}

	logger.Infof(ctx, "ChatCompletionStream model: %s, start", request.Model)

	for {

		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logger.Infof(ctx, "ChatCompletionStream model: %s, finished", request.Model)
			stream.Close()
			responseContent <- response
			return nil
		}

		if err != nil {
			logger.Errorf(ctx, "ChatCompletionStream model: %s, error: %v", request.Model, err)
			return err
		}

		responseContent <- response
	}
}

func GenImage(ctx context.Context, prompt string) (url string, err error) {

	logger.Infof(ctx, "GenImage prompt: %s", prompt)

	now := gtime.Now().Unix()

	defer func() {
		logger.Infof(ctx, "GenImage url: %s", url)
		logger.Infof(ctx, "GenImage 总耗时: %d", gtime.Now().Unix()-now)
	}()

	reqUrl := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respUrl, err := getClient(openai.GPT3Dot5Turbo16K).CreateImage(ctx, reqUrl)
	if err != nil {
		logger.Errorf(ctx, "GenImage creation error: %v", err)
		time.Sleep(5 * time.Second)
		Init(ctx, openai.GPT3Dot5Turbo16K)
		return GenImage(ctx, prompt)
	}

	url = respUrl.Data[0].URL

	return url, nil
}

func GenImageBase64(ctx context.Context, prompt string, retry ...int) (string, error) {

	logger.Infof(ctx, "GenImageBase64 prompt: %s", prompt)

	now := gtime.Now().Unix()
	imgBase64 := ""

	defer func() {
		logger.Infof(ctx, "GenImageBase64 len: %d", len(imgBase64))
		logger.Infof(ctx, "GenImageBase64 总耗时: %d", gtime.Now().Unix()-now)
	}()

	if len(retry) == 5 {
		return "", errors.New("响应超时, 请重试...")
	}

	reqBase64 := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := getClient(openai.GPT3Dot5Turbo16K).CreateImage(ctx, reqBase64)
	if err != nil {
		logger.Errorf(ctx, "GenImageBase64 creation error: %v", err)

		e := &openai.APIError{}
		if errors.As(err, &e) {
			switch e.HTTPStatusCode {
			case 400:
				if gstr.Contains(err.Error(), "Your request was rejected as a result of our safety system") {
					return "", err
				}
				time.Sleep(5 * time.Second)
				Init(ctx, openai.GPT3Dot5Turbo16K)
				return GenImageBase64(ctx, prompt, append(retry, 1)...)
			default:
				time.Sleep(5 * time.Second)
				Init(ctx, openai.GPT3Dot5Turbo16K)
				return GenImageBase64(ctx, prompt, append(retry, 1)...)
			}
		}
	}

	imgBase64 = respBase64.Data[0].B64JSON

	return imgBase64, nil
}

func setClient(model string, client *openai.Client) {
	clientMap.Store(model, client)
}

func getClient(model string) *openai.Client {
	value, ok := clientMap.Load(model)
	if ok {
		return value.(*openai.Client)
	}
	return nil
}
