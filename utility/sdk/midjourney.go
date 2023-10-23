package sdk

import (
	"context"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/iimeta/iim-sdk/internal/config"
	"github.com/iimeta/iim-sdk/internal/errors"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/util"
	"time"
)

func MidjourneyProxy(ctx context.Context, prompt string) (*model.Image, error) {

	midjourneyProxy := config.Cfg.Sdk.Midjourney.MidjourneyProxy

	header := make(map[string]string)
	header[midjourneyProxy.ApiSecretHeader] = midjourneyProxy.ApiSecret

	midjourneyProxyImagineReq := &model.MidjourneyProxyImagineReq{
		Prompt: prompt,
	}

	midjourneyProxyImagineRes := new(model.MidjourneyProxyImagineRes)

	err := util.HttpPostJson(ctx, midjourneyProxy.ImagineUrl, header, midjourneyProxyImagineReq, &midjourneyProxyImagineRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxy(ctx, prompt)
	}

	var imageInfo *model.Image
	if midjourneyProxyImagineRes.Result != "" {

		for {
			time.Sleep(3 * time.Second)
			midjourneyProxyFetchRes := new(model.MidjourneyProxyFetchRes)
			imageInfo, midjourneyProxyFetchRes, err = MidjourneyProxyFetch(ctx, midjourneyProxyImagineRes.Result)
			if err != nil {
				logger.Error(ctx, err)
				return nil, gerror.Newf("Prompt: %s, Result: %s", prompt, err.Error())
			}

			logger.Infof(ctx, "midjourneyProxyFetchRes: %s", gjson.MustEncodeString(midjourneyProxyFetchRes))

			if midjourneyProxyFetchRes.Status == "SUCCESS" {
				if imageInfo == nil {
					imageInfo = new(model.Image)
				}
				imageInfo.Url = midjourneyProxyFetchRes.ImageUrl
				imageInfo.TaskId = midjourneyProxyFetchRes.Id
				return imageInfo, nil
			} else if midjourneyProxyFetchRes.Status == "FAILURE" || midjourneyProxyFetchRes.FailReason != "" {
				return nil, errors.New(midjourneyProxyFetchRes.FailReason)
			}
		}
	} else if midjourneyProxyImagineRes.Description != "" {
		return nil, gerror.Newf("Prompt: %s, Result: %s\"%s\"", prompt, midjourneyProxyImagineRes.Description, midjourneyProxyImagineRes.Properties.BannedWord)
	} else {
		return nil, errors.New("未知错误, 请联系作者处理...")
	}
}

func MidjourneyProxyChanges(ctx context.Context, prompt string) (*model.Image, error) {

	prompts := gstr.Split(prompt, "::")
	midjourneyProxyChangeReq := &model.MidjourneyProxyChangeReq{
		Action: prompts[0],
		Index:  gconv.Int(prompts[1]),
		TaskId: prompts[2],
	}

	midjourneyProxyChangeRes, err := MidjourneyProxyChange(ctx, midjourneyProxyChangeReq)

	var imageInfo *model.Image
	if midjourneyProxyChangeRes.Result != "" {

		for {
			time.Sleep(3 * time.Second)
			midjourneyProxyFetchRes := new(model.MidjourneyProxyFetchRes)
			imageInfo, midjourneyProxyFetchRes, err = MidjourneyProxyFetch(ctx, midjourneyProxyChangeRes.Result)
			if err != nil {
				logger.Error(ctx, err)
				return nil, gerror.Newf("Prompt: %s, Result: %s", prompt, err.Error())
			}

			logger.Infof(ctx, "midjourneyProxyFetchRes: %s", gjson.MustEncodeString(midjourneyProxyFetchRes))

			if midjourneyProxyFetchRes.Status == "SUCCESS" {
				if imageInfo == nil {
					imageInfo = new(model.Image)
				}
				imageInfo.Url = midjourneyProxyFetchRes.ImageUrl
				imageInfo.TaskId = midjourneyProxyFetchRes.Id
				return imageInfo, nil
			} else if midjourneyProxyFetchRes.Status == "FAILURE" || midjourneyProxyFetchRes.FailReason != "" {
				return nil, errors.New(midjourneyProxyFetchRes.FailReason)
			}
		}
	} else if midjourneyProxyChangeRes.Description != "" {
		return nil, gerror.Newf("Prompt: %s, Result: %s\"%s\"", prompt, midjourneyProxyChangeRes.Description, midjourneyProxyChangeRes.Properties.BannedWord)
	} else {
		return nil, errors.New("未知错误, 请联系作者处理...")
	}
}

func MidjourneyProxyImagine(ctx context.Context, midjourneyProxyImagineReq *model.MidjourneyProxyImagineReq) (*model.MidjourneyProxyImagineRes, error) {

	midjourneyProxy := config.Cfg.Sdk.Midjourney.MidjourneyProxy

	header := make(map[string]string)
	header[midjourneyProxy.ApiSecretHeader] = midjourneyProxy.ApiSecret

	midjourneyProxyImagineRes := new(model.MidjourneyProxyImagineRes)

	err := util.HttpPostJson(ctx, midjourneyProxy.ImagineUrl, header, midjourneyProxyImagineReq, &midjourneyProxyImagineRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxyImagine(ctx, midjourneyProxyImagineReq)
	}

	return midjourneyProxyImagineRes, nil
}

func MidjourneyProxyChange(ctx context.Context, midjourneyProxyChangeReq *model.MidjourneyProxyChangeReq) (*model.MidjourneyProxyChangeRes, error) {

	midjourneyProxy := config.Cfg.Sdk.Midjourney.MidjourneyProxy

	header := make(map[string]string)
	header[midjourneyProxy.ApiSecretHeader] = midjourneyProxy.ApiSecret

	midjourneyProxyChangeRes := new(model.MidjourneyProxyChangeRes)

	err := util.HttpPostJson(ctx, midjourneyProxy.ChangeUrl, header, midjourneyProxyChangeReq, &midjourneyProxyChangeRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxyChange(ctx, midjourneyProxyChangeReq)
	}

	return midjourneyProxyChangeRes, nil
}

func MidjourneyProxyDescribe(ctx context.Context, midjourneyProxyDescribeReq *model.MidjourneyProxyDescribeReq) (*model.MidjourneyProxyDescribeRes, error) {

	midjourneyProxy := config.Cfg.Sdk.Midjourney.MidjourneyProxy

	header := make(map[string]string)
	header[midjourneyProxy.ApiSecretHeader] = midjourneyProxy.ApiSecret

	midjourneyProxyDescribeRes := new(model.MidjourneyProxyDescribeRes)

	err := util.HttpPostJson(ctx, midjourneyProxy.DescribeUrl, header, midjourneyProxyDescribeReq, &midjourneyProxyDescribeRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxyDescribe(ctx, midjourneyProxyDescribeReq)
	}

	return midjourneyProxyDescribeRes, nil
}

func MidjourneyProxyBlend(ctx context.Context, midjourneyProxyBlendReq *model.MidjourneyProxyBlendReq) (*model.MidjourneyProxyBlendRes, error) {

	midjourneyProxy := config.Cfg.Sdk.Midjourney.MidjourneyProxy

	header := make(map[string]string)
	header[midjourneyProxy.ApiSecretHeader] = midjourneyProxy.ApiSecret

	midjourneyProxyBlendRes := new(model.MidjourneyProxyBlendRes)

	err := util.HttpPostJson(ctx, midjourneyProxy.BlendUrl, header, midjourneyProxyBlendReq, &midjourneyProxyBlendRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxyBlend(ctx, midjourneyProxyBlendReq)
	}

	return midjourneyProxyBlendRes, nil
}

func MidjourneyProxyFetch(ctx context.Context, taskId string) (imageInfo *model.Image, midjourneyProxyFetchRes *model.MidjourneyProxyFetchRes, err error) {

	midjourneyProxy := config.Cfg.Sdk.Midjourney.MidjourneyProxy

	header := make(map[string]string)
	header[midjourneyProxy.ApiSecretHeader] = midjourneyProxy.ApiSecret

	fetchUrl := gstr.Replace(midjourneyProxy.FetchUrl, "${task_id}", taskId, -1)

	midjourneyProxyFetchRes = new(model.MidjourneyProxyFetchRes)
	err = util.HttpGet(ctx, fetchUrl, header, nil, &midjourneyProxyFetchRes)
	if err != nil {
		logger.Error(ctx, err)
		return nil, nil, err
	}

	logger.Infof(ctx, "midjourneyProxyFetchRes: %s", gjson.MustEncodeString(midjourneyProxyFetchRes))

	if midjourneyProxyFetchRes.Status == "SUCCESS" && gfile.ExtName(midjourneyProxyFetchRes.ImageUrl) == "webp" {

		imageUrl := gstr.Replace(midjourneyProxyFetchRes.ImageUrl, midjourneyProxy.CdnProxyUrl, midjourneyProxy.CdnOriginalUrl)

		imgBytes := util.HttpDownloadFile(ctx, imageUrl, config.Cfg.Sdk.Midjourney.ProxyUrl)

		imageInfo, err = service.File().SaveImage(ctx, imgBytes, gfile.Ext(imageUrl)) // todo
		if err != nil {
			logger.Error(ctx, err)
			return nil, nil, err
		}

		imageUrl = config.Cfg.Filesystem.Local.Domain + "/" + imageInfo.FilePath

		midjourneyProxyFetchRes.ImageUrl = imageUrl
	} else if midjourneyProxyFetchRes.Status == "FAILURE" || midjourneyProxyFetchRes.FailReason != "" {
		return nil, midjourneyProxyFetchRes, errors.New(midjourneyProxyFetchRes.FailReason)
	}

	return imageInfo, midjourneyProxyFetchRes, nil
}
