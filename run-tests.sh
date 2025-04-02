#!/bin/bash

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Track overall status
STATUS=0
LOG_DIR="logs"

mkdir -p "$LOG_DIR"

function sanitize() {
  echo "$1" | tr ' /' '__'
}

function run_step() {
  local name="$1"
  local cmd="$2"
  local log_file="$LOG_DIR/$(sanitize "$name").log"

  echo -e "\n🔧 Running step: ${name}"
  echo "    ↪ Logging to: $log_file"

  if bash -c "$cmd" > "$log_file" 2>&1; then
    echo -e "${GREEN}✅ Step succeeded: ${name}${NC}"
  else
    echo -e "${RED}❌ Step failed: ${name}${NC} (see $log_file)"
    STATUS=1
  fi
}

# Run steps
run_step "Kubebuilder smoke and unit tests" "make test"
run_step "Helm lint" "make helm-lint"
run_step "Helm template" "make helm-template"
run_step "End-to-end tests" "make e2e"

# Final status
echo ""
if [ "$STATUS" -eq 0 ]; then
  echo -e "${GREEN}🎉 All tests passed!${NC}"
  exit 0
else
  echo -e "${RED}🔥 Some tests failed. Check logs in: $LOG_DIR${NC}"
  exit 1
fi