#!/bin/sh
# Postfix health check: verify master process is running
postfix status 2>/dev/null
exit $?
