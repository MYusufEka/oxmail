# Rspamd normal worker configuration
# Handles message scanning
# Optimized for small team (< 50 users)

bind_socket = "0.0.0.0:11333";
count = 1;
task_timeout = 20s;
