#!/usr/bin/env bash

set -Eeuo pipefail

REPOSITORY="caoyek/New-Gocron"
DEFAULT_INSTALL_DIR="/usr/local/gocron"
DEFAULT_WEB_PORT="5920"
DEFAULT_NODE_PORT="5921"
RUN_USER="new-gocron"
SERVICE_NAME="new-gocron.service"
NODE_SERVICE_NAME="new-gocron-node.service"
LEGACY_SERVICE_NAME="gocron.service"
LEGACY_NODE_SERVICE_NAME="gocron-node.service"
INSTALL_STATE_FILE="/etc/new-gocron/install.conf"
LOG_FILE="/var/log/new-gocron-install.log"

INSTALL_DIR="$DEFAULT_INSTALL_DIR"
WEB_PORT="$DEFAULT_WEB_PORT"
NODE_PORT="$DEFAULT_NODE_PORT"
VERSION=""
NODE_MODE="auto"
FORCE_UPGRADE=0
INSTALL_DIR_SET=0
CURRENT_STEP="初始化"
ERROR_REPORTED=0
TMP_DIR=""
BACKUP_DIR=""
EXISTING_INSTALL=0
CURRENT_VERSION_ID=""
ACTIVE_LEGACY_SERVICE=""
ACTIVE_LEGACY_NODE_SERVICE=""

usage() {
    cat <<'EOF'
New-Gocron Linux 在线安装程序

用法:
  sudo bash install.sh [选项]
  curl -fsSL https://raw.githubusercontent.com/caoyek/New-Gocron/main/scripts/install.sh | sudo bash

选项:
  --version VERSION       安装指定版本，例如 v2.0.2；默认安装 GitHub Latest
  --install-dir PATH      安装目录，默认 /usr/local/gocron
  --port PORT             Web 服务端口，默认 5920
  --node-port PORT        任务节点端口，默认 5921
  --with-node             同时安装或升级任务节点
  --without-node          不安装或升级任务节点
  --upgrade               要求当前目录已存在安装，再执行升级
  -h, --help              显示帮助

重复运行脚本会自动识别现有安装并升级，同时保留 conf、log 和 backups。
EOF
}

log() {
    printf '  %s\n' "$*"
}

ok() {
    printf '  [OK] %s\n' "$*"
}

info() {
    printf '  [->] %s\n' "$*"
}

step() {
    CURRENT_STEP="$2"
    printf '\n[%s/10] %s\n' "$1" "$2"
}

fail() {
    ERROR_REPORTED=1
    printf '\n  [ERROR] %s\n' "$*" >&2
    printf '  安装日志: %s\n' "$LOG_FILE" >&2
    exit 1
}

on_error() {
    local line="$1"
    if [[ "$ERROR_REPORTED" -eq 0 ]]; then
        printf '\n  [ERROR] %s失败，脚本停止于第 %s 行。\n' "$CURRENT_STEP" "$line" >&2
        printf '  安装日志: %s\n' "$LOG_FILE" >&2
    fi
}

cleanup() {
    if [[ -n "$TMP_DIR" && -d "$TMP_DIR" ]]; then
        rm -rf -- "$TMP_DIR"
    fi
}

trap 'on_error "$LINENO"' ERR
trap cleanup EXIT

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)
            [[ $# -ge 2 ]] || { printf '缺少 --version 参数\n' >&2; exit 1; }
            VERSION="$2"
            shift 2
            ;;
        --install-dir)
            [[ $# -ge 2 ]] || { printf '缺少 --install-dir 参数\n' >&2; exit 1; }
            INSTALL_DIR="$2"
            INSTALL_DIR_SET=1
            shift 2
            ;;
        --port)
            [[ $# -ge 2 ]] || { printf '缺少 --port 参数\n' >&2; exit 1; }
            WEB_PORT="$2"
            shift 2
            ;;
        --node-port)
            [[ $# -ge 2 ]] || { printf '缺少 --node-port 参数\n' >&2; exit 1; }
            NODE_PORT="$2"
            shift 2
            ;;
        --with-node)
            NODE_MODE="yes"
            shift
            ;;
        --without-node)
            NODE_MODE="no"
            shift
            ;;
        --upgrade)
            FORCE_UPGRADE=1
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            printf '未知参数: %s\n\n' "$1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

if [[ "$EUID" -ne 0 ]]; then
    printf '请使用 root 权限运行，例如: sudo bash install.sh\n' >&2
    exit 1
fi

install -m 0600 /dev/null "$LOG_FILE"
exec > >(tee -a "$LOG_FILE") 2>&1

printf '%s\n' '=================================================='
printf '%s\n' ' New-Gocron 在线安装程序'
printf '%s\n' '=================================================='
printf '开始时间: %s\n' "$(date '+%Y-%m-%d %H:%M:%S')"

read_state_value() {
    local key="$1"
    [[ -f "$INSTALL_STATE_FILE" ]] || return 0
    sed -n "s/^${key}=//p" "$INSTALL_STATE_FILE" | tail -n 1
}

detect_version_id() {
    local version_file="$INSTALL_DIR/conf/.version"
    local value output major minor patch
    if [[ -f "$version_file" ]]; then
        value="$(tr -d '[:space:]' < "$version_file")"
        [[ "$value" =~ ^[0-9]+$ ]] || fail "版本文件内容无效: $version_file"
        printf '%s' "$value"
        return
    fi

    output="$("$INSTALL_DIR/gocron" --version 2>/dev/null || true)"
    if [[ "$output" =~ ([0-9]+)\.([0-9]+)(\.([0-9]+))? ]]; then
        major="${BASH_REMATCH[1]}"
        minor="${BASH_REMATCH[2]}"
        patch="${BASH_REMATCH[4]:-0}"
        printf '%s%s%s' "$major" "$minor" "$patch"
    fi
}

service_uses_binary() {
    local service="$1"
    local binary="$2"
    systemctl cat "$service" 2>/dev/null | grep -Fq "$binary"
}

validate_port() {
    local value="$1"
    local label="$2"
    [[ "$value" =~ ^[0-9]+$ ]] || fail "$label 必须是数字"
    (( value >= 1 && value <= 65535 )) || fail "$label 必须在 1-65535 之间"
}

validate_install_dir() {
    [[ "$INSTALL_DIR" == /* ]] || fail "安装目录必须是绝对路径"
    [[ "$INSTALL_DIR" =~ ^[A-Za-z0-9._/-]+$ ]] || fail "安装目录只能包含字母、数字、点、下划线、横线和斜杠"
    [[ "/$INSTALL_DIR/" != *"/../"* ]] || fail "安装目录不能包含 .."
    case "$INSTALL_DIR" in
        /|/bin|/boot|/dev|/etc|/home|/lib|/lib64|/proc|/root|/run|/sbin|/sys|/tmp|/usr|/usr/local|/var)
            fail "不能使用系统关键目录作为安装目录: $INSTALL_DIR"
            ;;
    esac
    [[ ! -L "$INSTALL_DIR" ]] || fail "安装目录不能是符号链接: $INSTALL_DIR"
}

step 1 "检查运行环境"
[[ "$(uname -s)" == "Linux" ]] || fail "当前脚本仅支持 Linux"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    *)
        fail "当前版本仅提供 Linux amd64 安装包，检测到架构: $ARCH"
        ;;
esac
[[ -d /run/systemd/system ]] || fail "当前系统未使用 systemd"
for command_name in curl tar sha256sum awk sed grep install systemctl getent useradd groupadd; do
    command -v "$command_name" >/dev/null 2>&1 || fail "缺少必要命令: $command_name"
done
ok "操作系统: Linux"
ok "系统架构: $ARCH"
ok "权限检查: root"
ok "服务管理: systemd"

if [[ "$INSTALL_DIR_SET" -eq 0 ]]; then
    saved_install_dir="$(read_state_value INSTALL_DIR)"
    if [[ -n "$saved_install_dir" ]]; then
        INSTALL_DIR="$saved_install_dir"
    fi
fi
saved_web_port="$(read_state_value WEB_PORT)"
saved_node_port="$(read_state_value NODE_PORT)"
[[ -z "$saved_web_port" || "$WEB_PORT" != "$DEFAULT_WEB_PORT" ]] || WEB_PORT="$saved_web_port"
[[ -z "$saved_node_port" || "$NODE_PORT" != "$DEFAULT_NODE_PORT" ]] || NODE_PORT="$saved_node_port"
INSTALL_DIR="${INSTALL_DIR%/}"
validate_install_dir
validate_port "$WEB_PORT" "Web 端口"
validate_port "$NODE_PORT" "任务节点端口"

step 2 "获取发行版本"
if [[ -z "$VERSION" ]]; then
    latest_url="$(curl -fsSLI --retry 3 --connect-timeout 15 -o /dev/null -w '%{url_effective}' "https://github.com/${REPOSITORY}/releases/latest")"
    VERSION="${latest_url##*/}"
fi
[[ "$VERSION" == v* ]] || VERSION="v${VERSION}"
[[ "$VERSION" =~ ^v[0-9A-Za-z._-]+$ ]] || fail "版本号格式无效: $VERSION"
RELEASE_BASE_URL="https://github.com/${REPOSITORY}/releases/download/${VERSION}"
ok "GitHub 仓库: $REPOSITORY"
ok "目标版本: $VERSION"

step 3 "检查现有安装"
if [[ -f "$INSTALL_DIR/gocron" ]]; then
    EXISTING_INSTALL=1
    ok "检测到现有安装: $INSTALL_DIR"
    current_version="$($INSTALL_DIR/gocron --version 2>/dev/null || true)"
    [[ -n "$current_version" ]] && log "当前版本: $current_version"
    CURRENT_VERSION_ID="$(detect_version_id)"
    [[ -n "$CURRENT_VERSION_ID" ]] || fail "无法识别现有版本，请检查 $INSTALL_DIR/conf/.version"
    log "当前版本编号: $CURRENT_VERSION_ID"
    if systemctl is-active --quiet "$LEGACY_SERVICE_NAME" && service_uses_binary "$LEGACY_SERVICE_NAME" "$INSTALL_DIR/gocron"; then
        ACTIVE_LEGACY_SERVICE="$LEGACY_SERVICE_NAME"
        log "检测到旧版服务: $ACTIVE_LEGACY_SERVICE"
    fi
    if systemctl is-active --quiet "$LEGACY_NODE_SERVICE_NAME" && service_uses_binary "$LEGACY_NODE_SERVICE_NAME" "$INSTALL_DIR/gocron-node"; then
        ACTIVE_LEGACY_NODE_SERVICE="$LEGACY_NODE_SERVICE_NAME"
        log "检测到旧版节点服务: $ACTIVE_LEGACY_NODE_SERVICE"
    fi
else
    [[ "$FORCE_UPGRADE" -eq 0 ]] || fail "指定了 --upgrade，但未找到 $INSTALL_DIR/gocron"
    ok "将执行首次安装: $INSTALL_DIR"
fi
if [[ "$NODE_MODE" == "auto" ]]; then
    if [[ -f "$INSTALL_DIR/gocron-node" ]]; then
        NODE_MODE="yes"
    else
        NODE_MODE="no"
    fi
fi
if [[ "$NODE_MODE" == "yes" ]]; then
    log "任务节点: 安装或升级"
else
    log "任务节点: 本次不变更"
fi

step 4 "下载发行文件"
TMP_DIR="$(mktemp -d /tmp/new-gocron-install.XXXXXX)"
MAIN_PACKAGE="new-gocron-${VERSION}-linux-amd64.tar.gz"
NODE_PACKAGE="new-gocron-node-${VERSION}-linux-amd64.tar.gz"
curl -fL --retry 3 --connect-timeout 15 -o "$TMP_DIR/SHA256SUMS.txt" "$RELEASE_BASE_URL/SHA256SUMS.txt"
info "$MAIN_PACKAGE"
curl -fL --retry 3 --connect-timeout 15 -o "$TMP_DIR/$MAIN_PACKAGE" "$RELEASE_BASE_URL/$MAIN_PACKAGE"
if [[ "$NODE_MODE" == "yes" ]]; then
    info "$NODE_PACKAGE"
    curl -fL --retry 3 --connect-timeout 15 -o "$TMP_DIR/$NODE_PACKAGE" "$RELEASE_BASE_URL/$NODE_PACKAGE"
fi
ok "发行文件下载完成"

verify_package() {
    local package="$1"
    local expected actual
    expected="$(awk -v name="$package" '$2 == name {print $1}' "$TMP_DIR/SHA256SUMS.txt")"
    [[ -n "$expected" ]] || fail "校验文件中未找到 $package"
    actual="$(sha256sum "$TMP_DIR/$package" | awk '{print $1}')"
    [[ "$actual" == "$expected" ]] || fail "SHA-256 校验失败: $package"
    ok "SHA-256 校验通过: $package"
}

step 5 "校验发行文件"
verify_package "$MAIN_PACKAGE"
[[ "$NODE_MODE" != "yes" ]] || verify_package "$NODE_PACKAGE"
tar -tzf "$TMP_DIR/$MAIN_PACKAGE" | grep -qx 'gocron-linux-amd64/gocron' || fail "主程序压缩包结构无效"
if [[ "$NODE_MODE" == "yes" ]]; then
    tar -tzf "$TMP_DIR/$NODE_PACKAGE" | grep -qx 'gocron-node-linux-amd64/gocron-node' || fail "节点压缩包结构无效"
fi
mkdir -p "$TMP_DIR/extract-main"
tar -xzf "$TMP_DIR/$MAIN_PACKAGE" -C "$TMP_DIR/extract-main"
if [[ "$NODE_MODE" == "yes" ]]; then
    mkdir -p "$TMP_DIR/extract-node"
    tar -xzf "$TMP_DIR/$NODE_PACKAGE" -C "$TMP_DIR/extract-node"
fi
ok "压缩包结构检查通过"

step 6 "准备运行环境"
if ! getent group "$RUN_USER" >/dev/null 2>&1; then
    groupadd --system "$RUN_USER"
fi
if ! getent passwd "$RUN_USER" >/dev/null 2>&1; then
    useradd --system --gid "$RUN_USER" --home-dir "$INSTALL_DIR" --no-create-home --shell /usr/sbin/nologin "$RUN_USER"
fi
RUN_GROUP="$(id -gn "$RUN_USER")"
install -d -o root -g root -m 0755 "$INSTALL_DIR"
install -d -o "$RUN_USER" -g "$RUN_GROUP" -m 0750 "$INSTALL_DIR/conf" "$INSTALL_DIR/log"
install -d -o root -g root -m 0700 "$INSTALL_DIR/backups"
chown -R "$RUN_USER:$RUN_GROUP" "$INSTALL_DIR/conf" "$INSTALL_DIR/log"
ok "运行用户: $RUN_USER"
ok "安装目录: $INSTALL_DIR"

if [[ "$EXISTING_INSTALL" -eq 1 ]]; then
    backup_stamp="$(date '+%Y%m%d-%H%M%S')"
    BACKUP_DIR="$INSTALL_DIR/backups/$backup_stamp"
    install -d -o root -g root -m 0700 "$BACKUP_DIR"
    cp -a "$INSTALL_DIR/gocron" "$BACKUP_DIR/gocron"
    [[ ! -f "$INSTALL_DIR/gocron-node" ]] || cp -a "$INSTALL_DIR/gocron-node" "$BACKUP_DIR/gocron-node"
    [[ ! -d "$INSTALL_DIR/conf" ]] || cp -a "$INSTALL_DIR/conf" "$BACKUP_DIR/conf"
    ok "现有安装已备份: $BACKUP_DIR"
fi

step 7 "安装程序"
if systemctl is-active --quiet "$SERVICE_NAME"; then
    systemctl stop "$SERVICE_NAME"
fi
if [[ -n "$ACTIVE_LEGACY_SERVICE" ]]; then
    systemctl stop "$ACTIVE_LEGACY_SERVICE"
fi
if [[ "$NODE_MODE" == "yes" ]] && systemctl is-active --quiet "$NODE_SERVICE_NAME"; then
    systemctl stop "$NODE_SERVICE_NAME"
fi
if [[ "$NODE_MODE" == "yes" && -n "$ACTIVE_LEGACY_NODE_SERVICE" ]]; then
    systemctl stop "$ACTIVE_LEGACY_NODE_SERVICE"
fi
install -o root -g root -m 0755 "$TMP_DIR/extract-main/gocron-linux-amd64/gocron" "$INSTALL_DIR/gocron.new"
mv -f "$INSTALL_DIR/gocron.new" "$INSTALL_DIR/gocron"
if [[ "$NODE_MODE" == "yes" ]]; then
    install -o root -g root -m 0755 "$TMP_DIR/extract-node/gocron-node-linux-amd64/gocron-node" "$INSTALL_DIR/gocron-node.new"
    mv -f "$INSTALL_DIR/gocron-node.new" "$INSTALL_DIR/gocron-node"
fi
ok "主程序已安装: $INSTALL_DIR/gocron"
[[ "$NODE_MODE" != "yes" ]] || ok "任务节点已安装: $INSTALL_DIR/gocron-node"

write_web_service() {
    cat > "/etc/systemd/system/$SERVICE_NAME" <<EOF
[Unit]
Description=New-Gocron web service
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=$RUN_USER
Group=$RUN_GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/gocron web --host 0.0.0.0 --port $WEB_PORT --env prod
Restart=on-failure
RestartSec=5s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
    chmod 0644 "/etc/systemd/system/$SERVICE_NAME"
}

write_node_service() {
    cat > "/etc/systemd/system/$NODE_SERVICE_NAME" <<EOF
[Unit]
Description=New-Gocron task node
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=$RUN_USER
Group=$RUN_GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/gocron-node -s 0.0.0.0:$NODE_PORT
Restart=on-failure
RestartSec=5s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
    chmod 0644 "/etc/systemd/system/$NODE_SERVICE_NAME"
}

step 8 "配置系统服务"
write_web_service
if [[ "$NODE_MODE" == "yes" ]]; then
    write_node_service
fi
install -d -o root -g root -m 0755 "$(dirname "$INSTALL_STATE_FILE")"
cat > "$INSTALL_STATE_FILE" <<EOF
INSTALL_DIR=$INSTALL_DIR
WEB_PORT=$WEB_PORT
NODE_PORT=$NODE_PORT
VERSION=$VERSION
EOF
chmod 0600 "$INSTALL_STATE_FILE"
systemctl daemon-reload
systemctl enable "$SERVICE_NAME" >/dev/null
if [[ "$NODE_MODE" == "yes" ]]; then
    systemctl enable "$NODE_SERVICE_NAME" >/dev/null
fi
ok "已创建: $SERVICE_NAME"
[[ "$NODE_MODE" != "yes" ]] || ok "已创建: $NODE_SERVICE_NAME"
ok "已设置开机启动"

rollback_binaries() {
    [[ -n "$BACKUP_DIR" && -f "$BACKUP_DIR/gocron" ]] || return 0
    info "正在恢复升级前程序"
    cp -a "$BACKUP_DIR/gocron" "$INSTALL_DIR/gocron"
    if [[ -f "$BACKUP_DIR/gocron-node" ]]; then
        cp -a "$BACKUP_DIR/gocron-node" "$INSTALL_DIR/gocron-node"
    fi
    if [[ -f "$BACKUP_DIR/conf/.version" ]]; then
        cp -a "$BACKUP_DIR/conf/.version" "$INSTALL_DIR/conf/.version"
    else
        rm -f -- "$INSTALL_DIR/conf/.version"
    fi
    systemctl restart "$SERVICE_NAME" || true
    if [[ -f "$INSTALL_DIR/gocron-node" ]] && systemctl is-enabled --quiet "$NODE_SERVICE_NAME" 2>/dev/null; then
        systemctl restart "$NODE_SERVICE_NAME" || true
    fi
}

step 9 "升级数据库并启动服务"
if [[ -f "$INSTALL_DIR/conf/install.lock" ]]; then
    if [[ -n "$CURRENT_VERSION_ID" && "$CURRENT_VERSION_ID" -lt 200 ]]; then
        if [[ ! -f "$INSTALL_DIR/conf/.version" ]]; then
            printf '%s' "$CURRENT_VERSION_ID" > "$INSTALL_DIR/conf/.version"
            chown "$RUN_USER:$RUN_GROUP" "$INSTALL_DIR/conf/.version"
            chmod 0644 "$INSTALL_DIR/conf/.version"
        fi
        info "检测到 2.0 之前的版本，将在首次启动时按原生顺序升级数据库"
    else
        info "执行数据库结构检查"
        if ! "$INSTALL_DIR/gocron" db-upgrade; then
            rollback_binaries
            fail "数据库升级失败，已尝试恢复升级前程序"
        fi
        ok "数据库结构检查完成"
    fi
else
    log "首次安装尚未配置数据库，跳过数据库升级"
fi
if ! systemctl restart "$SERVICE_NAME"; then
    rollback_binaries
    fail "Web 服务启动失败"
fi
if [[ "$NODE_MODE" == "yes" ]]; then
    if ! systemctl restart "$NODE_SERVICE_NAME"; then
        rollback_binaries
        fail "任务节点启动失败"
    fi
fi
systemctl is-active --quiet "$SERVICE_NAME" || { rollback_binaries; fail "Web 服务未进入 active 状态"; }
health_ok=0
for ((attempt = 1; attempt <= 20; attempt++)); do
    if curl -fsS --connect-timeout 2 "http://127.0.0.1:$WEB_PORT/" >/dev/null; then
        health_ok=1
        break
    fi
    sleep 1
done
[[ "$health_ok" -eq 1 ]] || { rollback_binaries; fail "Web 服务健康检查失败"; }
if [[ -n "$ACTIVE_LEGACY_SERVICE" ]]; then
    systemctl disable "$ACTIVE_LEGACY_SERVICE" >/dev/null
fi
if [[ "$NODE_MODE" == "yes" && -n "$ACTIVE_LEGACY_NODE_SERVICE" ]]; then
    systemctl disable "$ACTIVE_LEGACY_NODE_SERVICE" >/dev/null
fi
ok "Web 服务状态: active"
ok "HTTP 健康检查通过: 127.0.0.1:$WEB_PORT"
if [[ "$NODE_MODE" == "yes" ]]; then
    systemctl is-active --quiet "$NODE_SERVICE_NAME" || { rollback_binaries; fail "任务节点未进入 active 状态"; }
    ok "任务节点状态: active"
fi

step 10 "安装完成"
server_ip="127.0.0.1"
if command -v hostname >/dev/null 2>&1; then
    detected_ip="$(hostname -I 2>/dev/null | awk '{print $1}')"
    [[ -z "$detected_ip" ]] || server_ip="$detected_ip"
fi
printf '%s\n' '=================================================='
printf ' 版本: %s\n' "$VERSION"
printf ' 目录: %s\n' "$INSTALL_DIR"
printf ' 服务: %s\n' "$SERVICE_NAME"
printf ' 地址: http://%s:%s\n' "$server_ip" "$WEB_PORT"
if [[ "$NODE_MODE" == "yes" ]]; then
    printf ' 节点: %s (端口 %s)\n' "$NODE_SERVICE_NAME" "$NODE_PORT"
fi
printf '%s\n' '=================================================='
if [[ ! -f "$INSTALL_DIR/conf/install.lock" ]]; then
    printf '请打开上述地址，完成数据库和管理员配置。\n'
fi
printf '安装日志: %s\n' "$LOG_FILE"
printf '服务日志: journalctl -u %s -f\n' "$SERVICE_NAME"
