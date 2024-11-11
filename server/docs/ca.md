# CA 和 CA 公钥生成

要生成 CA 和 CA 公钥，可以使用以下命令：

```shell
# 生成 CA 密钥对（ca 和 ca.pub），强制要求输入强制要求passphrase
ssh-keygen -C "CA" -f ca -b 4096
```

- ca：生成的私钥文件，用于签署证书。
- ca.pub：生成的公钥文件，将其分发到所有节点，以便验证签名。

passphrase、ca、ca.pub 都需要配置于config.yaml，其中ca.pub会分发到所有node。