# 赛特乂丕 - 在线 DDNS 工具

[![GitHub release](https://img.shields.io/github/release/dco/ddns-cli.svg)](https://github.com/dco/ddns-cli/releases)
[![License](https://img.shields.io/github/license/dco/ddns-cli.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows%20%7C%20macOS-blue)](https://github.com/dco/ddns-cli/releases)
[![Arch](https://img.shields.io/badge/arch-x86_64%20%7C%20arm64%20%7C%20armv7-green)](https://github.com/dco/ddns-cli/releases)


赛特乂丕是一个免费的在线动态DNS（DDNS）服务，方便您通过域名访问您的家庭网络或设备，即使您的IP地址发生变化。

**官网:** [https://setip.eu.org](https://setip.eu.org)

**客户端:** ddns-cli

**特性:**

* **免费子域名:**  提供免费的子域名，方便您快速上手。
* **IPv4 和 IPv6 支持:** 支持 IPv4 和 IPv6 地址的动态更新。
* **网卡选择:**  可以选择绑定特定网卡的公网 IP。
* **多平台多架构支持:** 提供 Linux、Windows 和 macOS 等多平台的客户端，支持 x86_64、arm64 和 armv7 等多种架构。
* **主动故障切换（待开发）:**  未来将支持主动故障切换，提高服务的稳定性。


## 使用方法

1. **下载 ddns-cli 客户端:**  前往 [Releases](https://github.com/dco/ddns-cli/releases) 页面下载适合您操作系统的客户端。
2. **注册并添加客户端:**  访问官网 [https://setip.eu.org](https://setip.eu.org) 注册账号，并添加您的 ddns-cli 客户端，获取客户端 ID (cid)。
3. **运行 ddns-cli:**  执行以下命令启动客户端：

    ```bash
    ddns-cli -cid <您的客户端ID>
    ```

4. **访问您的子域名:** 几分钟后，您就可以通过分配的子域名访问您的 IP 地址了。

## 示例

```bash
./ddns-cli -cid a1b2c3d4-e5f6-7890-1234-567890abcdef
```
## 加星与分享

如果你觉得这个项目对你有帮助，请给我们加个星 ⭐ 并分享给更多人！