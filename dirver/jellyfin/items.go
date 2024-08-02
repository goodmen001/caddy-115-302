package jellyfin

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-resty/resty/v2"
)

type MeidaJobItem struct {
	Id         string `json:"Id"`
	OutputPath string `json:"OutputPath"`
}

type meidaItemResult struct {
	Items []MeidaItem `json:"Items"`
}

type mediaItemJobResult struct {
	Items []MeidaJobItem `json:"Items"`
}

func (d *DirverJellyfin) Init() *resty.Client {
	d.RestyClient = resty.New()
	return d.RestyClient
}

// 获取这个 item 的文件路径
func (d *DirverJellyfin) GetItemFilePath(reqItemInfo *ReqItemInfo) (*MediaItemFile, error) {
	// 获取 itemurl
	d.GetItemInfoUrl(reqItemInfo)
	result := &MediaItemFile{}
	contentType := []string{"application/json;charset=utf-8"}
	contentLen := []string{"0"}
	headers := map[string][]string{
		"Content-Length": contentLen,
		"Content-Type":   contentType,
	}
	// proxyURL := "http://127.0.0.1:9090" // 替换为你的代理服务器地址
	// d.RestyClient.SetProxy(proxyURL)
	resp, err := d.RestyClient.R().
		SetHeaderMultiValues(headers).
		// SetResult(&result).
		Get(reqItemInfo.ItemInfoUri)
	if err != nil {
		fmt.Println(err)
		return result, err
	}
	if resp.IsError() {
		return result, fmt.Errorf("error: emby_api %d %s", resp.StatusCode(), resp.Status())
	}
	if strings.Contains(reqItemInfo.ItemInfoUri, "JobItems") {
		res := &mediaItemJobResult{}
		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return result, err
		}
		items := res.Items
		idx := slices.IndexFunc(items, func(m MeidaJobItem) bool { return m.Id == reqItemInfo.ItemId })
		if idx > -1 {
			result.Path = items[idx].OutputPath
			result.NotLocal = checkIsStrmByPath(items[idx].OutputPath)
		} else {
			return result, errors.New("error: emby_api /JobItems response is null")
		}
		return result, nil
	} else {
		res := &meidaItemResult{}
		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return result, err
		}
		if len(res.Items) <= 0 {
			return result, errors.New("error: emby_api /Items response is null")
		}
		// item.MediaSources on Emby has one, on Jellyfin has many!
		item := res.Items[0]
		if item.MediaSources != nil && len(item.MediaSources) > 0 {
			var mediaSource MediaSource
			// ETag only on Jellyfin
			if reqItemInfo.Tag != "" {
				idx := slices.IndexFunc(item.MediaSources, func(m MediaSource) bool { return m.ETag == reqItemInfo.Tag })
				if idx > -1 {
					mediaSource = item.MediaSources[idx]
				}
			}
			if reqItemInfo.MediaSourceId != "" {
				idx := slices.IndexFunc(item.MediaSources, func(m MediaSource) bool { return m.Id == reqItemInfo.MediaSourceId })
				if idx > -1 {
					mediaSource = item.MediaSources[idx]
				} else {
					return result, errors.New("error: emby_api mediaSourceId " + reqItemInfo.MediaSourceId + " not found")
				}
			}
			result.Path = mediaSource.Path
			result.ItemName = item.Name
			/**
			 * note1: MediaSourceInfo{ Protocol }, String ($enum)(File, Http, Rtmp, Rtsp, Udp, Rtp, Ftp, Mms)
			 * note2: live stream "IsInfiniteStream": true
			 * eg1: MediaSourceInfo{ IsRemote }: true
			 * eg1: MediaSourceInfo{ IsRemote }: false, but MediaSourceInfo{ Protocol }: File, this is scraped
			 */
			result.NotLocal = mediaSource.IsInfiniteStream || mediaSource.IsRemote || checkIsStrmByPath(item.Path)
			return result, nil
		} else {
			// "MediaType": "Photo"... not have "MediaSources" field
			result.Path = item.Path
			result.ItemName = item.Name
			return result, nil
		}
	}
}
