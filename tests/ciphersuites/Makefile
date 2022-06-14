GW_HORNERO=212.129.4.141
OK="Peer Connection Initiated"
SLEEP_AND_KILL=sleep 3 && sudo pkill openvpn &

certs:
	/usr/bin/curl -k https://black.riseup.net/ca.crt > /tmp/ca.crt
	/usr/bin/curl -k https://api.black.riseup.net/3/cert > /tmp/cert.pem

test-hornero:
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-128-CBC PORT=53 PROTO=tcp4 ./vpn-connect.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=53 PROTO=tcp4 ./vpn-connect.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-128-CBC PORT=80 PROTO=udp4 ./vpn-connect.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=80 PROTO=udp4 ./vpn-connect.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-128-CBC PORT=1194 PROTO=tcp4 ./vpn-connect.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-128-CBC PORT=1194 PROTO=udp4 ./vpn-connect.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 PROTO=tcp4 ./vpn-connect.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 PROTO=udp4 ./vpn-connect.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 TLS=DHE-RSA-AES128-SHA PROTO=tcp4 ./vpn-connect-tls.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 TLS=DHE-RSA-AES128-SHA PROTO=udp4 ./vpn-connect-tls.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 TLS=TLS-ECDHE-ECDSA-WITH-AES-256-GCM-SHA384 PROTO=tcp4 ./vpn-connect-tls.sh | grep --color ${OK}
	${SLEEP_AND_KILL}
	GW=${GW_HORNERO} CIPHER=AES-256-GCM PORT=1194 TLS=TLS-ECDHE-ECDSA-WITH-AES-256-GCM-SHA384 PROTO=udp4 ./vpn-connect-tls.sh | grep --color ${OK}
