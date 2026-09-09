#!/usr/bin/env bash
# ==============================================================================
# Wavelet Agent Skills 离线拉取与更新工具
# 用法:
#   ./scripts/update_skills.sh <skill-name>
#   ./scripts/update_skills.sh --all-sp
#   ./scripts/update_skills.sh --all-ecc
#   ./scripts/update_skills.sh --all
#   ./scripts/update_skills.sh --list
# ==============================================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILLS_DIR="$ROOT_DIR/.agents/skills"

SP_REPO="https://github.com/obra/superpowers"
ECC_REPO="https://github.com/affaan-m/ECC"
DSH_REPO="https://github.com/czm15053/write-notes-like-deepseek"
AUTORESEARCH_REPO="https://github.com/dave1010/autoresearch"

SP_SKILLS=(
  "using-superpowers"
  "brainstorming"
  "systematic-debugging"
  "writing-plans"
  "executing-plans"
  "verification-before-completion"
  "receiving-code-review"
  "using-git-worktrees"
)

ECC_SKILLS=(
  "accessibility"
  "api-design"
  "clickhouse-io"
  "content-hash-cache-pattern"
  "data-throughput-accelerator"
  "deployment-patterns"
  "design-system"
  "docker-patterns"
  "e2e-testing"
  "frontend-patterns"
  "golang-patterns"
  "golang-testing"
  "hexagonal-architecture"
  "motion-patterns"
  "motion-ui"
  "nextjs-turbopack"
  "production-audit"
  "react-patterns"
  "react-performance"
  "react-testing"
  "security-bounty-hunter"
  "security-review"
  "shadcn"
  "tdd-workflow"
  "code-review-skill"
)

usage() {
  cat << USAGE
Wavelet Agent Skills 离线更新工具

用法:
  $0 <skill-name>        更新指定的单个 Skill
  $0 --all-sp            批量更新所有 8 个 Superpowers 流程治理技能
  $0 --all-ecc           批量更新所有 25 个 ECC 社区技能
  $0 --all               更新所有外部技能 (Superpowers + ECC + DeepSeek + Autoresearch)
  $0 --list              列出本地已安装技能及其上游来源
  $0 --help              显示本帮助信息

示例:
  $0 systematic-debugging
  $0 verification-before-completion
  $0 golang-patterns
  $0 write-notes-like-deepseek
USAGE
}

list_skills() {
  echo "=== Wavelet 已安装技能清单 ==="
  for d in "$SKILLS_DIR"/*; do
    if [ -d "$d" ]; then
      s="$(basename "$d")"
      if [[ " ${SP_SKILLS[*]} " =~ " ${s} " ]]; then
        printf "  %-32s -> %s\n" "$s" "[Superpowers] $SP_REPO"
      elif [[ " ${ECC_SKILLS[*]} " =~ " ${s} " ]]; then
        printf "  %-32s -> %s\n" "$s" "[ECC] $ECC_REPO"
      elif [ "$s" = "write-notes-like-deepseek" ]; then
        printf "  %-32s -> %s\n" "$s" "[DeepSeek Notes] $DSH_REPO"
      elif [ "$s" = "autoresearch" ]; then
        printf "  %-32s -> %s\n" "$s" "[Autoresearch] $AUTORESEARCH_REPO"
      else
        printf "  %-32s -> %s\n" "$s" "[Wavelet 自研] 本地核心技能"
      fi
    fi
  done
}

update_sp_skill() {
  local target="$1"
  local tmp_dir
  tmp_dir="$(mktemp -d -t sp_update_XXXXXX)"
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "==> 正在拉取 Superpowers 仓库 (浅克隆)..."
  git clone --depth 1 "$SP_REPO" "$tmp_dir/SP"

  if [ ! -d "$tmp_dir/SP/skills/$target" ]; then
    echo "❌ 错误: Superpowers 仓库中未找到技能 '$target'" >&2
    exit 1
  fi

  echo "==> 覆盖更新 $target ..."
  rm -rf "$SKILLS_DIR/$target"
  cp -R "$tmp_dir/SP/skills/$target" "$SKILLS_DIR/$target"
  echo "✓ 技能 '$target' 已成功更新至最新版本！"
}

update_all_sp() {
  local tmp_dir
  tmp_dir="$(mktemp -d -t sp_update_XXXXXX)"
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "==> 正在拉取 Superpowers 仓库最新代码..."
  git clone --depth 1 "$SP_REPO" "$tmp_dir/SP"

  echo "==> 批量更新 8 个 Superpowers 技能..."
  for s in "${SP_SKILLS[@]}"; do
    if [ -d "$tmp_dir/SP/skills/$s" ]; then
      rm -rf "$SKILLS_DIR/$s"
      cp -R "$tmp_dir/SP/skills/$s" "$SKILLS_DIR/$s"
      echo "  ✓ 已更新: $s"
    else
      echo "  ⚠ 警告: 上游未找到 $s"
    fi
  done
  echo "✓ 全部 Superpowers 流程技能更新完成！"
}

update_ecc_skill() {
  local target="$1"
  local tmp_dir
  tmp_dir="$(mktemp -d -t ecc_update_XXXXXX)"
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "==> 正在拉取 ECC 仓库 (浅克隆)..."
  git clone --depth 1 "$ECC_REPO" "$tmp_dir/ECC"

  if [ ! -d "$tmp_dir/ECC/skills/$target" ]; then
    echo "❌ 错误: ECC 仓库中未找到技能 '$target'" >&2
    exit 1
  fi

  echo "==> 覆盖更新 $target ..."
  rm -rf "$SKILLS_DIR/$target"
  cp -R "$tmp_dir/ECC/skills/$target" "$SKILLS_DIR/$target"
  echo "✓ 技能 '$target' 已成功更新至最新版本！"
}

update_all_ecc() {
  local tmp_dir
  tmp_dir="$(mktemp -d -t ecc_update_XXXXXX)"
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "==> 正在拉取 ECC 仓库最新代码..."
  git clone --depth 1 "$ECC_REPO" "$tmp_dir/ECC"

  echo "==> 批量更新 25 个 ECC 技能..."
  for s in "${ECC_SKILLS[@]}"; do
    if [ -d "$tmp_dir/ECC/skills/$s" ]; then
      rm -rf "$SKILLS_DIR/$s"
      cp -R "$tmp_dir/ECC/skills/$s" "$SKILLS_DIR/$s"
      echo "  ✓ 已更新: $s"
    else
      echo "  ⚠ 警告: 上游未找到 $s"
    fi
  done
  echo "✓ 全部 ECC 社区技能更新完成！"
}

update_deepseek_note() {
  local tmp_dir
  tmp_dir="$(mktemp -d -t dsh_update_XXXXXX)"
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "==> 正在拉取 write-notes-like-deepseek 最新代码..."
  git clone --depth 1 "$DSH_REPO" "$tmp_dir/dsh"

  local target="$SKILLS_DIR/write-notes-like-deepseek"
  rm -rf "$target"
  mkdir -p "$target"

  cp -R "$tmp_dir/dsh/SKILL.md" "$tmp_dir/dsh/README.md" "$tmp_dir/dsh/package.json" "$tmp_dir/dsh/board.html" \
        "$tmp_dir/dsh/references" "$tmp_dir/dsh/scripts" "$tmp_dir/dsh/templates" "$tmp_dir/dsh/assets" \
        "$target/"

  echo "✓ write-notes-like-deepseek 更新完成！"
}

update_autoresearch() {
  local tmp_dir
  tmp_dir="$(mktemp -d -t ar_update_XXXXXX)"
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "==> 正在拉取 autoresearch 最新代码..."
  git clone --depth 1 "$AUTORESEARCH_REPO" "$tmp_dir/ar"

  local target="$SKILLS_DIR/autoresearch"
  rm -rf "$target"
  mkdir -p "$target"

  cp -R "$tmp_dir/ar/SKILL.md" "$tmp_dir/ar"/*.md \
        "$tmp_dir/ar/references" "$tmp_dir/ar/scripts" "$tmp_dir/ar/agents" \
        "$target/" 2>/dev/null || true

  echo "✓ autoresearch 更新完成！"
}

main() {
  if [ $# -eq 0 ]; then
    usage
    exit 1
  fi

  case "$1" in
    --help|-h)
      usage
      ;;
    --list|-l)
      list_skills
      ;;
    --all-sp)
      update_all_sp
      ;;
    --all-ecc)
      update_all_ecc
      ;;
    --all)
      update_all_sp
      update_all_ecc
      update_deepseek_note
      update_autoresearch
      echo "==> 所有外部技能已更新完毕！"
      ;;
    write-notes-like-deepseek)
      update_deepseek_note
      ;;
    autoresearch)
      update_autoresearch
      ;;
    *)
      if [[ " ${SP_SKILLS[*]} " =~ " $1 " ]]; then
        update_sp_skill "$1"
      elif [[ " ${ECC_SKILLS[*]} " =~ " $1 " ]]; then
        update_ecc_skill "$1"
      else
        echo "❌ 技能 '$1' 不是外部管理的第三方技能，或不存在于已知技能列表中。"
        echo "运行 '$0 --list' 可查看所有已注册技能及其来源。"
        exit 1
      fi
      ;;
  esac
}

main "$@"
