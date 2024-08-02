# 使用官方 Go 镜像作为基础镜像
FROM golang:latest as builder

# 安装 xcaddy
# RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
# https://goproxy.io
RUN go env -w GOPROXY=https://goproxy.cn,https://gocenter.io,https://goproxy.io,direct
# RUN echo "$(uname -a)"
RUN apt-get update
RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

# 创建一个工作目录
WORKDIR /build

COPY ./ /build/
RUN pwd

WORKDIR /build
RUN go mod download
RUN go mod tidy

# 从 xcaddy 构建 Caddy 并添加插件
RUN xcaddy build --with github.com/jianxcao/caddy-115-302=./

# 创建运行镜像
FROM alpine:latest

RUN mkdir -p /app
RUN mkdir -p /config
ENV CACHE115_SZIE=64
ENV CACHE115=15
ENV MATCH_REDIR_115=""
ENV BADGER_CACHE=/config/badger/cache
ENV BADGER_CONFIG=/config/badger/config
# 将构建的 Caddy 二进制文件复制到运行镜像中
COPY --from=builder /build/caddy /app/caddy

# 将你的 Caddyfile 复制到镜像中
COPY ./entrypoint /app/entrypoint
RUN chmod +x /app/entrypoint
COPY ./Caddyfile /app/Caddyfile_template
ENTRYPOINT ["sh", "/app/entrypoint" ]

EXPOSE 8082