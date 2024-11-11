# SSH 证书支持和生成
由于某些服务器的 SSH Server 版本较低，可能不支持某些证书类型，例如 RSA。以下内容介绍了如何检查支持的证书类型，以及如何生成相应的密钥。

## 检查 SSH 服务器支持的密钥类型
检查服务器支持哪些密钥类型，可以使用以下命令：
```shell
ssh -Q key

# ssh-ed25519
# ssh-ed25519-cert-v01@openssh.com
# ssh-rsa
# ssh-dss
# ecdsa-sha2-nistp256
# ecdsa-sha2-nistp384
# ecdsa-sha2-nistp521
# ssh-rsa-cert-v01@openssh.com
# ssh-dss-cert-v01@openssh.com
# ecdsa-sha2-nistp256-cert-v01@openssh.com
# ecdsa-sha2-nistp384-cert-v01@openssh.com
# ecdsa-sha2-nistp521-cert-v01@openssh.com
```
通过检查输出，可以判断 SSH 服务器是否支持某些特定的密钥类型或证书格式。如果你发现某种证书格式不受支持，可以选择其他类型，如 ecdsa-sha2-nistp256 或 ssh-ed25519。

## 生成 ECDSA SSH 密钥对
如果服务器支持 ecdsa-sha2-nistp256 类型的密钥证书，可以使用以下命令生成 ECDSA SSH 密钥对：
```shell
ssh-keygen -t ecdsa -b 256 -f ~/.ssh/ecdsa_nistp256_key -C ""
```
