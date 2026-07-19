#!/bin/sh
set -eu

default_rspamd_controller_password='$2$replace-with-rspamadm-bcrypt-hash'
rspamd_controller_password=${OXMAIL_RSPAMD_PASSWORD:-$default_rspamd_controller_password}

cat > /etc/rspamd/local.d/worker-controller.inc <<EOF
bind_socket = "0.0.0.0:11334";
count = 1;
secure_ip = "127.0.0.1";
password = "${rspamd_controller_password}";
static_dir = "\${WWWDIR}";
EOF

exec "$@"
