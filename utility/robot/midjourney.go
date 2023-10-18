package robot

import (
	"context"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/iimeta/iim-sdk/internal/config"
	"github.com/iimeta/iim-sdk/utility/logger"
	"github.com/iimeta/iim-sdk/utility/sdk"
	"github.com/iimeta/iim-sdk/utility/util"
	"net/url"
	"strings"
)

type midjourney struct{}

var Midjourney *midjourney

func init() {
	Midjourney = &midjourney{}
}

func (m *midjourney) Image(ctx context.Context, senderId, receiverId, talkType int, text string, proxy string) (*util.ImageInfo, string, error) {

	if talkType == 2 {
		content := gstr.Split(text, " ")
		if len(content) > 1 {
			text = content[1]
		} else {
			content = gstr.Split(text, " ")
			if len(content) > 1 {
				text = content[1]
			}
		}
	}

	if len(text) == 0 {
		return nil, "", nil
	}

	logger.Infof(ctx, "Midjourney Image prompt: %s", text)

	text = gstr.Replace(text, "\n", "")
	text = gstr.Replace(text, "\r", "")
	text = gstr.TrimLeftStr(text, "/mj")
	text = gstr.TrimLeftStr(text, "/imagine")
	text = strings.TrimSpace(text)

	imageURL := ""
	taskId := ""
	var err error
	var imageInfo *util.ImageInfo

	switch proxy {
	case "midjourney_proxy":
		if gstr.HasPrefix(text, "UPSCALE") || gstr.HasPrefix(text, "VARIATION") {
			taskId, imageInfo, imageURL, err = sdk.MidjourneyProxyChanges(ctx, text)
		} else {
			taskId, imageInfo, imageURL, err = sdk.MidjourneyProxy(ctx, text)
		}
	}

	logger.Infof(ctx, "Midjourney Image imageURL: %s", imageURL)

	if err != nil || imageURL == "" {
		logger.Error(ctx, err)
		return nil, taskId, err
	}

	if imageInfo == nil {

		cdn_url, err := config.Get(ctx, "midjourney.cdn_url")
		if err != nil {
			logger.Error(ctx, err)
		}

		if cdn_url.String() != "" {

			imageInfo = &util.ImageInfo{
				Size:   1024 * 1024 * 5,
				Width:  512,
				Height: 512,
			}

			_ = grpool.AddWithRecover(ctx, func(ctx context.Context) {

				imgBytes := util.HttpDownloadFile(ctx, imageURL, false)

				if len(imgBytes) != 0 {
					_, err = util.SaveImage(ctx, imgBytes, gfile.Ext(imageURL), gfile.Basename(imageURL))
					if err != nil {
						logger.Error(ctx, err)
						return
					}
				} else {
					logger.Errorf(ctx, "HttpDownloadFile %s fail", imageURL)
				}

			}, nil)

			originalUrl, err := url.Parse(imageURL)
			if err != nil {
				logger.Error(ctx, err)
				return nil, taskId, err
			}

			// 替换CDN
			imageURL = cdn_url.String() + originalUrl.Path

		} else {

			imgBytes := util.HttpDownloadFile(ctx, imageURL, false)

			if len(imgBytes) == 0 {
				return nil, taskId, err
			}

			imageInfo, err = util.SaveImage(ctx, imgBytes, gfile.Ext(imageURL))
			if err != nil {
				logger.Error(ctx, err)
				return nil, taskId, err
			}

			domain, err := config.Get(ctx, "filesystem.local.domain")
			if err != nil {
				logger.Error(ctx, err)
				return nil, taskId, err
			}

			imageURL = domain.String() + "/" + imageInfo.FilePath
		}
	}

	logger.Infof(ctx, "SendImage imageURL: %s, Width: %d, Height: %d, Size: %d", imageURL, imageInfo.Width, imageInfo.Height, imageInfo.Size)

	return &util.ImageInfo{
		ImageURL: imageURL,
		Width:    imageInfo.Width,
		Height:   imageInfo.Height,
		Size:     imageInfo.Size,
	}, taskId, nil
}
