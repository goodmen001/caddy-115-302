package jellyfin

type MediaSource struct {
	Path             string `json:"Path"`
	IsInfiniteStream bool   `json:"IsInfiniteStream"`
	IsRemote         bool   `json:"IsRemote"`
	Protocol         string `json:"Protocol"`
	Id               string `json:"Id"`
	ETag             string `json:"ETag"`
	Name             string `json:"Name"`
	Size             int64  `json:"Size"`
	VideoType        string `json:"VideoType"`
	Type             string `json:"Type"`
}

type MeidaItem struct {
	MediaSources []MediaSource `json:"MediaSources"`
	Path         string        `json:"Path"`
	Name         string        `json:"Name"`
	Id           string        `json:"Id"`
	ServerId     string        `json:"ServerId"`
	VideoType    string        `json:"VideoType"`
	Type         string        `json:"Type"`
	IsFolder     bool          `json:"IsFolder"`
}

type MediaItemFile struct {
	Path     string `json:"path"`
	ItemName string `json:"itemName"`
	NotLocal bool   `json:"notLocal"`
}
