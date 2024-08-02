package main

import (
	"fmt"
	"os"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/jianxcao/caddy-115-302/plugins"
)

func main() {
	p, _ := os.Getwd()
	// 这么 run 要注意设置环境变了，Caddyfile 中会读取的，
	os.Args = []string{"caddy", "run", "--config", fmt.Sprintf("%s/Caddyfile", p)}
	// os.Args = []string{"caddy", "list-modules"}
	caddycmd.Main()
}
