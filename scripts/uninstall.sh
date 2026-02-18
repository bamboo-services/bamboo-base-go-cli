#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

print_info() { echo -e "${CYAN}$1${NC}"; }
print_success() { echo -e "${GREEN}$1${NC}"; }
print_warning() { echo -e "${YELLOW}$1${NC}"; }
print_error() { echo -e "${RED}$1${NC}" >&2; }

confirm() {
  local prompt="$1"
  read -r -p "$(echo -e "${CYAN}${prompt} [y/N] ${NC}")" answer
  case "$answer" in
    [yY]|[yY][eE][sS]) return 0 ;;
    *) return 1 ;;
  esac
}

main() {
  local install_path="$HOME/.local/bin/bamboo"

  if [ ! -f "$install_path" ]; then
    print_warning "未在默认位置找到 bamboo: $install_path"
    if command -v bamboo >/dev/null 2>&1; then
      local found
      found=$(command -v bamboo)
      print_warning "PATH 中检测到 bamboo: $found"
      if confirm "是否删除该路径的 bamboo？"; then
        install_path="$found"
      else
        print_info "已取消"
        exit 0
      fi
    else
      print_info "未检测到已安装的 bamboo"
      exit 0
    fi
  fi

  print_info "将删除: $install_path"
  if ! confirm "确认卸载 bamboo 吗？"; then
    print_info "已取消"
    exit 0
  fi

  rm -f "$install_path"
  print_success "卸载完成"

  if command -v bamboo >/dev/null 2>&1; then
    print_warning "PATH 中仍存在 bamboo: $(command -v bamboo)"
  fi
}

main "$@"
