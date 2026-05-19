FROM alpine:3.19

RUN apk add --no-cache \
    dovecot \
    dovecot-lmtpd \
    dovecot-pigeonhole-plugin \
    && (addgroup -S vmail 2>/dev/null || true) \
    && (adduser -S -G vmail -u 5000 vmail 2>/dev/null || true) \
    && mkdir -p /var/mail/vhosts \
    && chown -R vmail:vmail /var/mail \
    && mkdir -p /run/dovecot \
    && chown dovecot:dovecot /run/dovecot \
    && rm -rf /var/cache/apk/* /tmp/*

# Performance tuning for small team usage
RUN echo "auth_cache_size = 100" >> /etc/dovecot/dovecot.conf \
    && echo "auth_cache_ttl = 1 hour" >> /etc/dovecot/dovecot.conf \
    && echo "mail_max_userip_connections = 10" >> /etc/dovecot/dovecot.conf \
    && echo "login_max_processes_count = 16" >> /etc/dovecot/dovecot.conf \
    && echo "default_process_limit = 20" >> /etc/dovecot/dovecot.conf

COPY healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

EXPOSE 143 993 24

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

CMD ["dovecot", "-F"]
