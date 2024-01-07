# 选择构建用基础镜像（选择原则：在包含所有用到的依赖前提下尽可能体积小）。如需更换，请到[dockerhub官方仓库](https://hub.docker.com/_/golang?tab=tags)自行选择后替换。
#FROM golang:1.17.1-alpine3.14 as builder
FROM golang:1.21 as builder

# 指定构建过程中的工作目录
WORKDIR /app

# 将当前目录（dockerfile所在目录）下所有文件都拷贝到工作目录下（.dockerignore中文件除外）
COPY . /app/

# 执行代码编译命令。操作系统参数为linux，编译后的二进制产物命名为main，并存放在当前目录下。
RUN  cd /app/internal/account && \
     go env -w GOPROXY=https://goproxy.cn,direct && go build -o main .

# 选用运行时所用基础镜像（GO语言选择原则：尽量体积小、包含基础linux内容的基础镜像）
#FROM alpine:3.19.0
FROM ubuntu:22.04


# 指定运行时的工作目录
WORKDIR /app

RUN mkdir -p /app/config

COPY --from=builder /app/internal/account/main  /app/

COPY --from=builder /app/internal/account/config/app.env  /app/config/

EXPOSE 8080

CMD ["/app/main"]
