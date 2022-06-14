GW_HORNERO=212.129.4.141

certs:
	/usr/bin/curl -k https://black.riseup.net/ca.crt > /tmp/ca.crt
	/usr/bin/curl -k https://api.black.riseup.net/3/cert > /tmp/cert.pem

test-hornero:
	GW=${GW_HORNERO} CIPHER=AES-128-CBC PORT=53 ./vpn-connect.sh
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=53 ./vpn-connect.sh
	GW=${GW_HORNERO} CIPHER=AES-128-CBC PORT=1194 ./vpn-connect.sh
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 ./vpn-connect.sh
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 TLS=DHE-RSA-AES128-SHA ./vpn-connect-tls.sh
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 TLS=TLS-ECDHE-ECDSA-WITH-AES-256-GCM-SHA384 ./vpn-connect-tls.sh


