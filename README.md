# Guard- A Guard System for protecting your system

Guard是一个用于保护你的系统安全的系统，它可以允许配置哪些用户可以访问你的系统，现在仅支持Linux系统的SSH访问保护。

Guard 包含服务端和客户端两部分：

Client 端文档
请参阅客户端文档以获取相关信息：[client-readme.md](./client/readme.md)

在接下来的部分中，将详细介绍 server 端的安装与使用方法。


## 安装
Guard使用Golang开发，并可以打包成二进制文件，你可以直接下载二进制文件运行，也可以使用docker进行部署，以下是两种方式的安装方法。

首先，拉取代码：
```shell
git clone git@github.com:SysArmor/guard.git
cd guard
```

### 二进制文件安装
使用以下命令构建 server 端的二进制文件：
```shell
make build-server
```

### 构建docker镜像
使用以下命令构建 Docker 镜像：
```shell
docker build --build-arg -f Dockerfile -t guard:v1alpha1 .
```

## 启动服务
```shell
./guard-server -config=config.yaml
```