# 从原版 gocron 升级到 New-Gocron v2.0.5

本指南面向已安装的 [ouqiang/gocron](https://github.com/ouqiang/gocron)，以原版 v1.5.3 为升级基线。数据库继续使用原来的 MySQL 或 PostgreSQL，不需要重新安装、导入任务或创建新管理员。

早于 v1.5.3、修改过数据库结构的分支，或缺少原有配置文件的环境，请先在数据库副本上验证。不要将本指南视为所有历史版本均可直接升级的保证。

## 1. 升级前确认

- 确认原版程序目录、启动命令、服务管理方式及运行用户。以下 Linux 示例使用 `/usr/local/gocron` 和 `gocron.service`，请按实际情况替换。
- 记录程序版本及 `conf/.version` 原内容，不要手动把版本标记改成 `205`。
- 保留 `conf/app.ini`、`conf/install.lock`、`conf/.version`。数据库连接、表前缀和原有认证配置仍使用原值；不要拿空配置覆盖它们。
- 确认迁移账号具备读写、建表和修改表结构权限，程序运行用户能够读配置、写 `conf/.version` 及运行日志等目录。
- 选择维护窗口，停止所有连接同一数据库的旧版 Web/调度进程，防止升级期间继续调度。已经下发到执行节点的任务不一定随 Web 停止；应先确认重要任务结束。
- 执行节点端口默认 `5921`，应通过云安全组和主机防火墙仅允许主程序服务器 IP 访问，不要开放给公网所有来源。未启用 TLS 时，IP 限制不等于加密或请求鉴权。

## 2. 下载匹配的发行包

从 [v2.0.5 发行页](https://github.com/caoyek/New-Gocron/releases/tag/v2.0.5) 下载主程序及 `SHA256SUMS.txt`，先解压到临时目录，不要直接覆盖旧目录。

| 系统 | 主程序包 | 解压后的目录 |
| --- | --- | --- |
| Linux x86_64 | `new-gocron-v2.0.5-linux-amd64.tar.gz` | `gocron-linux-amd64` |
| Windows x64 | `new-gocron-v2.0.5-windows-amd64.zip` | `gocron-windows-amd64` |
| macOS Intel | `new-gocron-v2.0.5-darwin-amd64.tar.gz` | `gocron-darwin-amd64` |
| macOS Apple Silicon | `new-gocron-v2.0.5-darwin-arm64.tar.gz` | `gocron-darwin-arm64` |

`gocron-node` 是执行节点，不是 Web 主程序。节点通信协议未变，此次升级可保留原版节点，不必同步替换；原 TLS 配置也应保留。

Linux 校验使用 `sha256sum 文件名`，macOS 使用 `shasum -a 256 文件名`，Windows 使用 `Get-FileHash 文件名 -Algorithm SHA256`，与校验文件中对应条目比较。

## 3. 停止服务并备份

先停止旧 Web 服务，确认没有其他守护进程将它重新拉起。若使用 Supervisor、Docker、nohup 或其他服务名称，应使用原来的管理方式停止。

Linux/systemd 示例：

```bash
sudo systemctl stop gocron.service
cd /usr/local/gocron
backup_dir="/usr/local/gocron/backups/pre-v2.0.5-$(date +%Y%m%d-%H%M%S)"
sudo mkdir -p "$backup_dir"
sudo cp -a gocron conf "$backup_dir/"
```

另行备份原启动脚本、服务单元及自定义文件。上述目录备份**不是数据库备份**。

使用数据库管理工具制作可恢复的完整备份，包括表结构、任务、用户、节点和配置表。MySQL/InnoDB 示例（替换用户名和库名，密码交互输入）：

```bash
mysqldump -u 数据库用户 -p --single-transaction --routines --triggers 原数据库名 > /安全可写目录/gocron-before-v2.0.5.sql
```

远程数据库需补充 `-h`、`-P` 等实际连接参数；PostgreSQL 可使用 `pg_dump`。确认命令成功、备份文件有效并妥善保管，文件可能包含账号、机器人地址和密钥。生产环境建议先把备份还原到隔离环境做升级演练；隔离环境不能连接真实执行节点或真实业务接口，以免重复执行任务。

## 4. 仅替换主程序

以下假定下载的主程序已解压到 `/tmp/new-gocron-v2.0.5/`：

```bash
cd /usr/local/gocron
sudo cp /tmp/new-gocron-v2.0.5/gocron-linux-amd64/gocron ./gocron
sudo chmod 0755 ./gocron
./gocron --version
```

保留原运行用户及目录权限，不要顺带切换服务账号。尤其不要覆盖 `conf`、`log`、`outputs` 或自定义脚本目录。主程序已内嵌前端，无需单独安装 Node.js、npm 或替换网页目录。

## 5. 单独执行数据库迁移

保持服务停止，以原服务运行用户在程序目录执行：

```bash
cd /usr/local/gocron
./gocron db-upgrade
```

命令使用程序目录的 `conf/app.ini`，不启动任务调度，可重复执行。若使用受限服务账号，请通过该账号执行；使用其他账号迁移后应检查配置和日志文件的属主，避免服务重启后无权写入。

v2.0.5 会检查并调整任务命令、通知规则、通知接收对象等字段，新增登录安全表，以及 HTTP Header/Body、任务到期时间、日志开关和保留期限六个任务字段。详细字段见 [README 数据库升级说明](../README.md#从旧版本升级)。

看到升级完成且退出码为 `0` 后，才能启动 Web 服务。迁移报错请先保留错误输出并处理，不要反复重启，也不要修改版本文件绕过检查。MySQL 的部分 DDL 不能随事务回滚，迁移失败也可能已产生部分结构变更。

## 6. 启动并验收

systemd 示例：

```bash
sudo systemctl start gocron.service
sudo systemctl status gocron.service --no-pager
sudo journalctl -u gocron.service -n 100 --no-pager
```

没有服务管理器时，可先前台运行 `./gocron web` 查看启动输出，验收后再沿用原来的守护方式；不要同时运行前台和后台两个主程序。

正常启动仍会按原 `conf/.version` 做版本检查，并在版本迁移成功后更新标记。`db-upgrade` 本身不更新该标记；后续启动重复进行幂等的字段检查是正常的。

逐项确认：

1. 浏览器刷新或清缓存后可用原账号登录，页面显示 v2.0.5。
2. 任务、节点、标签、Cron 和通知配置与升级前一致。
3. 用无副作用的测试任务验证 HTTP、Shell、日志和通知；不要直接手动执行数据库变更或删除任务来验收。
4. 旧 HTTP POST 任务继续使用原有 URL 问号参数作为表单提交；只有显式配置 JSON Body 的任务才使用 JSON 模式。
5. 旧任务默认不过期、开启日志并永久保留；新建任务默认保留日志 90 天，定期清理不删除执行中的日志。
6. 登录安全默认按 IP 和账号统计：10 分钟失败 5 次封禁 30 分钟。反向代理环境中当前取 TCP 来源 IP，可能是代理地址；启用白名单前先确认页面显示的访问 IP，防止把自己锁在外面。

若旧 MySQL 的 `notify_receiver_id` 曾为整数，迁移会修复类型，但已丢失并变成 `0` 的企微群选择不能恢复，需要重新选择并保存。

## 7. Windows 与 macOS

同样遵循“停服务 → 备份数据库和程序配置 → 替换主程序 → 迁移 → 启动”的顺序。Windows 在原目录执行：

```powershell
.\gocron.exe --version
.\gocron.exe db-upgrade
# 确认成功后再启动，或恢复原 Windows 服务。
.\gocron.exe web
```

macOS 选择对应芯片的发行包，在原目录执行：

```bash
chmod 0755 ./gocron
./gocron --version
./gocron db-upgrade
# 确认成功后再启动，或恢复原 launchd 服务。
./gocron web
```

macOS 补充包为交叉编译产物，未进行 Apple 签名、公证或 Mac 真机运行验证。系统拦截时先核验来源和 SHA-256，再按系统提示允许该程序运行，不建议全局关闭 Gatekeeper。现有一键安装脚本仅适用于 Linux/systemd，不适用于 macOS。

## 8. 回滚

升级后出现无法接受的问题时，先停止新版 Web 服务，确认已下发的重要任务状态，再恢复升级前的数据库备份、旧二进制及同一时间的 `conf` 和启动配置，最后启动旧服务并核验。

**不能只换回旧二进制或降低 `conf/.version`。** 新版可能已改变字段类型；旧程序连接迁移后的库不属于可靠回滚。恢复数据库会丢失备份后新增的数据，应先保存现场并确认恢复时间点。
