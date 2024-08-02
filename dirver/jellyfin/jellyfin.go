package jellyfin

import "github.com/go-resty/resty/v2"

type DirverJellyfin struct {
	Url         string
	Token       string
	RestyClient *resty.Client
}
