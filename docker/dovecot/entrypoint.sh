#!/bin/sh
# Oxmail Dovecot entrypoint
# Ensures rendered config exists before starting Dovecot

CONF_DIR="/etc/oxmail/dovecot"
OXMAIL_DOMAIN=${OXMAIL_DOMAIN:-local.test}
CERT_COMMON_NAME="mail.$OXMAIL_DOMAIN"
LETSENCRYPT_CERT_FILE="/etc/letsencrypt/acme.json.d/certificates/$OXMAIL_DOMAIN.crt"
FALLBACK_CERT_DIR="/etc/ssl/dovecot/$CERT_COMMON_NAME"
FALLBACK_CERT_FILE="$FALLBACK_CERT_DIR/server.crt"
FALLBACK_KEY_FILE="$FALLBACK_CERT_DIR/server.key"

# Ensure sieve directories exist (always, regardless of config state)
mkdir -p /var/lib/sieve/scripts /var/lib/sieve/global
chown -R vmail:vmail /var/lib/sieve 2>/dev/null || true

# If rendered dovecot.conf doesn't exist yet, write minimal defaults
if [ ! -s "$CONF_DIR/dovecot.conf" ]; then
  mkdir -p "$CONF_DIR/conf.d"
  cat > "$CONF_DIR/dovecot.conf" << 'CONF'
protocols = imap lmtp pop3
log_path = /dev/stderr
info_log_path = /dev/stdout
mail_location = maildir:/var/mail/vhosts/%d/%n
ssl = no
disable_plaintext_auth = no
auth_mechanisms = plain login
!include conf.d/*.conf
CONF
fi

# Ensure 99-oxmail-auth.conf exists (passdb/userdb + LMTP/SASL sockets)
if [ ! -s "$CONF_DIR/conf.d/99-oxmail-auth.conf" ]; then
  cat > "$CONF_DIR/conf.d/99-oxmail-auth.conf" << 'CONF'
passdb {
  driver = passwd-file
  args = scheme=BLF-CRYPT /etc/oxmail/dovecot/passdb
}
userdb {
  driver = passwd-file
  args = /etc/oxmail/dovecot/userdb
  default_fields = uid=5000 gid=5000 home=/var/mail/vhosts/%d/%n
}
service lmtp {
  unix_listener /var/spool/postfix/private/dovecot-lmtp {
    mode = 0600
    user = postfix
  }
}
service auth {
  unix_listener /var/spool/postfix/private/auth {
    mode = 0600
    user = postfix
  }
}
ssl = no
disable_plaintext_auth = no
CONF
fi

if [ -s "$CONF_DIR/dovecot.conf" ]; then
  if grep -q 'ssl_cert = </etc/letsencrypt' "$CONF_DIR/dovecot.conf" 2>/dev/null; then
    if [ ! -f "$LETSENCRYPT_CERT_FILE" ]; then
      if [ ! -f "$FALLBACK_CERT_FILE" ] || [ ! -f "$FALLBACK_KEY_FILE" ]; then
        mkdir -p "$FALLBACK_CERT_DIR"
        openssl req -new -x509 -nodes -days 3650 \
          -newkey rsa:2048 \
          -keyout "$FALLBACK_KEY_FILE" \
          -out "$FALLBACK_CERT_FILE" \
          -subj "/CN=$CERT_COMMON_NAME/O=Oxmail Dev/C=US" \
          2>/dev/null
      fi
      sed -i 's/^ssl = no/ssl = yes/' "$CONF_DIR/dovecot.conf"
      sed -i '/^ssl_cert =/d' "$CONF_DIR/dovecot.conf"
      sed -i '/^ssl_key =/d' "$CONF_DIR/dovecot.conf"
      sed -i '/^ssl_min_protocol =/d' "$CONF_DIR/dovecot.conf"
      sed -i '/^ssl_prefer_server_ciphers =/d' "$CONF_DIR/dovecot.conf"
      cat >> "$CONF_DIR/dovecot.conf" << CONF
ssl_cert = <$FALLBACK_CERT_FILE
ssl_key = <$FALLBACK_KEY_FILE
CONF
      echo "entrypoint: SSL fallback cert generated for $CERT_COMMON_NAME"
    fi
  fi
fi

# Ensure pop3 is in protocols line (API binary may have old embedded template)
if [ -s "$CONF_DIR/dovecot.conf" ]; then
  if ! grep -q 'pop3' "$CONF_DIR/dovecot.conf" 2>/dev/null; then
    sed -i 's/^protocols = imap lmtp$/protocols = imap lmtp pop3/' "$CONF_DIR/dovecot.conf"
    echo "entrypoint: pop3 added to protocols"
  fi
fi

if [ -s "$CONF_DIR/dovecot.conf" ]; then
  if ! grep -q 'service pop3-login' "$CONF_DIR/dovecot.conf" 2>/dev/null; then
    cat >> "$CONF_DIR/dovecot.conf" << 'CONF'

# POP3 service added by entrypoint
service pop3-login {
  inet_listener pop3 {
    port = 110
  }
  inet_listener pop3s {
    port = 995
    ssl = yes
  }
}
CONF
    echo "entrypoint: POP3 login listener added"
  fi
fi

exec dovecot -c "$CONF_DIR/dovecot.conf" -F
