# syntax=docker/dockerfile:1.7

# 二开推荐阅读[如何提高项目构建效率](https://developers.weixin.qq.com/miniprogram/dev/wxcloudrun/src/scene/build/speed.html)
# 选择构建用基础镜像（选择原则：在包含所有用到的依赖前提下尽可能体积小）。如需更换，请到[dockerhub官方仓库](https://hub.docker.com/_/golang?tab=tags)自行选择后替换。
FROM golang:1.22-alpine as builder

ENV GOPROXY=https://goproxy.cn,direct

# 指定构建过程中的工作目录， 之后如果是相对路径，都相对于这个目录[除非是绝对路径]
WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go env -w GOPROXY=https://proxy.golang.org,direct


COPY go.mod go.sum ./

# RUN go mod download
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# 将当前目录（dockerfile所在目录）下所有文件都拷贝到工作目录下（.dockerignore中文件除外）
COPY . /app/

# 执行代码编译命令。操作系统参数为linux，编译后的二进制产物命名为main，并存放在当前目录下。
#RUN GOOS=linux go build -o main .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o main .

# 选用运行时所用基础镜像（GO语言选择原则：尽量体积小、包含基础linux内容的基础镜像）
FROM alpine:latest

# 容器默认时区为UTC，如需使用上海时间请启用以下时区设置命令
# RUN apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo Asia/Shanghai > /etc/timezone

# 使用 HTTPS 协议访问容器云调用证书安装
RUN apk --no-cache add ca-certificates

# 指定运行时的工作目录
WORKDIR /app

# 将构建产物/app/main拷贝到运行时的工作目录中
COPY --from=builder /app/main /app/config.json /app/init.sql /app/
# Ensure default asset exists by copying from frontend images if present

RUN mkdir -p /app/data/user_data

EXPOSE 8080

# 执行启动命令
# 写多行独立的CMD命令是错误写法！只有最后一行CMD命令会被执行，之前的都会被忽略，导致业务报错。
# 请参考[Docker官方文档之CMD命令](https://docs.docker.com/engine/reference/builder/#cmd)
# 如果在 docker-compose.yml 中指定了 command，它会覆盖 Dockerfile 中的 CMD
# CMD ["/app/main"]
