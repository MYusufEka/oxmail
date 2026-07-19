# Rspamd controller worker configuration
# Provides HTTP API and web interface
# Minimal config for low resource usage

bind_socket = "0.0.0.0:11334";
count = 1;
secure_ip = "127.0.0.1";
static_dir = "${WWWDIR}";
