package midjourney

import (
	"context"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/iimeta/iim-sdk/internal/config"
	"github.com/iimeta/iim-sdk/internal/model"
	"github.com/iimeta/iim-sdk/internal/service"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/sdk"
	"github.com/iimeta/iim-sdk/utility/util"
	"net/url"
)

type sMidjourney struct{}

func init() {
	service.RegisterMidjourney(New())
}

func New() service.IMidjourney {
	return &sMidjourney{}
}

func (s *sMidjourney) Image(ctx context.Context, robot *model.Robot, message *model.Message) (imageInfo *model.Image, err error) {

	if gstr.HasPrefix(message.Prompt, "UPSCALE") || gstr.HasPrefix(message.Prompt, "VARIATION") {
		imageInfo, err = sdk.MidjourneyProxyChanges(ctx, message.Prompt)
	} else {
		imageInfo, err = sdk.MidjourneyProxy(ctx, message.Prompt)
	}

	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	logger.Infof(ctx, "Midjourney Image URL: %s", imageInfo.Url)

	if imageInfo.Size == 0 {

		cdn_url, err := config.Get(ctx, "midjourney.cdn_url")
		if err != nil {
			logger.Error(ctx, err)
			return nil, err
		}

		if cdn_url.String() != "" {

			imageInfo.Size = 1024 * 1024 * 5
			imageInfo.Width = 512
			imageInfo.Height = 512

			_ = grpool.AddWithRecover(ctx, func(ctx context.Context) {

				imgBytes := util.HttpDownloadFile(ctx, imageInfo.Url, false)

				if len(imgBytes) != 0 {
					_, err = service.File().SaveImage(ctx, imgBytes, gfile.Ext(imageInfo.Url), gfile.Basename(imageInfo.Url))
					if err != nil {
						logger.Error(ctx, err)
						return
					}
				} else {
					logger.Errorf(ctx, "HttpDownloadFile %s fail", imageInfo.Url)
				}

			}, nil)

			originalUrl, err := url.Parse(imageInfo.Url)
			if err != nil {
				logger.Error(ctx, err)
				return nil, err
			}

			// 替换CDN
			imageInfo.Url = cdn_url.String() + originalUrl.Path

		} else {

			imgBytes := util.HttpDownloadFile(ctx, imageInfo.Url, false)

			if len(imgBytes) == 0 {
				return nil, err
			}

			imageInfo, err = service.File().SaveImage(ctx, imgBytes, gfile.Ext(imageInfo.Url))
			if err != nil {
				logger.Error(ctx, err)
				return nil, err
			}

			domain, err := config.Get(ctx, "filesystem.local.domain")
			if err != nil {
				logger.Error(ctx, err)
				return nil, err
			}

			imageInfo.Url = domain.String() + "/" + imageInfo.FilePath
		}
	}

	return imageInfo, nil
}
