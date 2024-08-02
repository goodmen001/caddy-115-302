# caddy 302 115


> 一个 caddy 的服务器支持将 115直接 302 到 115 网盘去



## docker 方式安装
``` sh
# docker-compose的配置
version: '3.8'

services:
    media302:
      image: jianxcao/redir115:latest
      container_name: media302
      volumes:
        # 这个目录是配置目录，Caddyfile 将被生成在这个目录下，可以针对请求做各种变化
        - /share/docker/media302/config:/config
      environment:
        - MEDIA_SERVER=这里填 jellyfin,emby 的服务器地址
        - MEDIA_TOKEN=这里填jellyfin,emby的 token
        - COOKIE_115="这里填 115 的 cookie"
        # 原始路径，即你在 jellyfin 下挂载的路径，
        - ORIGIN_PATH=/share/pan/cd/115/disk
        # 替换路径，即真实 115 网盘的路径，比如ORIGIN_PATH最后会被替换成 /disk
        - REPLACE_PATH=/disk
        # 这个可以不需要如果你不是混合内容的话，混合内容指部分资源在本地，部分资源在 115，所以会以这个路径开始的区分是不是 115 的资源
        - MATCH_REDIR_115=/share/pan/cd/115
      ports:
        - "8090:8082"

```


``` sh
# docker run 的命令

docker run -d \
  --name media302 \
  -v /share/docker/media302/config:/config \
  -e MEDIA_SERVER="这里填 jellyfin,emby 的服务器地址" \
  -e MEDIA_TOKEN="这里填 jellyfin,emby 的 token" \
  -e COOKIE_115="这里填 115 的 cookie" \
  -e ORIGIN_PATH="/share/pan/cd/115/disk" \
  -e REPLACE_PATH="/disk" \
  -e MATCH_REDIR_115="/share/pan/cd/115" \
  -p 8090:8082 \
  jianxcao/redir115:latest

```

## caddy模式
> 因为是一个 caddy 的插件，所以可以自己构建一个 caddy，然后使用

``` sh
# 主要先去https://github.com/caddyserver/xcaddy下载自己系统需要的 xcaddy，放在项目目录下 
# 然后构建 caddy，如果需要其他插件包，可以自己自行增加
sh ./build-caddy.sh
```


## 其他 env 说明
``` sh
# 这里是302到 115 连接在内存中占用空间大小，默认64M,不懂不要改
export CACHE115_SZIE=64
# 这里是302到 115连接的时效性，目前是 15 分钟，不懂不要改
export CACHE115=15
# 这 2 个是 caddy 的缓存配置
export BADGER_CACHE=/config/badger/cache
export BADGER_CONFIG=/config/badger/config
``` 


> 最后/config/caddyfile里面有具体的说明，如果不想缓存图片和字幕可以自行删除相关的配置