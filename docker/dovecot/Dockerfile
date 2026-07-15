FROM alpine:3.20

RUN apk add --no-cache \
    dovecot \
    dovecot-lmtpd \
    dovecot-pigeonhole-plugin \
    dovecot-pop3d \
    dovecot-fts-flatcurve \
    && (addgroup -g 5000 vmail 2>/dev/null || true) \
    && (adduser -S -G vmail -u 5000 vmail 2>/dev/null || true) \
    && (addgroup -S postfix 2>/dev/null || true) \
    && (adduser -S -G postfix -u 100 -D postfix 2>/dev/null || true) \
    && mkdir -p /var/mail/vhosts \
    && chown -R vmail:vmail /var/mail \
    && mkdir -p /run/dovecot \
    && chown dovecot:dovecot /run/dovecot \
    && mkdir -p /var/spool/postfix/private \
    && mkdir -p /etc/oxmail/dovecot/conf.d \
    && rm -rf /var/cache/apk/* /tmp/*

COPY healthcheck.sh /usr/local/bin/healthcheck.sh
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/healthcheck.sh /usr/local/bin/entrypoint.sh

EXPOSE 143 993 24

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
