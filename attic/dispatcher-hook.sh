#!/bin/sh
# RETIRED 2026-08-23 — replaced by the kernel netlink watcher (internal/watcher).
# Kept as protocol reference: the root NM dispatcher hook used to curl the
# serve unix socket on up/pre-down. Removed so install never needs pkexec.
#
# SOCK=<config dir>/serve.sock
# case "$2" in
#   up)       printf 'connect-current\n' | curl -s --unix-socket "$SOCK" http://localhost/hook & ;;
#   pre-down) printf 'disconnect\n'      | curl -s --unix-socket "$SOCK" http://localhost/hook & ;;
# esac
exit 0
