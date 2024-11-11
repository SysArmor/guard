# 基于角色签发证书

这种情况是指 CA 为一个抽象的角色签发证书，角色并不一定直接对应于具体的系统用户，而是可以表示一类用户或权限组。例如，某些场景中可以为 "管理员" 角色（admin）或 "开发者" 角色（developer）签发证书。
```shell
ssh-keygen -s ca -I went@demo.com -n root  -V +16w -z 2 id_rsa.pub
```

参数解释：
- -s:
    -s表示签署（sign），后面跟的是用于签发证书的 CA 私钥文件。在这个例子中，ca_key是 CA 的私钥文件。
    该私钥用于签署user_key.pub，从而生成相应的用户证书。
- -I:
    -I表示证书的标识符（ID），用于指定证书的名称或身份信息（identity）。
    user_cert是证书的 ID（也称为 "identity" 字段），这个字段帮助 CA 识别和区分不同的证书。它可以是任意字符串，通常用于标记证书的用途或持有者身份，比如用户名、角色名称或描述信息。
- -n:
    -n用于指定证书的 principals（主体）。principals是证书中允许的登录名，也就是可以通过这个证书进行身份验证的用户（或角色）。可以包含多个 principal，用逗号分隔。
- -V:
    -V用于指定证书的有效期。+52w表示证书从签发日起生效，并在 52 周后过期。52w代表 52 周，相当于一年。如果缺省则永久有效
    可以通过类似-V +1d（1 天）、-V +1m（1 个月）这样的方式灵活设置有效期。
    这个参数还可以指定开始和结束的时间，比如-V "2024-10-12T00:00:00Z"表示从某个特定时间开始生效。


## 查看证书
```shell
ssh-keygen -L -f ~/.ssh/id_rsa-cert.pub

#/root/.ssh/id_rsa-cert.pub:
#        Type: ssh-rsa-cert-v01@openssh.com user certificate
#        Public key: RSA-CERT SHA256:LAWr+1Vzy/tLwiGQzq+NBVRgl5E9xJeQvYRTYis8J0o
#        Signing CA: ED25519 SHA256:Dg1NL2svrzSUe/HGljuJ7NJIJQTiRL8N79k0hcYb+3E (using ssh-ed25519)
#        Key ID: "went@demo.com"
#        Serial: 1
#        Valid: from 2024-10-21T17:04:11 to 2024-11-02T06:50:50
#        Principals: 
#                went@demo.com
#        Critical Options: (none)
#        Extensions: 
#                permit-X11-forwarding
#                permit-agent-forwarding
#                permit-port-forwarding
#                permit-pty
#                permit-user-rc
```