FROM alpine:3.19

RUN apk add --no-cache \
    postfix \
    postfix-pcre \
    ca-certificates \
    cyrus-sasl \
    && (addgroup -S vmail 2>/dev/null || true) \
    && (adduser -S -G vmail -u 5000 vmail 2>/dev/null || true) \
    && mkdir -p /var/mail/vhosts \
    && chown -R vmail:vmail /var/mail \
    && mkdir -p /var/spool/postfix/private \
    && rm -rf /var/cache/apk/* /tmp/*

# Performance tuning: limit processes for small team usage
RUN postconf -e "default_process_limit = 10" \
    && postconf -e "smtpd_client_connection_count_limit = 10" \
    && postconf -e "smtpd_client_connection_rate_limit = 30" \
    && postconf -e "maximal_queue_lifetime = 1d" \
    && postconf -e "bounce_queue_lifetime = 1d" \
    && postconf -e "header_checks =" \
    && postconf -e "smtp_tls_session_cache_database =" \
    && postconf -e "smtpd_tls_session_cache_database ="

COPY healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

EXPOSE 25 587

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

CMD ["postfix", "start-fg"]
