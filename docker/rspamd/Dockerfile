FROM alpine:3.19

RUN apk add --no-cache \
    rspamd \
    rspamd-controller \
    rspamd-proxy \
    curl \
    && mkdir -p /etc/rspamd/local.d \
    && mkdir -p /var/lib/rspamd \
    && chown -R rspamd:rspamd /var/lib/rspamd

COPY healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

EXPOSE 11333 11334

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

CMD ["rspamd", "-f", "-u", "rspamd", "-g", "rspamd"]
