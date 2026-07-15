FROM alpine:3.19

RUN apk add --no-cache \
    rspamd \
    rspamd-controller \
    rspamd-proxy \
    curl \
    && mkdir -p /etc/rspamd/local.d \
    && mkdir -p /etc/rspamd/override.d \
    && mkdir -p /var/lib/rspamd \
    && chown -R rspamd:rspamd /var/lib/rspamd \
    && rm -rf /var/cache/apk/* /tmp/*

# Performance tuning: reduce DNS requests, history size
RUN echo "dns_max_requests = 16;" > /etc/rspamd/local.d/options.inc \
    && echo "history_rows = 100;" >> /etc/rspamd/local.d/options.inc \
    && echo "dns_retransmits = 2;" >> /etc/rspamd/local.d/options.inc

# Disable heavy modules
RUN echo "enabled = false;" > /etc/rspamd/override.d/fuzzy_check.conf \
    && echo "enabled = false;" > /etc/rspamd/override.d/neural.conf \
    && echo "enabled = false;" > /etc/rspamd/override.d/chartable.conf

# Bind to all interfaces so other containers can connect
RUN echo 'bind_socket = "0.0.0.0:11333";' > /etc/rspamd/local.d/worker-normal.inc

COPY healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

EXPOSE 11333 11334

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

CMD ["rspamd", "-f", "-u", "rspamd", "-g", "rspamd"]
