# New-Gocron

New-Gocron 是一个定时任务管理系统，当前版本为 2.0.1。项目提供 Web 管理端和任务节点，可集中管理 Shell、HTTP 等定时任务。

## New-Gocron 2.0.1 主要改动

### 数据看板

- 展示启用任务数、停用任务数、今日执行成功数和今日执行失败数。
- 按 1-24 时展示今日任务执行趋势。
- 展示即将执行的任务和最近执行异常，并区分 Shell、HTTP 执行方式。

### 管理界面

- 使用 New-Gocron 品牌名称、图标、浏览器图标和站点清单。
- 重构登录页、左侧导航以及任务、日志、节点、用户和推送设置页面。
- 统一紧凑型页面布局，优化表格列宽、滚动区域、筛选栏和操作按钮。
- 新增、编辑、查看任务统一使用弹窗界面。
- 命令及执行结果使用深色代码框展示，长命令支持悬浮预览。

### 任务管理

- 任务 ID 和任务名称合并为一个模糊搜索条件。
- 标签、执行方式、任务节点和状态支持选择后自动查询。
- 标签可选择已有数据，也可在编辑任务时直接创建。
- 子任务改为从现有任务中多选，并在列表中显示对应主任务。
- 任务列表增加上次执行结果、执行时间和耗时。
- 执行、编辑、保存等操作完成后自动刷新列表数据。
- Cron 表达式提供每天、每周和每月快捷设置。
- HTTP 任务超时由任务配置控制，支持 `0–86400` 秒；`0` 表示不限制，新增 HTTP 任务默认为 `3600` 秒。
- 任务通知支持包含、不包含、通配符、正则表达式和数值比较规则，并可选择满足任意或全部规则。

### 后端与数据库

- 移除任务命令原有的 256 字符存储限制。
- 修复异常数据库中任务删除字段类型不正确导致的删除失败。
- 增加 New-Gocron 2.0 数据库自动迁移和独立升级命令。
- 增加数据看板统计接口和任务列表最近执行信息。

### 登录安全

- 同时按来源 IP 和登录账号统计失败次数，统计周期、触发次数和封禁时长可在后台配置。
- 记录登录成功、密码错误、封禁拒绝和白名单拒绝事件。
- 支持后台 IP 白名单，可填写单个 IP 或 CIDR 网段。
- 在“系统 -> 登录安全”中管理策略、查看当前封禁并手动解封。

## 二进制部署

从 [Releases](https://github.com/caoyek/New-Gocron/releases) 下载对应系统的程序包并解压。运行目录至少需要包含以下内容：

```text
gocron
gocron-node
conf/
```

Windows 程序文件名为 `gocron.exe` 和 `gocron-node.exe`。

### 启动 Web 服务

Linux：

```bash
chmod +x gocron
./gocron web
```

Windows：

```powershell
.\gocron.exe web
```

Web 服务默认监听 `0.0.0.0:5920`。可按需指定地址、端口和运行环境：

```bash
./gocron web --host 0.0.0.0 --port 5920 --env prod
```

首次部署时访问 `http://服务器IP:5920`，根据安装页面完成数据库配置。安装后的配置保存在 `conf/app.ini`。

### 启动任务节点

Shell 任务需要至少一个可用的任务节点。请在实际执行命令的服务器上启动 `gocron-node`。

Linux：

```bash
chmod +x gocron-node
./gocron-node
```

Windows：

```powershell
.\gocron-node.exe
```

任务节点默认监听 `0.0.0.0:5921`，指定监听地址可使用：

```bash
./gocron-node -s 0.0.0.0:5921
```

启动后在 New-Gocron 的“任务节点”页面添加节点地址。

## 从旧版本升级

New-Gocron 2.0 相对原项目涉及以下数据库字段调整：

| 表 | 字段 | 原类型或异常类型 | New-Gocron 2.0 类型 |
| --- | --- | --- | --- |
| `{prefix}task` | `command` | `VARCHAR(256)` | `TEXT NOT NULL` |
| `{prefix}task_log` | `command` | `VARCHAR(256)` | `TEXT NOT NULL` |
| `{prefix}task` | `notify_keyword` | `VARCHAR(128)` | `TEXT NOT NULL` |
| `{prefix}task` | `deleted` | 部分异常数据库为数值类型 | MySQL `DATETIME NULL` / PostgreSQL `TIMESTAMP NULL` |

登录安全功能还会新增两张表：

| 表 | 用途 |
| --- | --- |
| `{prefix}login_security_event` | 保存所有登录成功和失败事件 |
| `{prefix}login_block` | 保存 IP 和账号的封禁状态 |

登录安全策略保存在已有的 `{prefix}setting` 表中，`code` 为 `login_security`，`key` 为 `policy`。

其中 `{prefix}` 是 `conf/app.ini` 中 `db.prefix` 配置的表前缀。前缀会直接与表名拼接，例如 `db.prefix = gocron` 对应的任务表为 `gocrontask`。

推荐按以下顺序升级生产环境：

1. 备份当前数据库和 `conf/app.ini`。
2. 停止旧版 Web 服务，避免升级过程中继续写入数据。
3. 解压 New-Gocron 2.0.1，并保留原有的 `conf/app.ini`、`conf/install.lock` 和 `conf/.version`。
4. 在 New-Gocron 程序目录执行数据库升级命令。
5. 数据库升级成功后再启动 Web 服务和任务节点。

Linux：

```bash
./gocron db-upgrade
./gocron web
```

Windows：

```powershell
.\gocron.exe db-upgrade
.\gocron.exe web
```

`db-upgrade` 会读取程序目录中的 `conf/app.ini`，执行 New-Gocron 2.0 所需的字段调整并创建登录安全数据表，不启动 Web 服务或任务调度。该命令可以重复执行。

正常从旧版本首次启动 New-Gocron 2.0 时也会进入版本迁移流程，但生产环境建议先单独执行 `db-upgrade`，确认数据库升级成功后再启动服务。

> 不要在未备份数据库的情况下直接升级，也不要让旧版 Web 服务长期连接已经完成 2.0 字段调整的数据库。

## 源码开发

### 环境准备

- Go（项目使用 Go Modules）
- Node.js 16.20.2（仓库提供 `.node-version`，可使用 fnm 切换）
- npm 或 Yarn
- MySQL 或 PostgreSQL

克隆项目：

```bash
git clone https://github.com/caoyek/New-Gocron.git
cd New-Gocron
```

使用 fnm 准备前端环境：

```bash
fnm install
fnm use
```

### 后端和任务节点

编译两个程序：

```bash
make
```

编译结果位于 `bin/`：

```text
bin/gocron
bin/gocron-node
```

编译并以开发模式启动后端和任务节点：

```bash
make run
```

也可以分别启动：

```bash
./bin/gocron web --env dev
./bin/gocron-node
```

### 前端

安装依赖并启动开发服务器：

```bash
cd web/vue
npm install
npm run dev
```

前端开发地址为 `http://localhost:8080`，`/api` 请求会代理到本地后端 `http://localhost:5920`。

也可以在项目根目录使用 Makefile：

```bash
make install-vue
make run-vue
```

### 测试与打包

运行 Go 测试：

```bash
go test ./...
```

在 Windows 本地生成 Linux、Windows amd64 完整发布包：

```powershell
.\scripts\package-release.ps1 -Version v2.0.2
```

Linux 或 macOS 可使用 Make：

```bash
make package VERSION=v2.0.2
```

两个入口都会构建前端、更新内嵌静态资源、运行 Go 测试，并在 `dist/` 生成 Windows/Linux 主程序、节点程序和 `SHA256SUMS.txt`。本地构建只生成文件，不会自动上传 GitHub。

### GitHub Actions 自动构建

- 推送 `main`：自动测试并生成 Actions Artifact，不创建 Release。
- 在 Actions 页面手动运行：生成测试 Artifact，不创建 Release。
- 推送 `v*` 标签：自动测试、编译并创建或更新对应的 GitHub Release。

正式发布示例：

```bash
git push origin main
git tag v2.0.2
git push origin v2.0.2
```

需要中文发行说明时，在推送标签前新增 `docs/releases/v2.0.2.md`。工作流会把该文件放在 Release 正文顶部，并在下方追加 GitHub 根据提交和 PR 自动生成的变更记录；未提供该文件时则只使用自动生成内容。

本地和 GitHub 可以同时编译，但正式 Release 的同名资产建议只由 GitHub Actions 上传，避免并发覆盖。

## 常用端口

| 服务 | 默认地址 | 用途 |
| --- | --- | --- |
| Web 服务 | `0.0.0.0:5920` | 管理页面和 API |
| 任务节点 | `0.0.0.0:5921` | 接收并执行 Shell 任务 |
| 前端开发服务 | `localhost:8080` | Vue 开发与热更新 |

## 配置与数据安全

- 主配置文件为 `conf/app.ini`，不要将生产数据库密码提交到 Git。
- 升级或迁移前必须备份数据库和配置文件。
- 从线上数据库复制数据到开发环境后，应先停用所有任务，避免开发环境误执行线上任务。
- 生产环境建议使用进程管理工具托管 `gocron web` 和 `gocron-node`，并限制管理端及任务节点端口的访问范围。
- 如果 Web 服务经过 Nginx、Caddy 等反向代理，应在代理或防火墙层同时配置后台 IP 白名单；应用层默认使用 TCP 对端 IP，不信任可伪造的请求头。

## 问题反馈

请在 [New-Gocron Issues](https://github.com/caoyek/New-Gocron/issues) 提交问题，并附上版本、操作系统、数据库类型和相关日志。

## 项目来源与许可证

New-Gocron 基于 [ouqiang/gocron](https://github.com/ouqiang/gocron) 二次开发。原项目的功能说明、架构资料和历史版本记录请查看原项目仓库，本项目不再重复收录。

本项目继续保留原项目的开源许可证和版权声明，具体内容见 [LICENSE](LICENSE)。
