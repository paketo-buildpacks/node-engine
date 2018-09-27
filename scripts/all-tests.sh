#!/usr/bin/env bash
cd "$( dirname "${BASH_SOURCE[0]}" )/.."
./scripts/unit.sh && ./scripts/integration.sh && ./scripts/acceptance.sh

