EOS带外区块签名服务器
-----------------------------------

一个为 EOS.IO Software 区块链开发的，简易安全的OOB区块签名服务器，
来自你在 EOS Canada 的朋友。

在EOS.IO 1.0发布之前，Block.one为块生产商引入了带外签名。 
它涉及使用以下内容设置配置：

```
signature-provider=EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV=KEOSD:http://localhost:6666/v1/wallet/sign_digest
keosd-provider-timeout=20   # default value is 5 ms
```

这意味着 `nodeos` 可以通过外部与 `keosd` 兼容的程序，用与 
`EOS6MR ... W5CV` 对应的私钥对区块进行签名。

`eos-blocksigner` 就是这样的一个程序, 而且它结合了 `eosc` 的保险库、
钱包和命令行工具。

**警告**：你*不该*将此软件暴露给任何公共端点，_即使在内部网络中也不行_。 
它应该只在环回接口上运行和监听。 `sign_digest` 端点实际上可以使用关联的私
钥对*任何*内容进行签名。 如果你对某些帐户的 `owner` 权限、`active` 权限使用
相同的私钥（虽然你**不**应该这么做），你则可以使用 `/v1/wallet/sign_digest` 
端点对任何交易进行签名。

## 两种操作模式

截至2018年5月，`eos-blocksigner` 有两种操作模式：

1. 使用通过 Google Cloud Platform 的密钥管理系统加密的保险库
2. 使用纯文本私钥文件

随着需求的增长，我们可以添加更多策略，例如 AWS 的 KMS 系统，密语加密的保险库
或其他一些 HSM 系统。


## GCP（谷歌云平台） KMS 集成

要使用KMS-GCP策略，在 `eosc` 中用以下方式在本地创建保险库：

```
$ eosc vault create --import \
                    --vault-type kms-gcp \
                    --comment "Block signing key vault" \
                    --kms-gcp-keypath projects/PROJNAME/locations/LOC/keyRings/RINGNAME/cryptoKeys/KEYNAME
...
```

这代表你通过 `gcloud` 进行了身份验证，并且在指定的项目和密钥环中具有使用 
KMS *加密*的权限。

然后，在可以把 `eosc-vault.json` 钱包放在生产基础架构上，并使用以下参数运行 `eos-blocksigner`：
```
$ eos-blocksigner --wallet-path path/to/eosc-vault.json \
                  --kms-gcp-keypath projects/PROJNAME/locations/LOC/keyRings/RINGNAME/cryptoKeys/KEYNAME
Listening on 127.0.0.1:6666
```

你需要在服务器上使用 _Decrypt_ KMS 作用域来做保险库的解密。


## 纯文本私钥文件

这是一种不太安全的方法，但仍然比在 `config.ini` 中以纯文本的形式保存私钥更安全。

远程执行代码漏洞通常允许轻松读取进程的内存。 用带外（read: separate process）
签名服务器已经让使用私钥访问内存变得更加复杂， 通过适当的隔离（container、网络访问和
`eos-blocksigner`），你可以降低预料不到的的 `nodeos` 漏洞泄露私钥的风险。

`--keys-file` 是一个看起来像 `myfile.keys` 这样的简单文件:

```
5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFFF
5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3 // This matches EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV
```

它每行都是一个私钥，忽略空格后的任何内容。

有了 keys-file 文件，你不需要 `eosc-vault.json`，可以直接运行：
```
$ eos-blocksigner --keys-file=myfile.keys
Listening on 127.0.0.1:6666
```

# 证书

MIT
