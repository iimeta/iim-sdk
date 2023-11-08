package sdk

import (
	"context"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/grpool"
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

	now := gtime.Now().UnixMilli()

	defer func() {
		logger.Infof(ctx, "ChatCompletion model: %s, totalTime: %d ms", request.Model, gtime.Now().UnixMilli()-now)
	}()

	response, err := getClient(request.Model).CreateChatCompletion(ctx, request)

	if err != nil {
		logger.Errorf(ctx, "ChatCompletion model: %s, error: %v", request.Model, err)
		return openai.ChatCompletionResponse{}, err
	}

	logger.Infof(ctx, "ChatCompletion model: %s, response: %s", request.Model, gjson.MustEncodeString(response))

	return response, nil
}

func ChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest, retry ...int) (responseChan chan openai.ChatCompletionStreamResponse, err error) {

	logger.Infof(ctx, "ChatCompletionStream model: %s", request.Model)

	if len(retry) > 0 {
		Init(ctx, request.Model)
	}

	now := gtime.Now().UnixMilli()

	defer func() {
		if err != nil {
			logger.Infof(ctx, "ChatCompletionStream model: %s, totalTime: %d ms", request.Model, gtime.Now().UnixMilli()-now)
		}
	}()

	stream, err := getClient(request.Model).CreateChatCompletionStream(ctx, request)
	if err != nil {
		logger.Errorf(ctx, "ChatCompletionStream model: %s, error: %v", request.Model, err)
		return responseChan, err
	}

	logger.Infof(ctx, "ChatCompletionStream model: %s, start", request.Model)

	duration := gtime.Now().UnixMilli()

	responseChan = make(chan openai.ChatCompletionStreamResponse)

	if err = grpool.AddWithRecover(ctx, func(ctx context.Context) {

		defer func() {
			end := gtime.Now().UnixMilli()
			logger.Infof(ctx, "ChatCompletionStream model: %s, connTime: %d ms, duration: %d ms, totalTime: %d ms", request.Model, duration-now, end-duration, end-now)
		}()

		for {

			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				logger.Infof(ctx, "ChatCompletionStream model: %s, finished", request.Model)
				stream.Close()
				responseChan <- response
				return
			}

			if err != nil {
				logger.Errorf(ctx, "ChatCompletionStream model: %s, error: %v", request.Model, err)
				close(responseChan)
				return
			}

			responseChan <- response
		}
	}, nil); err != nil {
		logger.Error(ctx, err)
		return responseChan, err
	}

	return responseChan, nil
}

func GenImage(ctx context.Context, model, prompt string) (url string, err error) {

	logger.Infof(ctx, "GenImage model: %s, prompt: %s", model, prompt)

	now := gtime.Now().UnixMilli()

	defer func() {
		logger.Infof(ctx, "GenImage model: %s, url: %s", model, url)
		logger.Infof(ctx, "GenImage model: %s, totalTime: %d ms", model, gtime.Now().UnixMilli()-now)
	}()

	reqUrl := openai.ImageRequest{
		Model:          model,
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respUrl, err := getClient(model).CreateImage(ctx, reqUrl)
	if err != nil {
		logger.Errorf(ctx, "GenImage creation error: %v", err)
		time.Sleep(5 * time.Second)
		Init(ctx, model)
		return GenImage(ctx, model, prompt)
	}

	url = respUrl.Data[0].URL

	return url, nil
}

func GenImageBase64(ctx context.Context, model, prompt string, retry ...int) (string, error) {

	logger.Infof(ctx, "GenImageBase64 model: %s, prompt: %s", prompt)

	now := gtime.Now().UnixMilli()

	imgBase64 := ""

	defer func() {
		logger.Infof(ctx, "GenImageBase64 model: %s, len: %d", model, len(imgBase64))
		logger.Infof(ctx, "GenImageBase64 model: %s, totalTime: %d ms", model, gtime.Now().UnixMilli()-now)
	}()

	if len(retry) == 5 {
		return "", errors.New("响应超时, 请重试...")
	}

	reqBase64 := openai.ImageRequest{
		Model:          model,
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := getClient(model).CreateImage(ctx, reqBase64)
	if err != nil {
		logger.Errorf(ctx, "GenImageBase64 model: %s, creation error: %v", model, err)

		e := &openai.APIError{}
		if errors.As(err, &e) {
			switch e.HTTPStatusCode {
			case 400:
				if gstr.Contains(err.Error(), "Your request was rejected as a result of our safety system") {
					return "", err
				}
				time.Sleep(5 * time.Second)
				Init(ctx, model)
				return GenImageBase64(ctx, model, prompt, append(retry, 1)...)
			default:
				time.Sleep(5 * time.Second)
				Init(ctx, model)
				return GenImageBase64(ctx, model, prompt, append(retry, 1)...)
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
