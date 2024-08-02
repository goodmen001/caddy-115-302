package _115

import (
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/allegro/bigcache/v3"
	"go.uber.org/zap"
)

type Driver115 struct {
	Client *driver.Pan115Client
	Cache  *bigcache.BigCache
	log    *zap.Logger
}

func (d *Driver115) Init(cookie string, logger *zap.Logger) error {
	cr := &driver.Credential{}
	cr.FromCookie(cookie)
	d.log = logger
	d.Client = driver.Defalut().ImportCredential(cr)
	// proxyURL := "http://127.0.0.1:8899"
	// d.Client.SetProxy(proxyURL)
	if err := d.Client.LoginCheck(); err != nil {
		return err
	}
	if d.log == nil {
		logger, err := zap.NewProduction() // 你也可以使用 zap.NewDevelopment()
		if err != nil {
			return err
		}
		d.log = logger
		defer logger.Sync() // 确保日志缓冲区写入
	}
	return nil
}

func (d *Driver115) GetDirId(p string) (string, error) {
	result := GetFileIdResponse{}
	req := d.Client.NewRequest().
		SetQueryString(fmt.Sprintf("path=%s", url.PathEscape(p))).
		ForceContentType("application/json;charset=UTF-8").
		SetResult(&result)
	resp, err := req.Get(APIGetFileId)
	if err != nil {
		return "", err
	}
	fmt.Println(string(resp.Body()))
	if resp.StatusCode() != 200 {
		return "", errors.New("request status code is not 200")
	} else if p == "/" { // 根路径处理
		return "0", nil
	} else {
		id, ok := result.Id.(string)
		if !ok {
			return "", errors.New("not found")
		}
		return id, nil
	}
}
func (d *Driver115) GetRedirUrl(p string, ua string) (string, error) {
	if ua != "" {
		d.Client.SetUserAgent(ua)
	}
	dir := path.Dir(p)
	base := path.Base(p)
	dirId, err := d.GetDirId(dir)
	if err != nil {
		return "", err
	}
	files, err := d.Client.ListPage(dirId, 0, 1000)
	if err != nil {
		return "", err
	}
	if len(*files) == 0 {
		return "", errors.New("not found")
	}
	pickCode := ""
	for _, file := range *files {
		if file.Name == base {
			pickCode = file.PickCode
			break
		}
	}
	if pickCode == "" {
		return "", errors.New("not found")
	}
	downloadInfo, err := d.Client.Download(pickCode)
	if err != nil {
		return "", err
	}
	if downloadInfo.Url.Url != "" {
		return downloadInfo.Url.Url, nil
	}
	return "", errors.New("not found")
}
