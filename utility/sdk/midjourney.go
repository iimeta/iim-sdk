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

	api_secret, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	api_secret_header, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret_header")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	imagine_url, err := config.Get(ctx, "midjourney.midjourney_proxy.imagine_url")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	header := make(map[string]string)
	header[api_secret_header.String()] = api_secret.String()

	midjourneyProxyImagineReq := &model.MidjourneyProxyImagineReq{
		Prompt: prompt,
	}

	midjourneyProxyImagineRes := new(model.MidjourneyProxyImagineRes)

	err = util.HttpPost(ctx, imagine_url.String(), header, midjourneyProxyImagineReq, &midjourneyProxyImagineRes)
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

	api_secret, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	api_secret_header, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret_header")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	imagine_url, err := config.Get(ctx, "midjourney.midjourney_proxy.imagine_url")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	header := make(map[string]string)
	header[api_secret_header.String()] = api_secret.String()

	midjourneyProxyImagineRes := new(model.MidjourneyProxyImagineRes)

	err = util.HttpPost(ctx, imagine_url.String(), header, midjourneyProxyImagineReq, &midjourneyProxyImagineRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxyImagine(ctx, midjourneyProxyImagineReq)
	}

	return midjourneyProxyImagineRes, nil
}

func MidjourneyProxyChange(ctx context.Context, midjourneyProxyChangeReq *model.MidjourneyProxyChangeReq) (*model.MidjourneyProxyChangeRes, error) {

	api_secret, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	api_secret_header, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret_header")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	change_url, err := config.Get(ctx, "midjourney.midjourney_proxy.change_url")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	header := make(map[string]string)
	header[api_secret_header.String()] = api_secret.String()

	midjourneyProxyChangeRes := new(model.MidjourneyProxyChangeRes)

	err = util.HttpPost(ctx, change_url.String(), header, midjourneyProxyChangeReq, &midjourneyProxyChangeRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxyChange(ctx, midjourneyProxyChangeReq)
	}

	return midjourneyProxyChangeRes, nil
}

func MidjourneyProxyDescribe(ctx context.Context, midjourneyProxyDescribeReq *model.MidjourneyProxyDescribeReq) (*model.MidjourneyProxyDescribeRes, error) {

	api_secret, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	api_secret_header, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret_header")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	describe_url, err := config.Get(ctx, "midjourney.midjourney_proxy.describe_url")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	header := make(map[string]string)
	header[api_secret_header.String()] = api_secret.String()

	midjourneyProxyDescribeRes := new(model.MidjourneyProxyDescribeRes)

	err = util.HttpPost(ctx, describe_url.String(), header, midjourneyProxyDescribeReq, &midjourneyProxyDescribeRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxyDescribe(ctx, midjourneyProxyDescribeReq)
	}

	return midjourneyProxyDescribeRes, nil
}

func MidjourneyProxyBlend(ctx context.Context, midjourneyProxyBlendReq *model.MidjourneyProxyBlendReq) (*model.MidjourneyProxyBlendRes, error) {

	api_secret, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	api_secret_header, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret_header")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	blend_url, err := config.Get(ctx, "midjourney.midjourney_proxy.blend_url")
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	header := make(map[string]string)
	header[api_secret_header.String()] = api_secret.String()

	midjourneyProxyBlendRes := new(model.MidjourneyProxyBlendRes)

	err = util.HttpPost(ctx, blend_url.String(), header, midjourneyProxyBlendReq, &midjourneyProxyBlendRes)
	if err != nil {
		logger.Error(ctx, err)
		time.Sleep(5 * time.Second)
		return MidjourneyProxyBlend(ctx, midjourneyProxyBlendReq)
	}

	return midjourneyProxyBlendRes, nil
}

func MidjourneyProxyFetch(ctx context.Context, taskId string) (imageInfo *model.Image, midjourneyProxyFetchRes *model.MidjourneyProxyFetchRes, err error) {

	fetch_url, err := config.Get(ctx, "midjourney.midjourney_proxy.fetch_url")
	if err != nil {
		logger.Error(ctx, err)
		return nil, nil, err
	}

	api_secret, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret")
	if err != nil {
		logger.Error(ctx, err)
		return nil, nil, err
	}

	api_secret_header, err := config.Get(ctx, "midjourney.midjourney_proxy.api_secret_header")
	if err != nil {
		logger.Error(ctx, err)
		return nil, nil, err
	}

	header := make(map[string]string)
	header[api_secret_header.String()] = api_secret.String()

	fetchUrl := gstr.Replace(fetch_url.String(), "${task_id}", taskId, -1)

	midjourneyProxyFetchRes = new(model.MidjourneyProxyFetchRes)
	err = util.HttpGet(ctx, fetchUrl, header, nil, &midjourneyProxyFetchRes)
	if err != nil {
		logger.Error(ctx, err)
		return nil, nil, err
	}

	logger.Infof(ctx, "midjourneyProxyFetchRes: %s", gjson.MustEncodeString(midjourneyProxyFetchRes))

	if midjourneyProxyFetchRes.Status == "SUCCESS" && gfile.ExtName(midjourneyProxyFetchRes.ImageUrl) == "webp" {

		cdn_proxy_url, err := config.Get(ctx, "midjourney.midjourney_proxy.cdn_proxy_url")
		if err != nil {
			logger.Error(ctx, err)
			return nil, nil, err
		}

		cdn_original_url, err := config.Get(ctx, "midjourney.midjourney_proxy.cdn_original_url")
		if err != nil {
			logger.Error(ctx, err)
			return nil, nil, err
		}

		imageUrl := gstr.Replace(midjourneyProxyFetchRes.ImageUrl, cdn_proxy_url.String(), cdn_original_url.String())

		imgBytes := util.HttpDownloadFile(ctx, imageUrl, true)

		imageInfo, err = service.File().SaveImage(ctx, imgBytes, gfile.Ext(imageUrl)) // todo
		if err != nil {
			logger.Error(ctx, err)
			return nil, nil, err
		}

		domain, err := config.Get(ctx, "filesystem.local.domain")
		if err != nil {
			logger.Error(ctx, err)
			return nil, nil, err
		}

		imageUrl = domain.String() + "/" + imageInfo.FilePath

		midjourneyProxyFetchRes.ImageUrl = imageUrl
	} else if midjourneyProxyFetchRes.Status == "FAILURE" || midjourneyProxyFetchRes.FailReason != "" {
		return nil, midjourneyProxyFetchRes, errors.New(midjourneyProxyFetchRes.FailReason)
	}

	return imageInfo, midjourneyProxyFetchRes, nil
}
