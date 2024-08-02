package jellyfin

import (
	"fmt"
	"regexp"
	"strings"
)

// 获取itemId
func (d *DirverJellyfin) GetItemIdByUri(uri string) string {
	// 从字符串中删除 "emby" 和 "Sync"
	uri = strings.Replace(uri, "emby", "", 1)
	uri = strings.Replace(uri, "Sync", "", 1)
	uri = strings.ReplaceAll(uri, "-", "")

	// 使用正则表达式匹配字母和数字
	re := regexp.MustCompile(`[A-Za-z0-9]+`)
	matches := re.FindAllString(uri, -1)

	// 检查匹配结果的长度并返回第二个匹配项
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// 获取请求 item 的地址
// mediaSourceId emby 或者 jellyfin的文件id
func (d *DirverJellyfin) GetItemInfoUrl(reqItemInfo *ReqItemInfo) *ReqItemInfo {
	if reqItemInfo.MediaSourceId == "" || reqItemInfo.OriginUri == "" {
		return reqItemInfo
	}
	itemId := d.GetItemIdByUri(reqItemInfo.OriginUri)
	reqItemInfo.ItemId = itemId
	if strings.Contains(reqItemInfo.OriginUri, "JobItems") {
		reqItemInfo.ItemInfoUri = fmt.Sprintf("%s/Sync/JobItems?api_key=%s", d.Url, reqItemInfo.ApiKey)
	} else {
		if reqItemInfo.MediaSourceId != "" {
			// before is GUID like "3c25399d9cbb41368a5abdb71cfe3dc9", V4.9.0.25 is "mediasource_447039" fomrmat
			// 447039 is't main itemId, is mutiple video mediaSourceId
			newMediaSourceId := ""
			if strings.HasPrefix(reqItemInfo.MediaSourceId, "mediasource_") {
				newMediaSourceId = strings.Replace(reqItemInfo.MediaSourceId, "mediasource_", "", 1)
			} else {
				newMediaSourceId = reqItemInfo.MediaSourceId
			}
			reqItemInfo.ItemInfoUri = fmt.Sprintf(`%s/Items?Ids=%s&Fields=Path,MediaSources&Limit=1&api_key=%s`, d.Url, newMediaSourceId, reqItemInfo.ApiKey)
		} else {
			reqItemInfo.ItemInfoUri = fmt.Sprintf(`%s/Items?Ids=%s&Fields=Path,MediaSources&Limit=1&api_key=%s`, d.Url, itemId, reqItemInfo.ApiKey)
		}
	}
	return reqItemInfo
}

// CheckIsStrmByPath 检查给定的文件路径是否以 ".strm" 结尾
func checkIsStrmByPath(filePath string) bool {
	if filePath != "" {
		// 将文件路径转换为小写并检查是否以 ".strm" 结尾
		return strings.HasSuffix(strings.ToLower(filePath), ".strm")
	}
	return false
}
