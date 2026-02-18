#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

REPO_OWNER="bamboo-services"
REPO_NAME="bamboo-base-go-cli"
BINARY_PREFIX="bamboo-base-cli"
INSTALL_NAME="bamboo"

print_info() { echo -e "${CYAN}$1${NC}"; }
print_success() { echo -e "${GREEN}$1${NC}"; }
print_warning() { echo -e "${YELLOW}$1${NC}"; }
print_error() { echo -e "${RED}$1${NC}" >&2; }

detect_os() {
  case "$(uname -s)" in
    Linux*) echo "linux" ;;
    Darwin*) echo "darwin" ;;
    FreeBSD*) echo "freebsd" ;;
    *)
      print_error "不支持的操作系统: $(uname -s)"
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)
      print_error "不支持的架构: $(uname -m)"
      exit 1
      ;;
  esac
}

check_dependencies() {
  local missing=()
  for cmd in curl grep sed awk; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      missing+=("$cmd")
    fi
  done
  if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
    missing+=("sha256sum/shasum")
  fi
  if [ ${#missing[@]} -gt 0 ]; then
    print_error "缺少必要命令: ${missing[*]}"
    exit 1
  fi
}

download_with_retry() {
  local url="$1"
  local output="$2"
  local max_attempts=3
  local attempt=1
  while [ $attempt -le $max_attempts ]; do
    if curl -fsSL "$url" -o "$output"; then
      return 0
    fi
    print_warning "下载失败，重试 $attempt/$max_attempts"
    sleep 2
    attempt=$((attempt + 1))
  done
  return 1
}

hash_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi
  shasum -a 256 "$1" | awk '{print $1}'
}

get_latest_version() {
  local api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
  local raw
  raw=$(curl -fsSL "$api_url")
  local version
  version=$(echo "$raw" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/' | head -n 1)
  if [ -z "$version" ]; then
    print_error "无法获取最新版本"
    exit 1
  fi
  echo "$version"
}

main() {
  print_info "开始安装 bamboo-base-cli"
  check_dependencies

  local os
  local arch
  os=$(detect_os)
  arch=$(detect_arch)
  print_success "检测到系统: ${os}-${arch}"

  local version="${1:-latest}"
  if [ "$version" = "latest" ]; then
    version=$(get_latest_version)
  else
    version="${version#v}"
  fi
  print_success "目标版本: v${version}"

  local binary_name="${BINARY_PREFIX}-${os}-${arch}"
  local binary_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${version}/${binary_name}"
  local checksum_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${version}/checksums.txt"

  local tmp_dir
  tmp_dir=$(mktemp -d)
  trap 'rm -rf "$tmp_dir"' EXIT

  if ! download_with_retry "$binary_url" "$tmp_dir/$binary_name"; then
    print_error "下载失败: $binary_url"
    exit 1
  fi
  print_success "二进制下载完成"

  if download_with_retry "$checksum_url" "$tmp_dir/checksums.txt"; then
    local expected
    expected=$(grep "$binary_name" "$tmp_dir/checksums.txt" | awk '{print $1}' | head -n 1 || true)
    if [ -n "$expected" ]; then
      local actual
      actual=$(hash_file "$tmp_dir/$binary_name")
      if [ "$expected" != "$actual" ]; then
        print_error "校验失败"
        print_error "期望: $expected"
        print_error "实际: $actual"
        exit 1
      fi
      print_success "校验通过"
    else
      print_warning "未找到对应校验值，跳过校验"
    fi
  else
    print_warning "未下载到 checksums.txt，跳过校验"
  fi

  local install_dir="$HOME/.local/bin"
  mkdir -p "$install_dir"
  chmod +x "$tmp_dir/$binary_name"
  mv "$tmp_dir/$binary_name" "$install_dir/$INSTALL_NAME"
  print_success "安装完成: $install_dir/$INSTALL_NAME"

  if [[ ":$PATH:" != *":$install_dir:"* ]]; then
    print_warning "$install_dir 不在 PATH 中"
    print_info "可执行: export PATH=\"$HOME/.local/bin:$PATH\""
  else
    print_success "现在可运行: bamboo --help"
  fi
}

main "$@"
