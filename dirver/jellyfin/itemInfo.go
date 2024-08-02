package jellyfin

type ReqItemInfo struct {
	MediaSourceId string
	Tag           string
	ApiKey        string
	// 相对路径，不包含域名
	OriginUri string
	// 请求这个 item 的 url
	ItemInfoUri string

	ItemId string
}
