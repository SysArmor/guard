# Guard Client

Guard Client 是一个用于管理 SSH 配置和证书的客户端工具，支持更新授权用户列表、撤销证书密钥和启动守护进程。此工具可以帮助自动化 SSH 相关的运维任务，并提供便捷的命令行接口。

## 功能简介

- **SSHD 配置管理**：初始化并管理 `sshd` 配置文件，包括授权用户列表、受信任的用户 CA 密钥以及撤销的密钥。
- **守护进程**：定期更新 SSH CA、授权用户及撤销的密钥，确保安全性和自动化管理。
- **撤销密钥更新**：更新 SSH 撤销证书列表（CRL）。
- **授权用户更新**：管理和更新 SSH 证书的授权用户列表。

## 安装

你可以从源代码构建此工具。
```shell
git clone git@github.com:SysArmor/guard.git
make build-client
```

## 使用方法
以下是 Guard Client 的主要命令和使用示例：

### 初始化 sshd 配置
>
    - **重启 `sshd` 服务**：在完成 `sshd` 配置初始化后，必须重启 `sshd` 服务以使更改生效。使用以下命令重启服务：

      ```shell
      systemctl restart sshd
      ```

    - **更新 CA 和授权用户列表**：在关闭当前会话之前，务必完成 CA 和授权用户列表的更新，或者启动守护进程等待第一次自动更新完成。如果未执行此操作，所有用户将无法登录系统。确保执行以下命令：

      ```shell
      ./guard-client ca --address=<ADDRESS> --node-id=<NODE_ID> --node-secret=<NODE_SECRET>
      ./guard-client update-principals --address=<ADDRESS> --node-id=<NODE_ID> --node-secret=<NODE_SECRET>

      # OR
      ./guard-client daemon --section all --cron "0 0/5 * * *"  --address=<ADDRESS> --node-id=<NODE_ID> --node-secret=<NODE_SECRET>
      ```


```shell
./guard-client init-sshd-config --sshd-config-dir /etc/ssh/sshd_config.d/ --file-name guard.conf
```
init-sshd-config：初始化 sshd 配置文件，包括授权用户列表和 CA 密钥。
- --sshd-config-dir：指定 sshd 配置文件目录，默认为 /etc/ssh/sshd_config.d/。
- --file-name：指定配置文件名称，默认为 guard.conf。
  
### 启动守护进程
```
./guard-client daemon --section all --cron "0 0/5 * * *"  --address=<ADDRESS> --node-id=<NODE_ID> --node-secret=<NODE_SECRET>
```
daemon：启动守护进程，定期更新 SSH 配置。
- --section：指定要运行的部分，支持 all、ca、principals、revoke-keys，默认为 all。
- --cron：指定 cron 表达式来设定任务执行频率，默认为每 5 分钟执行一次。

### 更新CA
更新 SSH CA 的配置信息。
```shell
./guard-client ca --address=<ADDRESS> --node-id=<NODE_ID> --node-secret=<NODE_SECRET>
```

### 更新撤销的密钥
更新 SSH 证书的撤销列表（CRL），确保已被撤销的密钥不会被继续使用。
```shell
./guard-client revoke-keys --address=<ADDRESS> --node-id=<NODE_ID> --node-secret=<NODE_SECRET>
```

### 更新授权用户列表
更新 SSH 授权用户列表，确保只有授权用户才能登录。
```shell
./guard-client update-principals --address=<ADDRESS> --node-id=<NODE_ID> --node-secret=<NODE_SECRET>
```

## 选项说明
- --sshd-config-dir：指定 sshd 配置文件的目录，默认为 /etc/ssh/sshd_config.d/。
- --file-name：指定 sshd 配置文件的名称，默认为 guard.conf。
- --dry-run：启用此选项后将模拟执行命令，不会从server获取数据。
- --address：指定服务器地址 如： https://guard.com
- --node-id：指定节点的唯一标识符。
- --node-secret：指定节点的密钥，用于认证。
