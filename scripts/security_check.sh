#!/usr/bin/env bash
# Copyright 2026 Arctel.net
# SPDX-License-Identifier: Apache-2.0
#
# security_check.sh
# Scan the working tree for committed/committable secrets with gitleaks.
#
# Note: 默认 gitleaks telegram 规则要求附近出现 telegr 标识符，会漏掉裸 Bot Token —
# 见 .agents/notes/implemented/process/2026-09-15-gitleaks-code-check.md

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_PATH="${ROOT_DIR}/.gitleaks.toml"
GITLEAKS_VERSION="8.30.1"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}==>${NC} ${BOLD}$1${NC}" >&2
}

die() {
    echo -e "${RED}error:${NC} $1" >&2
    exit 1
}

os_arch_asset() {
    local os arch
    case "$(uname -s)" in
        Darwin) os="darwin" ;;
        Linux) os="linux" ;;
        *) die "unsupported OS $(uname -s); install gitleaks v${GITLEAKS_VERSION} manually" ;;
    esac
    case "$(uname -m)" in
        arm64|aarch64) arch="arm64" ;;
        x86_64|amd64) arch="x64" ;;
        *) die "unsupported arch $(uname -m); install gitleaks v${GITLEAKS_VERSION} manually" ;;
    esac
    echo "gitleaks_${GITLEAKS_VERSION}_${os}_${arch}.tar.gz"
}

install_gitleaks() {
    local cache_dir="${XDG_CACHE_HOME:-${HOME}/.cache}/wavelet/gitleaks/${GITLEAKS_VERSION}"
    local bin_path="${cache_dir}/gitleaks"
    if [[ -x "${bin_path}" ]]; then
        echo "${bin_path}"
        return
    fi

    command -v curl >/dev/null 2>&1 || die "gitleaks is not installed and curl is missing; install gitleaks v${GITLEAKS_VERSION}"

    local asset url tmp
    asset="$(os_arch_asset)"
    url="https://github.com/gitleaks/gitleaks/releases/download/v${GITLEAKS_VERSION}/${asset}"
    tmp="$(mktemp -d)"
    log_info "Installing gitleaks v${GITLEAKS_VERSION} to ${cache_dir}"
    curl -fsSL "${url}" -o "${tmp}/${asset}"
    mkdir -p "${cache_dir}"
    tar -xzf "${tmp}/${asset}" -C "${cache_dir}" gitleaks
    rm -rf "${tmp}"
    chmod +x "${bin_path}"
    echo "${bin_path}"
}

resolve_gitleaks() {
    if command -v gitleaks >/dev/null 2>&1; then
        command -v gitleaks
        return
    fi
    install_gitleaks
}

run_gitleaks_dir() {
    local bin="$1"
    shift
    if "${bin}" dir --help >/dev/null 2>&1; then
        "${bin}" dir "$@"
        return
    fi
    # gitleaks < 8.24 used `detect --no-git` instead of `dir`.
    local source="."
    local forwarded=()
    while (($#)); do
        case "$1" in
            --no-banner|--redact|--verbose|-v)
                forwarded+=("$1")
                shift
                ;;
            --config|-c|--exit-code|--max-target-megabytes|--report-format|--report-path)
                forwarded+=("$1" "$2")
                shift 2
                ;;
            --*)
                forwarded+=("$1")
                shift
                ;;
            *)
                source="$1"
                shift
                ;;
        esac
    done
    "${bin}" detect --no-git --source "${source}" "${forwarded[@]}"
}

[[ -f "${CONFIG_PATH}" ]] || die "missing ${CONFIG_PATH}"

GITLEAKS_BIN="$(resolve_gitleaks)"

echo -e "${BOLD}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}     Secret scan (gitleaks)                                      ${NC}"
echo -e "${BOLD}═══════════════════════════════════════════════════════════════${NC}"
log_info "using ${GITLEAKS_BIN} ($("${GITLEAKS_BIN}" version 2>/dev/null || echo unknown))"

# Prove the extra Telegram rule still fires on a generic token assignment
# (the bundled telegram rule requires a nearby "telegr" identifier and misses this).
# Assemble the canary at runtime so this script never contains a complete token.
CANARY_DIR="$(mktemp -d "${TMPDIR:-/tmp}/wavelet-gitleaks-canary.XXXXXX")"
cleanup() { rm -rf "${CANARY_DIR}"; }
trap cleanup EXIT
canary_bot_id="1234567890"
canary_secret="A$(printf 'A%.0s' {1..34})"
printf 'credentials: { token: "%s:%s" }\n' "${canary_bot_id}" "${canary_secret}" > "${CANARY_DIR}/canary.yaml"

set +e
run_gitleaks_dir "${GITLEAKS_BIN}" \
    --no-banner \
    --config "${CONFIG_PATH}" \
    --exit-code 1 \
    "${CANARY_DIR}" >/dev/null 2>&1
canary_status=$?
set -e
if [[ "${canary_status}" -ne 1 ]]; then
    die "gitleaks self-check failed: did not detect a Telegram Bot API token assigned to a generic token field (exit ${canary_status})"
fi
echo -e "  ${GREEN}✓${NC} detector self-check passed (Telegram Bot API token rule)"

log_info "scanning ${ROOT_DIR}"
set +e
run_gitleaks_dir "${GITLEAKS_BIN}" \
    --no-banner \
    --redact \
    --verbose \
    --config "${CONFIG_PATH}" \
    --max-target-megabytes 8 \
    --exit-code 1 \
    "${ROOT_DIR}"
scan_status=$?
set -e

echo -e "${BOLD}═══════════════════════════════════════════════════════════════${NC}"
if [[ "${scan_status}" -eq 0 ]]; then
    echo -e "${GREEN}${BOLD}✓ gitleaks: no secrets found${NC}"
    exit 0
fi

echo -e "${RED}${BOLD}✗ gitleaks: secrets detected. Remove real credentials from source, tests, and committed config.${NC}" >&2
echo -e "${RED}  Tests must use obvious fakes (e.g. mock_token). Mark a documented false positive with gitleaks:allow.${NC}" >&2
exit "${scan_status}"
