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
    && mkdir -p /var/spool/postfix/private

COPY healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

EXPOSE 25 587

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

CMD ["postfix", "start-fg"]
