#!/bin/sh
# Dovecot health check: verify process is running
doveadm process status 2>/dev/null || exit 1
exit 0
