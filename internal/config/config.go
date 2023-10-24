package config

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfsnotify"
	"github.com/iimeta/iim-sdk/utility/logger"
	"time"
)

var Cfg *Config

func init() {

	file, _ := gcfg.NewAdapterFile()
	path, _ := file.GetFilePath()

	if err := gjson.Unmarshal(gjson.MustEncode(gcfg.Instance().MustData(gctx.New())), &Cfg); err != nil {
		panic(fmt.Sprintf("解析配置文件 %s 错误: %v", path, err))
	}

	// 监听配置文件变化, 热加载
	_, _ = gfsnotify.Add(path, func(event *gfsnotify.Event) {
		ctx := gctx.New()
		data, err := gcfg.Instance().Data(ctx)
		if err != nil {
			logger.Errorf(ctx, "热加载 获取配置文件 %s 数据错误: %v", path, err)
		} else {
			if err = gjson.Unmarshal(gjson.MustEncode(data), &Cfg); err != nil {
				logger.Errorf(ctx, "热加载 解析配置文件 %s 错误: %v", path, err)
			}
		}
	})
}

// 配置信息
type Config struct {
	Sdk        *Sdk        `json:"sdk"`
	Filesystem *Filesystem `json:"filesystem"`
	Http       *Http       `json:"http"`
}

type Sdk struct {
	OpenAI     *OpenAI     `json:"openai"`
	Baidu      *Baidu      `json:"baidu"`
	Xfyun      *Xfyun      `json:"xfyun"`
	Aliyun     *Aliyun     `json:"aliyun"`
	Midjourney *Midjourney `json:"midjourney"`
}

type OpenAI struct {
	Models map[string]*Model `json:"models"`
}

type Baidu struct {
	AccessToken *AccessToken      `json:"access_token"`
	Models      map[string]*Model `json:"models"`
}

type AccessToken struct {
	BaseUrl  string `json:"base_url"`
	Path     string `json:"path"`
	ProxyUrl string `json:"proxy_url"`
}

type Xfyun struct {
	OriginalUrl string            `json:"original_url"`
	Models      map[string]*Model `json:"models"`
}

type Aliyun struct {
	Models map[string]*Model `json:"models"`
}

type Model struct {
	BaseUrl   string   `json:"base_url"`
	Path      string   `json:"path"`
	ProxyUrl  string   `json:"proxy_url"`
	ApiKeys   []string `json:"api_keys"`
	Apps      []*App   `json:"apps"`
	MaxTokens int      `json:"max_tokens"`
	Domain    string   `json:"domain"` // 星火模型特有
}

type App struct {
	Id     string `json:"id"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type Midjourney struct {
	CdnUrl          string           `json:"cdn_url"`
	ProxyUrl        string           `json:"proxy_url"`
	MidjourneyProxy *MidjourneyProxy `json:"midjourney_proxy"`
}

type MidjourneyProxy struct {
	CdnProxyUrl     string `json:"cdn_proxy_url"`
	CdnOriginalUrl  string `json:"cdn_original_url"`
	ApiSecret       string `json:"api_secret"`
	ApiSecretHeader string `json:"api_secret_header"`
	ImagineUrl      string `json:"imagine_url"`
	SimpleChangeUrl string `json:"simple_change_url"`
	ChangeUrl       string `json:"change_url"`
	DescribeUrl     string `json:"describe_url"`
	BlendUrl        string `json:"blend_url"`
	FetchUrl        string `json:"fetch_url"`
}

type Filesystem struct {
	Default string      `json:"default"`
	Local   LocalSystem `json:"local"`
	Oss     OssSystem   `json:"oss"`
	Qiniu   QiniuSystem `json:"qiniu"`
	Cos     CosSystem   `json:"cos"`
}

// 本地存储
type LocalSystem struct {
	Root   string `json:"root"`
	Domain string `json:"domain"`
}

// 阿里云 OSS 文件存储
type OssSystem struct {
	AccessID     string `json:"access_id"`
	AccessSecret string `json:"access_secret"`
	Bucket       string `json:"bucket"`
	Endpoint     string `json:"endpoint"`
}

// 七牛云文件存储
type QiniuSystem struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Bucket    string `json:"bucket"`
	Domain    string `json:"domain"`
}

// 腾讯云 COS 文件存储
type CosSystem struct {
	SecretId  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
}

type Http struct {
	Timeout   time.Duration `json:"timeout"`
	ProxyOpen bool          `json:"proxy_open"`
	ProxyUrl  string        `json:"proxy_url"`
}

func Get(ctx context.Context, pattern string, def ...interface{}) (*gvar.Var, error) {

	value, err := g.Cfg().Get(ctx, pattern, def...)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func GetString(ctx context.Context, pattern string, def ...interface{}) string {

	value, err := Get(ctx, pattern, def...)
	if err != nil {
		logger.Error(ctx, err)
	}

	return value.String()
}

func GetInt(ctx context.Context, pattern string, def ...interface{}) int {

	value, err := Get(ctx, pattern, def...)
	if err != nil {
		logger.Error(ctx, err)
	}

	return value.Int()
}

func GetBool(ctx context.Context, pattern string, def ...interface{}) (bool, error) {

	value, err := Get(ctx, pattern, def...)
	if err != nil {
		return false, err
	}

	return value.Bool(), nil
}

func GetMapStrStr(ctx context.Context, pattern string, def ...interface{}) map[string]string {

	value, err := Get(ctx, pattern, def...)
	if err != nil {
		logger.Error(ctx, err)
	}

	return value.MapStrStr()
}
