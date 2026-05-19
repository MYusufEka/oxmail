#!/bin/sh
# Rspamd health check: ping the normal worker
curl -sf http://localhost:11333/ping >/dev/null 2>&1
exit $?
