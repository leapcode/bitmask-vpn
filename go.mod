module 0xacab.org/leap/bitmask-vpn

go 1.17

require (
	0xacab.org/leap/obfsvpn v0.0.0-20220626143947-feff527c00e5
	git.torproject.org/pluggable-transports/goptlib.git v1.2.0
	git.torproject.org/pluggable-transports/snowflake.git v1.1.0
	github.com/ProtonMail/go-autostart v0.0.0-20210130080809-00ed301c8e9a
	github.com/apparentlymart/go-openvpn-mgmt v0.0.0-20200929191752-4d2ce95ae600
	github.com/cretz/bine v0.2.0
	github.com/dchest/siphash v1.2.3 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/keybase/go-ps v0.0.0-20190827175125-91aafc93ba19
	github.com/pion/webrtc/v3 v3.1.41
	github.com/sevlyar/go-daemon v0.1.5
	github.com/smartystreets/goconvey v1.6.4
	github.com/xtaci/kcp-go/v5 v5.6.1
	github.com/xtaci/smux v1.5.16
	// Do not update obfs4 past e330d1b7024b, a backwards incompatible change was
	// made that will break negotiation!! riseup should move to the newest asap.
	gitlab.com/yawning/obfs4.git v0.0.0-20210511220700-e330d1b7024b // indirect
	golang.org/x/sys v0.0.0-20220624220833-87e55d714810
)

require github.com/natefinch/npipe v0.0.0-20160621034901-c1b8fa8bdcce

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kalikaneko/socks5 v1.0.1 // indirect
	github.com/klauspost/cpuid/v2 v2.0.14 // indirect
	github.com/klauspost/reedsolomon v1.10.0 // indirect
	github.com/pion/datachannel v1.5.2 // indirect
	github.com/pion/dtls/v2 v2.1.5 // indirect
	github.com/pion/ice/v2 v2.2.6 // indirect
	github.com/pion/interceptor v0.1.11 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.5 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.9 // indirect
	github.com/pion/rtp v1.7.13 // indirect
	github.com/pion/sctp v1.8.2 // indirect
	github.com/pion/sdp/v3 v3.0.5 // indirect
	github.com/pion/srtp/v2 v2.0.9 // indirect
	github.com/pion/stun v0.3.5 // indirect
	github.com/pion/transport v0.13.1 // indirect
	github.com/pion/turn/v2 v2.0.8 // indirect
	github.com/pion/udp v0.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/smartystreets/assertions v0.0.0-20180927180507-b2de0cb4f26d // indirect
	github.com/templexxx/cpu v0.0.9 // indirect
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20191217153810-f85b25db303b // indirect
	github.com/templexxx/xorsimd v0.4.1 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/xtaci/kcp-go v5.4.20+incompatible // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)
