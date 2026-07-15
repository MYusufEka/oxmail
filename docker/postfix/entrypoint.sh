#!/bin/sh
# Oxmail Postfix entrypoint
# Ensures required config files exist before starting Postfix with rendered configs

mkdir -p /etc/oxmail/postfix
touch /etc/oxmail/postfix/virtual_domains
touch /etc/oxmail/postfix/virtual_aliases
touch /etc/oxmail/postfix/virtual_mailboxes

CERT_FILE="/etc/ssl/postfix/server.crt"
KEY_FILE="/etc/ssl/postfix/server.key"
if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
  mkdir -p /etc/ssl/postfix
  openssl req -new -x509 -nodes -days 3650 \
    -newkey rsa:2048 \
    -keyout "$KEY_FILE" \
    -out "$CERT_FILE" \
    -subj "/CN=mail.local.test/O=Oxmail Dev/C=US" \
    2>/dev/null
fi

# If rendered main.cf hasn't been created yet (API hasn't started), write defaults so
# Postfix can start immediately. The API will overwrite these once it starts.
if [ ! -s /etc/oxmail/postfix/main.cf ]; then
  cat > /etc/oxmail/postfix/main.cf << 'CONF'
myhostname = mail.local.test
mydomain = local.test
myorigin = $mydomain
mydestination = localhost
inet_interfaces = all
inet_protocols = ipv4
mynetworks = 127.0.0.0/8 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16
virtual_mailbox_domains = /etc/oxmail/postfix/virtual_domains
virtual_mailbox_maps = texthash:/etc/oxmail/postfix/virtual_mailboxes
virtual_alias_maps = texthash:/etc/oxmail/postfix/virtual_aliases
virtual_mailbox_base = /var/mail/vhosts
virtual_minimum_uid = 5000
virtual_uid_maps = static:5000
virtual_gid_maps = static:5000
virtual_transport = lmtp:unix:/var/spool/postfix/private/dovecot-lmtp
smtpd_sasl_type = dovecot
smtpd_sasl_path = private/auth
smtpd_sasl_auth_enable = yes
smtpd_sasl_security_options = noanonymous
smtpd_reject_unlisted_recipient = no
smtpd_relay_restrictions = permit_mynetworks, permit_sasl_authenticated, reject_unauth_destination
smtpd_recipient_restrictions = permit_mynetworks, permit_sasl_authenticated, reject_unauth_destination
maillog_file = /dev/stdout
message_size_limit = 10240000
compatibility_level = 3.6
CONF
fi

if [ ! -s /etc/oxmail/postfix/master.cf ]; then
  cat > /etc/oxmail/postfix/master.cf << 'CONF'
smtp      inet  n       -       n       -       -       smtpd
submission inet  n       -       n       -       -       smtpd
  -o syslog_name=postfix/submission
  -o smtpd_tls_security_level=encrypt
  -o smtpd_sasl_auth_enable=yes
  -o smtpd_sasl_type=dovecot
  -o smtpd_sasl_path=private/auth
  -o smtpd_recipient_restrictions=permit_sasl_authenticated,reject
  -o smtpd_relay_restrictions=permit_sasl_authenticated,reject
  -o milter_macro_daemon_name=ORIGINATING
smtps     inet  n       -       n       -       -       smtpd
  -o syslog_name=postfix/smtps
  -o smtpd_tls_wrappermode=yes
  -o smtpd_sasl_auth_enable=yes
  -o smtpd_sasl_type=dovecot
  -o smtpd_sasl_path=private/auth
  -o smtpd_recipient_restrictions=permit_sasl_authenticated,reject
  -o smtpd_relay_restrictions=permit_sasl_authenticated,reject
  -o milter_macro_daemon_name=ORIGINATING
pickup    unix  n       -       n       60      1       pickup
cleanup   unix  n       -       n       -       0       cleanup
qmgr      unix  n       -       n       300     1       qmgr
tlsmgr    unix  -       -       n       1000?   1       tlsmgr
rewrite   unix  -       -       n       -       -       trivial-rewrite
bounce    unix  -       -       n       -       0       bounce
defer     unix  -       -       n       -       0       bounce
trace     unix  -       -       n       -       0       bounce
verify    unix  -       -       n       -       1       verify
flush     unix  n       -       n       1000?   0       flush
proxymap  unix  -       -       n       -       -       proxymap
proxywrite unix -       -       n       -       1       proxymap
smtp      unix  -       -       n       -       -       smtp
relay     unix  -       -       n       -       -       smtp
  -o syslog_name=postfix/$service_name
showq     unix  n       -       n       -       -       showq
error     unix  -       -       n       -       -       error
retry     unix  -       -       n       -       -       error
discard   unix  -       -       n       -       -       discard
local     unix  -       n       n       -       -       local
virtual   unix  -       n       n       -       -       virtual
lmtp      unix  -       -       n       -       -       lmtp
anvil     unix  -       -       n       -       1       anvil
scache    unix  -       -       n       -       1       scache
postlog   unix-dgram n  -       n       -       1       postlogd
CONF
fi

# Copy rendered configs to Postfix config dir (needed for all supporting files)
cp /etc/oxmail/postfix/main.cf /etc/postfix/main.cf 2>/dev/null
cp /etc/oxmail/postfix/master.cf /etc/postfix/master.cf 2>/dev/null
postalias /etc/postfix/aliases 2>/dev/null

exec postfix start-fg
