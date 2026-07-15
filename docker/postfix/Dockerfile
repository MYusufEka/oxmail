FROM alpine:3.19

RUN apk add --no-cache \
    postfix \
    postfix-pcre \
    ca-certificates \
    cyrus-sasl \
    openssl \
    && (addgroup -S vmail 2>/dev/null || true) \
    && (adduser -S -G vmail -u 5000 vmail 2>/dev/null || true) \
    && mkdir -p /var/mail/vhosts \
    && chown -R vmail:vmail /var/mail \
    && mkdir -p /var/spool/postfix/private \
    && mkdir -p /etc/oxmail/postfix \
    && chown -R postfix:postfix /etc/oxmail \
    && rm -rf /var/cache/apk/* /tmp/*

# Add Docker subnet to mynetworks (rendered main.cf has localhost only)
RUN postconf -e "mynetworks = 127.0.0.0/8 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16"

COPY healthcheck.sh /usr/local/bin/healthcheck.sh
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/healthcheck.sh /usr/local/bin/entrypoint.sh

EXPOSE 25 587

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
