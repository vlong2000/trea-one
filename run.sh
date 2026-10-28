#!/bin/bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
BIN="${ROOT}/bin/nrf"
CFG="${ROOT}/config/nrfcfg.yaml"

if [ ! -x "${BIN}" ]; then
	make -C "${ROOT}" nrf
fi

exec "${BIN}" -c "${CFG}"
