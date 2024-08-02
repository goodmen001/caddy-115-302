package plugin

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	_115 "github.com/jianxcao/caddy-115-302/dirver/115"
	"github.com/jianxcao/caddy-115-302/dirver/jellyfin"
	"github.com/spf13/cast"

	_ "github.com/caddyserver/cache-handler"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(&Redir302{})
	httpcaddyfile.RegisterHandlerDirective("redir302", parseCaddyfile)
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	m := &Redir302{}
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Redir302 is a simple middleware that logs the start and end time of a request.
type Redir302 struct {
	MediaServer    string   `json:"media_server"`
	Server302      string   `json:"server_302"`
	Token          string   `json:"token"`
	Cookie115      string   `json:"cookie115"`
	Cache115       int      `json:"cache115,omitempty"`
	Cache115Szie   int      `json:"cache115_size,omitempty"`
	MatchRedir115  string   `json:"match_redir_115,omitempty"`
	ReplacePath    []string `json:"replace_path,omitempty"`
	OriginPath     []string `json:"origin_path,omitempty"`
	DirverJellyfin *jellyfin.DirverJellyfin
	Dirver115      *_115.Driver115
	log            *zap.Logger
	Cache          *bigcache.BigCache
}

func (*Redir302) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.redir302",
		New: func() caddy.Module { return new(Redir302) },
	}
}

func (t *Redir302) Provision(ctx caddy.Context) error {
	t.log = ctx.Logger(t)
	if t.MediaServer == "" || t.Token == "" || !strings.HasPrefix(t.MediaServer, "http") {
		t.log.Error("server or token is empty or not http")
		return errors.New("server or token is empty or not http")
	}
	t.DirverJellyfin = &jellyfin.DirverJellyfin{
		Url:   t.MediaServer,
		Token: t.Token,
	}
	t.DirverJellyfin.Init()
	t.Dirver115 = &_115.Driver115{}
	err := t.Dirver115.Init(t.Cookie115, t.log)
	if err != nil {
		t.log.Error("115登录出现错误，请重新设置 cookie")
		// 这里可以尝试扫码登录？？
		return err
	}
	if t.Cache115Szie == 0 {
		t.Cache115Szie = 64
	}
	t.log.Debug("init Cache start", zap.Int("size", t.Cache115Szie), zap.Duration("expire", time.Duration(t.Cache115)*time.Minute))
	if t.Cache115 > 0 {
		cfg := bigcache.DefaultConfig(time.Duration(t.Cache115) * time.Minute)
		cfg.HardMaxCacheSize = t.Cache115Szie
		cache, err := bigcache.New(context.Background(), cfg)
		t.log.Debug("init cache success", zap.Int("size", t.Cache115Szie), zap.Duration("expire", time.Duration(t.Cache115)*time.Minute))
		if err != nil {
			return err
		}
		t.Cache = cache
	}
	t.log.Info("init redir302",
		zap.String("server", t.MediaServer),
		zap.String("token", t.Token),
		zap.Any("originPath", t.OriginPath),
		zap.Any("replacePath", t.ReplacePath),
		zap.String("cookie115", t.Cookie115))
	return nil
}

// 根据配置替换路由
func (t *Redir302) mappingPath(p string) string {
	rp := p
	if len(t.OriginPath) == 0 {
		return p
	}
	for idx, item := range t.OriginPath {
		if strings.HasPrefix(p, item) && t.ReplacePath[idx] != "" {
			rp = strings.Replace(p, item, t.ReplacePath[idx], 1)
		}
	}
	return rp
}

func (t *Redir302) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	uri := r.URL.Path
	method := r.Method
	if method != "GET" {
		err := next.ServeHTTP(w, r)
		return err
	}
	// 只处理 get 请求
	// MediaSourceId
	// fuck，客户端取参数就不一样
	mediaSourceId := t.getMediaSourceId(r)
	if mediaSourceId == "" {
		t.log.Info("拦截请求失败,未获取到MediaSourceId", zap.String("uri", uri))
		err := next.ServeHTTP(w, r)
		return err
	}
	cacheKey := t.getCacheKey(r)
	if t.Cache != nil {
		t.log.Debug("cacheKey", zap.String("cacheKey", cacheKey))
		if v, err := t.Cache.Get(cacheKey); err == nil {
			u := string(v)
			if u != "" {
				time.Sleep(350 * time.Millisecond)
				t.log.Info("命中缓存，从缓存中获取 url", zap.String("cacheKey", cacheKey), zap.String("url", u))
				t.redirUrl(w, r, u)
				return nil
			}
		}
	}
	// 拦截请求
	t.log.Info("拦截请求", zap.String("uri", uri), zap.String("MediaSourceId", mediaSourceId))

	res, err := t.DirverJellyfin.GetItemFilePath(&jellyfin.ReqItemInfo{
		OriginUri:     uri,
		ApiKey:        t.Token,
		MediaSourceId: mediaSourceId,
	})
	t.log.Info("拦截请求结果", zap.Any("res", res), zap.Error(err))
	if err != nil || res.Path == "" {
		err = next.ServeHTTP(w, r)
		return err
	} else {
		// if res.NotLocal {
		// 	// 不是本地文件，先直接返回原始数据
		// 	err := next.ServeHTTP(w, r)
		// 	return err
		// 	t.log.Warn("notLocal decodeURIComponent embyRes.path", zap.String("embyRes.path", res.Path))
		// 	if p, err := url.QueryUnescape(res.Path); err == nil {
		// 		res.Path = p
		// 	}
		// }
		// 检查前缀，只有前缀符合的才进行 115 定向
		if t.MatchRedir115 != "" {
			if !strings.Contains(res.Path, t.MatchRedir115) {
				t.log.Info("取消重定向，因为路径不符合MatchRedir115", zap.String("res.Path", res.Path), zap.String("MatchRedir115", t.MatchRedir115))
				err := next.ServeHTTP(w, r)
				return err
			}
		}
		res.Path = strings.ReplaceAll(res.Path, "\\", "/")
		res.Path = t.mappingPath(res.Path)
		t.log.Info("拦截替换后的 path", zap.String("res.Path", res.Path))
		// 发现自定义了，替换服务
		if strings.HasPrefix(t.Server302, "http") {
			t.log.Info("发现自定义 302 服务", zap.String("Server302", t.Server302))
			if t.Server302[len(t.Server302)-1] == '/' {
				t.Server302 = t.Server302[:len(t.Server302)-1]
			}
			u := fmt.Sprintf("%s%s", t.Server302, res.Path)
			t.log.Info("重定向到 url", zap.String("url", u))
			http.Redirect(w, r, u, http.StatusFound)
			return nil
		} else {
			url, err := t.Dirver115.GetRedirUrl(res.Path, r.Header.Get("User-Agent"))
			if err != nil || url == "" {
				t.log.Warn("获取 115 重定向 url 失败", zap.Error(err), zap.String("path", res.Path))
				err = next.ServeHTTP(w, r)
				return err
			}
			t.log.Info("重定向到 url", zap.String("url", url))
			t.redirUrl(w, r, url)
			t.afterRedir(r, url)
			return nil
		}
	}
}

func (t *Redir302) redirUrl(w http.ResponseWriter, r *http.Request, url string) {
	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte(url))
	http.Redirect(w, r, url, http.StatusFound)
}

func (t *Redir302) getMediaSourceId(r *http.Request) string {
	query := r.URL.Query()
	// MediaSourceId
	// fuck，客户端取参数就不一样
	MediaSourceId := query.Get("mediaSourceId")
	if MediaSourceId == "" {
		MediaSourceId = query.Get("MediaSourceId")
	}
	return MediaSourceId
}

func (t *Redir302) getCacheKey(r *http.Request) string {
	ua := r.Header.Get("User-Agent")
	uri := r.URL.Path
	MediaSourceId := t.getMediaSourceId(r)
	key := fmt.Sprintf("%s:%s:%s", uri, MediaSourceId, ua)
	hash := md5.New()
	hash.Write([]byte(key))
	hashBytes := hash.Sum(nil)
	hashString := fmt.Sprintf("%x", hashBytes)
	return hashString
}

func (t *Redir302) afterRedir(r *http.Request, redirUrl string) {
	key := t.getCacheKey(r)
	if t.Cache != nil {
		t.log.Debug("set cache", zap.String("key", key), zap.String("redirUrl", redirUrl))
		t.Cache.Set(key, []byte(redirUrl))
	}
}

func (t *Redir302) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "match_redir_115":
				if !d.NextArg() {
					continue
				}
				t.MatchRedir115 = d.Val()
			case "cache115":
				if !d.NextArg() {
					continue
				}
				val := cast.ToInt(d.Val())
				if val > 0 {
					t.Cache115 = val
				}
			case "cache115_size":
				if !d.NextArg() {
					continue
				}
				val := cast.ToInt(d.Val())
				if val > 0 {
					t.Cache115Szie = val
				}
			case "api_key":
				if !d.NextArg() {
					return d.Err("请输入 jellyfin 或 emby的api_key")
				}
				t.Token = d.Val()
			case "server_302":
				if !d.NextArg() {
					continue
				}
				if d.Val() != "" {
					t.Server302 = d.Val()
				}
			case "media_server":
				if !d.NextArg() {
					return d.Err("请输入媒体服务器的地址： 如：http://127.0.0.1:8096/")
				}
				t.MediaServer = d.Val()
			case "replace_path":
				if !d.NextArg() {
					continue
				}
				p := d.Val()
				if p == "" {
					continue
				}
				if !strings.HasPrefix(p, "/") {
					p = "/" + p
				}
				t.ReplacePath = append(t.ReplacePath, p)
			case "origin_path":
				if !d.NextArg() {
					continue
				}
				p := d.Val()
				if p == "" {
					continue
				}
				if !strings.HasPrefix(p, "/") {
					p = "/" + p
				}
				t.OriginPath = append(t.OriginPath, p)
			case "cookie115":
				if !d.NextArg() {
					return d.Err("请输入 115 网盘的 cookie")
				}
				t.Cookie115 = d.Val()
			default:
				return d.Errf("unknown property %s", d.Val())
			}
		}
	}
	return nil
}

var (
	_ caddyhttp.MiddlewareHandler = (*Redir302)(nil)
	_ caddyfile.Unmarshaler       = (*Redir302)(nil)
	_ caddy.Provisioner           = (*Redir302)(nil)
)
